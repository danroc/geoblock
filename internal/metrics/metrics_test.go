package metrics

import (
	"strings"
	"sync"
	"testing"

	"github.com/danroc/geoblock/internal/version"
)

// setupTest resets metrics and registers cleanup to reset after test.
func setupTest(t *testing.T) {
	t.Helper()
	resetMetrics()
	t.Cleanup(resetMetrics)
}

func TestIncDenied(t *testing.T) {
	setupTest(t)

	IncDenied()

	snapshot := Get()
	if snapshot.Requests.Denied != 1 {
		t.Errorf(
			"Expected denied count to be 1, got %d",
			snapshot.Requests.Denied,
		)
	}
}

func TestIncAllowed(t *testing.T) {
	setupTest(t)

	IncAllowed()

	snapshot := Get()
	if snapshot.Requests.Allowed != 1 {
		t.Errorf(
			"Expected allowed count to be 1, got %d",
			snapshot.Requests.Allowed,
		)
	}
}

func TestIncInvalid(t *testing.T) {
	setupTest(t)

	IncInvalid()

	snapshot := Get()
	if snapshot.Requests.Invalid != 1 {
		t.Errorf(
			"Expected invalid count to be 1, got %d",
			snapshot.Requests.Invalid,
		)
	}
}

func TestMultipleIncrements(t *testing.T) {
	setupTest(t)

	// Test multiple increments
	for i := 0; i < 5; i++ {
		IncDenied()
	}
	for i := 0; i < 3; i++ {
		IncAllowed()
	}
	for i := 0; i < 2; i++ {
		IncInvalid()
	}

	snapshot := Get()

	if snapshot.Requests.Denied != 5 {
		t.Errorf(
			"Expected denied count to be 5, got %d",
			snapshot.Requests.Denied,
		)
	}
	if snapshot.Requests.Allowed != 3 {
		t.Errorf(
			"Expected allowed count to be 3, got %d",
			snapshot.Requests.Allowed,
		)
	}
	if snapshot.Requests.Invalid != 2 {
		t.Errorf(
			"Expected invalid count to be 2, got %d",
			snapshot.Requests.Invalid,
		)
	}
}

func TestConcurrentIncrements(t *testing.T) {
	setupTest(t)

	const numGoroutines = 100
	const incrementsPerGoroutine = 10

	var wg sync.WaitGroup

	// Test concurrent increments of denied requests
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				IncDenied()
			}
		}()
	}

	// Test concurrent increments of allowed requests
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				IncAllowed()
			}
		}()
	}

	// Test concurrent increments of invalid requests
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				IncInvalid()
			}
		}()
	}

	wg.Wait()

	snapshot := Get()
	expected := uint64(numGoroutines * incrementsPerGoroutine)

	if snapshot.Requests.Denied != expected {
		t.Errorf(
			"Expected denied count to be %d, got %d",
			expected,
			snapshot.Requests.Denied,
		)
	}
	if snapshot.Requests.Allowed != expected {
		t.Errorf(
			"Expected allowed count to be %d, got %d",
			expected,
			snapshot.Requests.Allowed,
		)
	}
	if snapshot.Requests.Invalid != expected {
		t.Errorf(
			"Expected invalid count to be %d, got %d",
			expected,
			snapshot.Requests.Invalid,
		)
	}
}

func TestGet(t *testing.T) {
	setupTest(t)

	snapshot := Get()

	// Check that snapshot contains version
	if snapshot.Version == "" {
		t.Error("Expected version to be non-empty")
	}

	// Check that snapshot.Version matches what version.Get() returns
	expectedVersion := version.Get()
	if snapshot.Version != expectedVersion {
		t.Errorf(
			"Expected version to be %q, got %q",
			expectedVersion,
			snapshot.Version,
		)
	}

	// Check initial values are zero
	if snapshot.Requests.Denied != 0 {
		t.Errorf(
			"Expected initial denied count to be 0, got %d",
			snapshot.Requests.Denied,
		)
	}
	if snapshot.Requests.Allowed != 0 {
		t.Errorf(
			"Expected initial allowed count to be 0, got %d",
			snapshot.Requests.Allowed,
		)
	}
	if snapshot.Requests.Invalid != 0 {
		t.Errorf(
			"Expected initial invalid count to be 0, got %d",
			snapshot.Requests.Invalid,
		)
	}
}

func TestTotalCalculation(t *testing.T) {
	testCases := []struct {
		name    string
		denied  int
		allowed int
		invalid int
	}{
		{"zero values", 0, 0, 0},
		{"only denied", 5, 0, 0},
		{"only allowed", 0, 3, 0},
		{"only invalid", 0, 0, 2},
		{"mixed values", 10, 5, 3},
		{"large values", 1000, 2000, 500},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTest(t)

			// Add metrics according to test case
			for i := 0; i < tc.denied; i++ {
				IncDenied()
			}
			for i := 0; i < tc.allowed; i++ {
				IncAllowed()
			}
			for i := 0; i < tc.invalid; i++ {
				IncInvalid()
			}

			snapshot := Get()

			if snapshot.Requests.Denied != uint64(tc.denied) {
				t.Errorf(
					"Expected denied to be %d, got %d",
					tc.denied,
					snapshot.Requests.Denied,
				)
			}
			if snapshot.Requests.Allowed != uint64(tc.allowed) {
				t.Errorf(
					"Expected allowed to be %d, got %d",
					tc.allowed,
					snapshot.Requests.Allowed,
				)
			}
			if snapshot.Requests.Invalid != uint64(tc.invalid) {
				t.Errorf(
					"Expected invalid to be %d, got %d",
					tc.invalid,
					snapshot.Requests.Invalid,
				)
			}
		})
	}
}

// BenchmarkIncDenied benchmarks the IncDenied function.
func BenchmarkIncDenied(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IncDenied()
	}
}

// BenchmarkIncAllowed benchmarks the IncAllowed function.
func BenchmarkIncAllowed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IncAllowed()
	}
}

// BenchmarkIncInvalid benchmarks the IncInvalid function.
func BenchmarkIncInvalid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IncInvalid()
	}
}

// BenchmarkGet benchmarks the Get function.
func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Get()
	}
}

// BenchmarkConcurrentIncrements benchmarks concurrent increments.
func BenchmarkConcurrentIncrements(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			IncDenied()
			IncAllowed()
			IncInvalid()
		}
	})
}

// resetMetrics resets the global metrics for testing.
func resetMetrics() {
	metrics.Denied.Store(0)
	metrics.Allowed.Store(0)
	metrics.Invalid.Store(0)
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
