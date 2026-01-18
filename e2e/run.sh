#!/bin/bash

set -euo pipefail

# Constants
BASE_URL="http://localhost:8080/v1/forward-auth"
HEALTH_URL="http://localhost:8080/v1/health"
METRICS_URL="http://localhost:8080/metrics"
MAX_RETRIES=30

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Check for required commands
for cmd in jq curl; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo ":: $cmd is required but not installed."
        exit 1
    fi
done

CGO_ENABLED=0 make build

export GEOBLOCK_CONFIG_FILE=/app/e2e/config.yaml
export GEOBLOCK_PORT=8080
export GEOBLOCK_LOG_LEVEL=debug
export GEOBLOCK_LOG_FORMAT=json

# Graceful shutdown
cleanup() {
    if [ -n "${SERVER_PID:-}" ]; then
        kill "$SERVER_PID" 2>/dev/null || true
    fi
}
trap cleanup EXIT

./dist/geoblock &> geoblock.log &
SERVER_PID=$!

# Health check with timeout
RETRY_COUNT=0
while ! curl -fs "$HEALTH_URL"; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ "$RETRY_COUNT" -ge "$MAX_RETRIES" ]; then
        echo ":: Server failed to start within ${MAX_RETRIES}s"
        exit 1
    fi
    sleep 1
done

function run_single_test() {
    local test_name=$1
    shift

    local expected_status=$1
    shift

    echo -n ":: Testing: $test_name ... "

    local status
    status=$(curl -s -o /dev/null -w "%{http_code}" "$@")

    if [ "$status" -ne "$expected_status" ]; then
        echo "FAILED"
        echo "   Expected: $expected_status"
        echo "   Actual:   $status"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    else
        echo "PASSED"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    fi
}

# Track if any test failed
ANY_FAILED=0

run_test() {
    run_single_test "$@" || ANY_FAILED=1
}

echo ""
echo "=== Running E2E Tests ==="
echo ""

# Invalid request tests
run_test 'missing "X-Forwarded-Host" header' 400 \
    "$BASE_URL" \
    -H "X-Forwarded-For: 127.0.0.1" \
    -H "X-Forwarded-Method: GET"

run_test 'invalid source IP address' 400 \
    "$BASE_URL" \
    -H "X-Forwarded-For: invalid-ip" \
    -H "X-Forwarded-Host: example.org" \
    -H "X-Forwarded-Method: GET"

run_test 'missing "X-Forwarded-For" header' 400 \
    "$BASE_URL" \
    -H "X-Forwarded-Host: example.com" \
    -H "X-Forwarded-Method: GET"

run_test 'missing "X-Forwarded-Method" header' 400 \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: example.com"

# Domain + country tests
run_test 'blocked by domain+country' 403 \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: example.org" \
    -H "X-Forwarded-Method: GET"

run_test 'allowed by domain+country' 204 \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: example.com" \
    -H "X-Forwarded-Method: GET"

# Local network tests
run_test 'allowed local network (IPv4)' 204 \
    "$BASE_URL" \
    -H "X-Forwarded-For: 127.0.0.1" \
    -H "X-Forwarded-Host: example.com" \
    -H "X-Forwarded-Method: GET"

run_test 'allowed local network (IPv6)' 204 \
    "$BASE_URL" \
    -H "X-Forwarded-For: ::1" \
    -H "X-Forwarded-Host: example.com" \
    -H "X-Forwarded-Method: GET"

# ASN blocking test (8.8.8.8 = AS15169 = Google)
run_test 'blocked by ASN' 403 \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: asn-blocked.test" \
    -H "X-Forwarded-Method: GET"

# HTTP method tests
run_test 'blocked by method (POST)' 403 \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: method-test.test" \
    -H "X-Forwarded-Method: POST"

run_test 'allowed by method (GET)' 204 \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: method-test.test" \
    -H "X-Forwarded-Method: GET"

# Wildcard domain test
run_test 'allowed by wildcard domain' 204 \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: api.wildcard.test" \
    -H "X-Forwarded-Method: GET"

# Default policy fallback test
run_test 'denied by default policy' 403 \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: unknown-domain.test" \
    -H "X-Forwarded-Method: GET"

echo ""
echo "=== Test Summary ==="
TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED))
echo ":: Passed: $TESTS_PASSED/$TOTAL_TESTS"
echo ":: Failed: $TESTS_FAILED/$TOTAL_TESTS"
echo ""

if [ "$ANY_FAILED" -ne 0 ]; then
    echo ":: SOME TESTS FAILED"
    exit 1
fi

echo "=== Verifying Metrics ==="
curl -s "$METRICS_URL" > metrics.prometheus

diff <(sed 's/{version="[^"]*"}//' metrics.prometheus) \
     <(sed 's/{version="[^"]*"}//' e2e/metrics-expected.prometheus)
echo ":: Metrics match expected values"

echo ""
echo "=== Verifying Logs ==="
diff <(jq --sort-keys 'del(.time, .version)' e2e/expected.log) \
     <(jq --sort-keys 'del(.time, .version)' geoblock.log)
echo ":: Logs match expected values"

echo ""
echo ":: ALL E2E TESTS PASSED"
