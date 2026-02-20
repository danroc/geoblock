#!/bin/bash

# Shared test utilities for e2e and integration tests.

# Test counters — sourcing scripts should NOT redeclare these.
TESTS_PASSED=0
TESTS_FAILED=0

# require_commands - Checks that all listed commands are available.
#
# Arguments:
#   CMD...  One or more command names to check.
require_commands() {
    for cmd in "$@"; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            echo ":: Error: $cmd is required but not installed"
            exit 1
        fi
    done
}

# assert_status - Runs a curl request and checks the HTTP status code.
#
# Always returns 0 (safe under set -e); use print_summary for exit code.
#
# Arguments:
#   EXPECTED_STATUS  Expected HTTP status code.
#   TEST_NAME        Label printed in the test output.
#   CURL_ARGS...     Additional arguments passed to curl.
#
# Examples:
#   assert_status 200 "allowed domain" -H "Host: example.test" "$URL"
assert_status() {
    local expected_status=$1
    local test_name=$2
    shift 2

    echo -n ":: Testing: ${test_name} ... "

    local status
    status=$(curl -s -o /dev/null -w "%{http_code}" "$@")

    if [ "$status" -ne "$expected_status" ]; then
        echo "FAILED"
        echo "   Expected: ${expected_status}"
        echo "   Actual:   ${status}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    else
        echo "PASSED"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi
}

# print_summary - Prints test results and returns non-zero if any test failed.
#
# Arguments:
#   LABEL  Section name shown in the summary header.
#
# Examples:
#   print_summary "E2E" || exit 1
print_summary() {
    local label=$1
    local total=$((TESTS_PASSED + TESTS_FAILED))

    echo ""
    echo "=== ${label} Summary ==="
    echo ":: Passed: ${TESTS_PASSED}/${total}"
    echo ":: Failed: ${TESTS_FAILED}/${total}"

    if [ "$TESTS_FAILED" -ne 0 ]; then
        echo ""
        echo ":: Some ${label} tests FAILED"
        return 1
    fi

    echo ""
    echo ":: All ${label} tests PASSED"
    return 0
}
