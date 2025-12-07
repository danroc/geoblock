package metrics

import (
	"strings"
	"testing"
)

// resetMetrics resets the global metrics for testing.
func resetMetrics() {
	requestCount.Denied.Store(0)
	requestCount.Allowed.Store(0)
	requestCount.Invalid.Store(0)
}

// setupTest resets metrics and registers cleanup to reset after test.
func setupTest(t *testing.T) {
	t.Helper()
	resetMetrics()
	t.Cleanup(resetMetrics)
}

func TestRequestCount_Empty(t *testing.T) {
	setupTest(t)

	var (
		allowed = requestCount.Allowed.Load()
		denied  = requestCount.Denied.Load()
		invalid = requestCount.Invalid.Load()
	)

	if denied != 0 {
		t.Errorf("Expected initial denied count to be 0, got %d", denied)
	}
	if allowed != 0 {
		t.Errorf("Expected initial allowed count to be 0, got %d", allowed)
	}
	if invalid != 0 {
		t.Errorf("Expected initial invalid count to be 0, got %d", invalid)
	}
}

func TestRequestCount_Many(t *testing.T) {
	setupTest(t)

	for i := 0; i < 5; i++ {
		IncDenied()
	}
	for i := 0; i < 3; i++ {
		IncAllowed()
	}
	for i := 0; i < 2; i++ {
		IncInvalid()
	}

	var (
		allowed = requestCount.Allowed.Load()
		denied  = requestCount.Denied.Load()
		invalid = requestCount.Invalid.Load()
	)

	if denied != 5 {
		t.Errorf("Expected denied count to be 5, got %d", denied)
	}
	if allowed != 3 {
		t.Errorf("Expected allowed count to be 3, got %d", allowed)
	}
	if invalid != 2 {
		t.Errorf("Expected invalid count to be 2, got %d", invalid)
	}
}

func TestPrometheus(t *testing.T) {
	setupTest(t)

	// Add some test data
	IncAllowed()
	IncAllowed()
	IncDenied()
	IncInvalid()

	output := Prometheus()

	// Check that output contains expected Prometheus format elements
	expectedStrings := []string{
		"# HELP geoblock_version_info Version information",
		"# TYPE geoblock_version_info gauge",
		"geoblock_version_info{version=",
		"# HELP geoblock_requests_total Total number of requests by status",
		"# TYPE geoblock_requests_total counter",
		"geoblock_requests_total{status=\"allowed\"} 2",
		"geoblock_requests_total{status=\"denied\"} 1",
		"geoblock_requests_total{status=\"invalid\"} 1",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf(
				"Expected output to contain %q, but it didn't.\nFull output:\n%s",
				expected,
				output,
			)
		}
	}
}
