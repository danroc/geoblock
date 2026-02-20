#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# shellcheck source=../lib.sh
source "${SCRIPT_DIR}/../lib.sh"

# Constants
BASE_URL="http://localhost:8080/v1/forward-auth"
HEALTH_URL="http://localhost:8080/v1/health"
METRICS_URL="http://localhost:8080/metrics"
MAX_RETRIES=30

require_commands jq curl

TEMP_DIR="$(mktemp -d)"

CGO_ENABLED=0 make -C "${ROOT_DIR}" build

LOG_FILE="${TEMP_DIR}/geoblock.log"
METRICS_FILE="${TEMP_DIR}/metrics.prometheus"

mkdir -p "${ROOT_DIR}/.cache"
export GEOBLOCK_CACHE_DIR="${ROOT_DIR}/.cache"
export GEOBLOCK_CONFIG_FILE="${SCRIPT_DIR}/config.yaml"
export GEOBLOCK_PORT=8080
export GEOBLOCK_LOG_LEVEL=debug
export GEOBLOCK_LOG_FORMAT=json

cleanup() {
    if [ -n "${SERVER_PID:-}" ]; then
        kill "$SERVER_PID" 2>/dev/null || true
    fi
    rm -rf "${TEMP_DIR}"
}
trap cleanup EXIT

"${ROOT_DIR}/dist/geoblock" &> "${LOG_FILE}" &
SERVER_PID=$!

# Health check with timeout
RETRY_COUNT=0
while ! curl -fs "$HEALTH_URL"; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ "$RETRY_COUNT" -ge "$MAX_RETRIES" ]; then
        echo ":: Error: Server failed to start within ${MAX_RETRIES}s"
        exit 1
    fi
    sleep 1
done

echo ""
echo "=== Running E2E Tests ==="

# Invalid request tests
assert_status 400 'missing "X-Forwarded-Host" header' \
    "$BASE_URL" \
    -H "X-Forwarded-For: 127.0.0.1" \
    -H "X-Forwarded-Method: GET"

assert_status 400 'invalid source IP address' \
    "$BASE_URL" \
    -H "X-Forwarded-For: invalid-ip" \
    -H "X-Forwarded-Host: country-blocked.test" \
    -H "X-Forwarded-Method: GET"

assert_status 400 'missing "X-Forwarded-For" header' \
    "$BASE_URL" \
    -H "X-Forwarded-Host: country-allowed.test" \
    -H "X-Forwarded-Method: GET"

assert_status 400 'missing "X-Forwarded-Method" header' \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: country-allowed.test"

# Domain + country tests
assert_status 403 'blocked by domain+country' \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: country-blocked.test" \
    -H "X-Forwarded-Method: GET"

assert_status 204 'allowed by domain+country' \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: country-allowed.test" \
    -H "X-Forwarded-Method: GET"

# Local network tests
assert_status 204 'allowed local network (IPv4)' \
    "$BASE_URL" \
    -H "X-Forwarded-For: 127.0.0.1" \
    -H "X-Forwarded-Host: local.test" \
    -H "X-Forwarded-Method: GET"

assert_status 204 'allowed local network (IPv6)' \
    "$BASE_URL" \
    -H "X-Forwarded-For: ::1" \
    -H "X-Forwarded-Host: local.test" \
    -H "X-Forwarded-Method: GET"

# ASN blocking test (8.8.8.8 = AS15169 = Google)
assert_status 403 'blocked by ASN' \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: asn-blocked.test" \
    -H "X-Forwarded-Method: GET"

# HTTP method tests
assert_status 403 'blocked by method (POST)' \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: method-test.test" \
    -H "X-Forwarded-Method: POST"

assert_status 204 'allowed by method (GET)' \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: method-test.test" \
    -H "X-Forwarded-Method: GET"

# Wildcard domain test
assert_status 204 'allowed by wildcard domain' \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: api.wildcard.test" \
    -H "X-Forwarded-Method: GET"

# Default policy fallback test
assert_status 403 'denied by default policy' \
    "$BASE_URL" \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: unknown-domain.test" \
    -H "X-Forwarded-Method: GET"

print_summary "E2E" || exit 1

echo ""
echo "=== Verifying Metrics ==="
curl -s "$METRICS_URL" > "${METRICS_FILE}"

# Filter function to remove dynamic values for comparison
filter_metrics() {
    sed -e '/^geoblock_version_info/d' \
        -e '/^geoblock_start_time_seconds/d' \
        -e '/^geoblock_config_last_reload_timestamp/d' \
        -e '/^geoblock_db_last_update_timestamp/d' \
        -e '/^geoblock_db_load_duration_seconds/d' \
        -e '/^geoblock_db_entries/d' \
        -e '/^geoblock_request_duration_seconds/d'
}

echo -n ":: Comparing metrics ... "
diff <(filter_metrics < "${METRICS_FILE}") \
     <(filter_metrics < "${SCRIPT_DIR}/metrics-expected.prometheus")
echo "PASSED"

echo ""
echo "=== Verifying Logs ==="
echo -n ":: Comparing logs ... "
diff <(jq --sort-keys 'del(.time, .version, .commit)' "${SCRIPT_DIR}/expected.log") \
     <(jq --sort-keys 'del(.time, .version, .commit)' "${LOG_FILE}")
echo "PASSED"

echo ""
echo "=== Testing Graceful Shutdown ==="
# Send SIGTERM and wait for process to exit
kill -TERM "$SERVER_PID"
wait "$SERVER_PID"
EXIT_CODE=$?

# Verify clean exit (exit code 0)
echo -n ":: Checking exit code ... "
if [ "$EXIT_CODE" -ne 0 ]; then
    echo "FAILED"
    echo "   Expected: 0"
    echo "   Actual:   $EXIT_CODE"
    exit 1
fi
echo "PASSED"

# Verify shutdown message in logs
echo -n ":: Checking shutdown message ... "
if ! grep -q "Shutting down server" "${LOG_FILE}"; then
    echo "FAILED"
    echo "   Expected: Shutdown message in logs"
    echo "   Actual:   Message not found"
    exit 1
fi
echo "PASSED"

# Clear SERVER_PID so cleanup doesn't try to kill it again
SERVER_PID=""

echo ""
echo ":: All E2E tests PASSED"
