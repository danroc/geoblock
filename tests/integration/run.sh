#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# shellcheck source=../lib.sh
source "${SCRIPT_DIR}/../lib.sh"

require_commands curl docker

if ! docker image inspect geoblock >/dev/null 2>&1; then
    echo ":: Error: 'geoblock' Docker image not found. Run 'make docker' first."
    exit 1
fi

mkdir -p "${ROOT_DIR}/.cache"

PORT="${GEOBLOCK_TEST_PORT:-8080}"
BASE_URL="http://localhost:${PORT}"

PROXIES=("traefik" "nginx" "caddy")
MAX_RETRIES=60

compose() {
    local proxy=$1
    shift
    docker compose \
        -f "${SCRIPT_DIR}/compose.base.yaml" \
        -f "${SCRIPT_DIR}/compose.${proxy}.yaml" \
        -p "geoblock-integration-${proxy}" "$@"
}

teardown() {
    compose "$1" down --volumes --timeout 10 2>/dev/null || true
}

cleanup_all() {
    for proxy in "${PROXIES[@]}"; do
        teardown "${proxy}"
    done
}
trap cleanup_all EXIT

wait_for_healthy() {
    local retry_count=0
    echo -n ":: Waiting for stack to be healthy "
    while ! curl -sf -o /dev/null -H "Host: whoami-1.test" "${BASE_URL}"; do
        retry_count=$((retry_count + 1))
        if [ "$retry_count" -ge "$MAX_RETRIES" ]; then
            echo " TIMEOUT"
            return 1
        fi
        echo -n "."
        sleep 1
    done
    echo " OK"
}

run_tests_for_proxy() {
    local proxy=$1
    local failed_before=$TESTS_FAILED

    echo ""
    echo "========================================"
    echo "  Testing with ${proxy}"
    echo "========================================"

    teardown "${proxy}"

    echo ":: Starting ${proxy} stack..."
    compose "${proxy}" up -d

    if ! wait_for_healthy; then
        echo ":: Dumping logs for debugging..."
        compose "${proxy}" logs
        teardown "${proxy}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return
    fi

    assert_status 200 "allowed domain via ${proxy}" \
        -H "Host: whoami-1.test" "${BASE_URL}"

    assert_status 403 "denied domain via ${proxy}" \
        -H "Host: whoami-2.test" "${BASE_URL}"

    assert_status 403 "denied method via ${proxy}" \
        -X POST -H "Host: method-test.test" "${BASE_URL}"

    assert_status 200 "allowed method via ${proxy}" \
        -H "Host: method-test.test" "${BASE_URL}"

    assert_status 200 "allowed wildcard domain via ${proxy}" \
        -H "Host: sub.wildcard.test" "${BASE_URL}"

    assert_status 403 "denied by default policy via ${proxy}" \
        -H "Host: unknown.test" "${BASE_URL}"

    if [ "$TESTS_FAILED" -gt "$failed_before" ]; then
        echo ":: Dumping logs for debugging..."
        compose "${proxy}" logs
    fi

    teardown "${proxy}"
}

echo ""
echo "=== Running Integration Tests ==="

for proxy in "${PROXIES[@]}"; do
    run_tests_for_proxy "${proxy}"
done

print_summary "Integration" || exit 1
