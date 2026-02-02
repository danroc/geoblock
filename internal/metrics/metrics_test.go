package metrics

import (
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/danroc/geoblock/internal/version"
)

// setupTest resets metrics and registers cleanup to reset after test.
func setupTest(t *testing.T) {
	t.Helper()
	Reset()
	t.Cleanup(Reset)
}

// getCounterValue returns the current value of a counter with the given label.
func getCounterValue(status string) float64 {
	return testutil.ToFloat64(requestsTotal.WithLabelValues(status))
}

func TestIncrementFunctions(t *testing.T) {
	tests := []struct {
		name   string
		inc    func()
		status string
	}{
		{"IncDenied", IncDenied, "denied"},
		{"IncAllowed", IncAllowed, "allowed"},
		{"IncInvalid", IncInvalid, "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest(t)
			tt.inc()
			if got := getCounterValue(tt.status); got != 1 {
				t.Errorf("Expected %s count to be 1, got %v", tt.status, got)
			}
		})
	}
}

func TestMultipleIncrements(t *testing.T) {
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

	if got := getCounterValue("denied"); got != 5 {
		t.Errorf("Expected denied count to be 5, got %v", got)
	}
	if got := getCounterValue("allowed"); got != 3 {
		t.Errorf("Expected allowed count to be 3, got %v", got)
	}
	if got := getCounterValue("invalid"); got != 2 {
		t.Errorf("Expected invalid count to be 2, got %v", got)
	}
}

func TestConcurrentIncrements(t *testing.T) {
	setupTest(t)

	const numGoroutines = 100
	const incrementsPerGoroutine = 10

	increments := []struct {
		inc    func()
		status string
	}{
		{IncDenied, "denied"},
		{IncAllowed, "allowed"},
		{IncInvalid, "invalid"},
	}

	var wg sync.WaitGroup
	for _, inc := range increments {
		wg.Add(numGoroutines)
		for range numGoroutines {
			go func() {
				defer wg.Done()
				for range incrementsPerGoroutine {
					inc.inc()
				}
			}()
		}
	}
	wg.Wait()

	expected := float64(numGoroutines * incrementsPerGoroutine)
	for _, inc := range increments {
		if got := getCounterValue(inc.status); got != expected {
			t.Errorf("Expected %s count to be %v, got %v", inc.status, expected, got)
		}
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

			for i := 0; i < tc.denied; i++ {
				IncDenied()
			}
			for i := 0; i < tc.allowed; i++ {
				IncAllowed()
			}
			for i := 0; i < tc.invalid; i++ {
				IncInvalid()
			}

			if got := getCounterValue("denied"); got != float64(tc.denied) {
				t.Errorf("Expected denied to be %d, got %v", tc.denied, got)
			}
			if got := getCounterValue("allowed"); got != float64(tc.allowed) {
				t.Errorf("Expected allowed to be %d, got %v", tc.allowed, got)
			}
			if got := getCounterValue("invalid"); got != float64(tc.invalid) {
				t.Errorf("Expected invalid to be %d, got %v", tc.invalid, got)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	setupTest(t)

	IncAllowed()
	IncAllowed()
	IncDenied()
	IncInvalid()

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()

	expectedStrings := []string{
		"# HELP geoblock_version_info Version information",
		"# TYPE geoblock_version_info gauge",
		"geoblock_version_info{version=\"" + version.Get() + "\"} 1",
		"# HELP geoblock_requests_total Total number of requests by status",
		"# TYPE geoblock_requests_total counter",
		"geoblock_requests_total{status=\"allowed\"} 2",
		"geoblock_requests_total{status=\"denied\"} 1",
		"geoblock_requests_total{status=\"invalid\"} 1",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(body, expected) {
			t.Errorf(
				"Expected output to contain %q, but it didn't.\nFull output:\n%s",
				expected,
				body,
			)
		}
	}
}

func TestReset(t *testing.T) {
	IncAllowed()
	IncDenied()
	IncInvalid()

	Reset()

	if got := getCounterValue("allowed"); got != 0 {
		t.Errorf("Expected allowed to be 0 after reset, got %v", got)
	}
	if got := getCounterValue("denied"); got != 0 {
		t.Errorf("Expected denied to be 0 after reset, got %v", got)
	}
	if got := getCounterValue("invalid"); got != 0 {
		t.Errorf("Expected invalid to be 0 after reset, got %v", got)
	}

	// Verify version info is restored after reset
	if got := testutil.ToFloat64(versionInfo.WithLabelValues(version.Get())); got != 1 {
		t.Errorf("Expected version info to be 1 after reset, got %v", got)
	}
}

func BenchmarkIncrements(b *testing.B) {
	benchmarks := []struct {
		name string
		inc  func()
	}{
		{"Denied", IncDenied},
		{"Allowed", IncAllowed},
		{"Invalid", IncInvalid},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for range b.N {
				bm.inc()
			}
		})
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
