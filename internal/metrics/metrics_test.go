package metrics

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/danroc/geoblock/internal/version"
)

func TestIncDenied(t *testing.T) {
	// Reset metrics before test
	resetMetrics()

	IncDenied()

	snapshot := Get()
	if snapshot.Requests.Denied != 1 {
		t.Errorf(
			"Expected denied count to be 1, got %d",
			snapshot.Requests.Denied,
		)
	}
	if snapshot.Requests.Total != 1 {
		t.Errorf(
			"Expected total count to be 1, got %d",
			snapshot.Requests.Total,
		)
	}
}

func TestIncAllowed(t *testing.T) {
	// Reset metrics before test
	resetMetrics()

	IncAllowed()

	snapshot := Get()
	if snapshot.Requests.Allowed != 1 {
		t.Errorf(
			"Expected allowed count to be 1, got %d",
			snapshot.Requests.Allowed,
		)
	}
	if snapshot.Requests.Total != 1 {
		t.Errorf(
			"Expected total count to be 1, got %d",
			snapshot.Requests.Total,
		)
	}
}

func TestIncInvalid(t *testing.T) {
	// Reset metrics before test
	resetMetrics()

	IncInvalid()

	snapshot := Get()
	if snapshot.Requests.Invalid != 1 {
		t.Errorf(
			"Expected invalid count to be 1, got %d",
			snapshot.Requests.Invalid,
		)
	}
	if snapshot.Requests.Total != 1 {
		t.Errorf(
			"Expected total count to be 1, got %d",
			snapshot.Requests.Total,
		)
	}
}

func TestMultipleIncrements(t *testing.T) {
	// Reset metrics before test
	resetMetrics()

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
	if snapshot.Requests.Total != 10 {
		t.Errorf(
			"Expected total count to be 10, got %d",
			snapshot.Requests.Total,
		)
	}
}

func TestConcurrentIncrements(t *testing.T) {
	// Reset metrics before test
	resetMetrics()

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
	if snapshot.Requests.Total != expected*3 {
		t.Errorf(
			"Expected total count to be %d, got %d",
			expected*3,
			snapshot.Requests.Total,
		)
	}
}

func TestGet(t *testing.T) {
	// Reset metrics before test
	resetMetrics()

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
	if snapshot.Requests.Total != 0 {
		t.Errorf(
			"Expected initial total count to be 0, got %d",
			snapshot.Requests.Total,
		)
	}
}

func TestSnapshotJSON(t *testing.T) {
	// Reset metrics before test
	resetMetrics()

	// Add some metrics
	IncDenied()
	IncAllowed()
	IncInvalid()

	snapshot := Get()

	// Test JSON marshalling
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("Failed to marshal snapshot to JSON: %v", err)
	}

	// Test JSON unmarshalling
	var unmarshalled Snapshot
	err = json.Unmarshal(data, &unmarshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal snapshot from JSON: %v", err)
	}

	// Compare original and unmarshalled
	if unmarshalled.Version != snapshot.Version {
		t.Errorf(
			"Expected version %q, got %q",
			snapshot.Version,
			unmarshalled.Version,
		)
	}
	if unmarshalled.Requests.Denied != snapshot.Requests.Denied {
		t.Errorf(
			"Expected denied %d, got %d",
			snapshot.Requests.Denied,
			unmarshalled.Requests.Denied,
		)
	}
	if unmarshalled.Requests.Allowed != snapshot.Requests.Allowed {
		t.Errorf(
			"Expected allowed %d, got %d",
			snapshot.Requests.Allowed,
			unmarshalled.Requests.Allowed,
		)
	}
	if unmarshalled.Requests.Invalid != snapshot.Requests.Invalid {
		t.Errorf(
			"Expected invalid %d, got %d",
			snapshot.Requests.Invalid,
			unmarshalled.Requests.Invalid,
		)
	}
	if unmarshalled.Requests.Total != snapshot.Requests.Total {
		t.Errorf(
			"Expected total %d, got %d",
			snapshot.Requests.Total,
			unmarshalled.Requests.Total,
		)
	}
}

func TestRequestCountSnapshotJSON(t *testing.T) {
	rcs := RequestCountSnapshot{
		Allowed: 10,
		Denied:  5,
		Invalid: 2,
		Total:   17,
	}

	// Test JSON marshalling
	data, err := json.Marshal(rcs)
	if err != nil {
		t.Fatalf("Failed to marshal RequestCountSnapshot to JSON: %v", err)
	}

	// Test JSON unmarshalling
	var unmarshalled RequestCountSnapshot
	err = json.Unmarshal(data, &unmarshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal RequestCountSnapshot from JSON: %v", err)
	}

	// Compare original and unmarshalled
	if unmarshalled != rcs {
		t.Errorf("Expected %+v, got %+v", rcs, unmarshalled)
	}
}

func TestTotalCalculation(t *testing.T) {
	// Reset metrics before test
	resetMetrics()

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
			// Reset metrics before each subtest
			resetMetrics()

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
			expectedTotal := uint64(tc.denied + tc.allowed + tc.invalid)

			if snapshot.Requests.Total != expectedTotal {
				t.Errorf(
					"Expected total to be %d, got %d",
					expectedTotal,
					snapshot.Requests.Total,
				)
			}
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
