#!/bin/bash

set -euo pipefail

# Check for required commands
for cmd in jq curl; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo ":: $cmd is required but not installed."
        exit 1
    fi
done

CGO_ENABLED=0 make build

export GEOBLOCK_CONFIG=/app/examples/config.yaml
export GEOBLOCK_PORT=8080
export GEOBLOCK_LOG_LEVEL=debug
export GEOBLOCK_LOG_FORMAT=json

./dist/geoblock &> geoblock.log &

while ! curl -fs http://localhost:8080/v1/health; do
  sleep 1
done

function test() {
    local test_name=$1
    shift

    local expected_status=$1
    shift

    local status
    status=$(curl -s -o /dev/null -w "%{http_code}" "$@")

    if [ "$status" -ne "$expected_status" ]; then
        echo ":: Test \"$test_name\" failed. Expected status $expected_status, got $status"
        exit 1
    fi
}

test 'missing "X-Forwarded-Host" header' 400 \
    http://localhost:8080/v1/forward-auth \
    -H "X-Forwarded-For: 127.0.0.1" \
    -H "X-Forwarded-Method: GET"

test 'invalid source IP address' 400 \
    http://localhost:8080/v1/forward-auth \
    -H "X-Forwarded-For: invalid-ip" \
    -H "X-Forwarded-Host: example.org" \
    -H "X-Forwarded-Method: GET"

test 'missing "X-Forwarded-For" header' 400 \
    http://localhost:8080/v1/forward-auth \
    -H "X-Forwarded-Host: example.com" \
    -H "X-Forwarded-Method: GET"

test 'missing "X-Forwarded-Method" header' 400 \
    http://localhost:8080/v1/forward-auth \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: example.com"

test 'blocked by domain+country' 403 \
    http://localhost:8080/v1/forward-auth \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: example.org" \
    -H "X-Forwarded-Method: GET"

test 'allowed by domain+country' 204 \
    http://localhost:8080/v1/forward-auth \
    -H "X-Forwarded-For: 8.8.8.8" \
    -H "X-Forwarded-Host: example.com" \
    -H "X-Forwarded-Method: GET"

test 'allowed local network' 204 \
    http://localhost:8080/v1/forward-auth \
    -H "X-Forwarded-For: 127.0.0.1" \
    -H "X-Forwarded-Host: example.com" \
    -H "X-Forwarded-Method: GET"

curl http://localhost:8080/metrics > metrics.prometheus

diff <(sed 's/{version="[^"]*"}//' metrics.prometheus) \
     <(sed 's/{version="[^"]*"}//' e2e/metrics-expected.prometheus)

diff <(jq --sort-keys 'del(.time, .version)' e2e/expected.log) \
     <(jq --sort-keys 'del(.time, .version)' geoblock.log)

echo ":: ALL E2E TESTS PASSED"
