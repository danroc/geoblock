package metrics

import (
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/danroc/geoblock/internal/config"
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

// recordTestRequest is a helper that calls RecordRequest with common test values.
func recordTestRequest(status, action string) {
	RecordRequest(status, "US", "GET", time.Millisecond, 0, action, false)
}

func TestRecordRequest(t *testing.T) {
	tests := []struct {
		name   string
		status string
		action string
	}{
		{"allowed", StatusAllowed, config.PolicyAllow},
		{"denied", StatusDenied, config.PolicyDeny},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest(t)
			recordTestRequest(tt.status, tt.action)
			if got := getCounterValue(tt.status); got != 1 {
				t.Errorf("Expected %s count to be 1, got %v", tt.status, got)
			}
		})
	}
}

func TestRecordInvalidRequest(t *testing.T) {
	setupTest(t)
	RecordInvalidRequest(time.Millisecond)
	if got := getCounterValue(StatusInvalid); got != 1 {
		t.Errorf("Expected invalid count to be 1, got %v", got)
	}
}

func TestMultipleRequests(t *testing.T) {
	setupTest(t)

	for i := 0; i < 5; i++ {
		recordTestRequest(StatusDenied, config.PolicyDeny)
	}
	for i := 0; i < 3; i++ {
		recordTestRequest(StatusAllowed, config.PolicyAllow)
	}
	for i := 0; i < 2; i++ {
		RecordInvalidRequest(time.Millisecond)
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

func TestConcurrentRequests(t *testing.T) {
	setupTest(t)

	const numGoroutines = 100
	const requestsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3)

	// Concurrent denied requests
	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range requestsPerGoroutine {
				recordTestRequest(StatusDenied, config.PolicyDeny)
			}
		}()
	}

	// Concurrent allowed requests
	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range requestsPerGoroutine {
				recordTestRequest(StatusAllowed, config.PolicyAllow)
			}
		}()
	}

	// Concurrent invalid requests
	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range requestsPerGoroutine {
				RecordInvalidRequest(time.Millisecond)
			}
		}()
	}

	wg.Wait()

	expected := float64(numGoroutines * requestsPerGoroutine)
	for _, status := range []string{StatusDenied, StatusAllowed, StatusInvalid} {
		if got := getCounterValue(status); got != expected {
			t.Errorf("Expected %s count to be %v, got %v", status, expected, got)
		}
	}
}

func TestHandler(t *testing.T) {
	setupTest(t)

	recordTestRequest(StatusAllowed, config.PolicyAllow)
	recordTestRequest(StatusAllowed, config.PolicyAllow)
	recordTestRequest(StatusDenied, config.PolicyDeny)
	RecordInvalidRequest(time.Millisecond)

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
	recordTestRequest(StatusAllowed, config.PolicyAllow)
	recordTestRequest(StatusDenied, config.PolicyDeny)
	RecordInvalidRequest(time.Millisecond)

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

func TestRecordConfigReload(t *testing.T) {
	tests := []struct {
		name       string
		success    bool
		rulesCount int
		wantResult string
	}{
		{"success", true, 5, "success"},
		{"failure", false, 0, "failure"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTest(t)
			RecordConfigReload(tt.success, tt.rulesCount)

			got := testutil.ToFloat64(configReloadTotal.WithLabelValues(tt.wantResult))
			if got != 1 {
				t.Errorf("Expected %s count to be 1, got %v", tt.wantResult, got)
			}

			if tt.success {
				rulesGot := testutil.ToFloat64(configRulesTotal)
				if rulesGot != float64(tt.rulesCount) {
					t.Errorf(
						"Expected rules count to be %d, got %v",
						tt.rulesCount,
						rulesGot,
					)
				}
			}
		})
	}
}

func TestRecordDBUpdate(t *testing.T) {
	setupTest(t)

	entries := map[string]uint64{
		"country": 1000,
		"asn":     500,
	}
	duration := 2 * time.Second

	RecordDBUpdate(entries, duration)

	countryGot := testutil.ToFloat64(dbEntries.WithLabelValues("country"))
	if countryGot != 1000 {
		t.Errorf("Expected country entries to be 1000, got %v", countryGot)
	}

	asnGot := testutil.ToFloat64(dbEntries.WithLabelValues("asn"))
	if asnGot != 500 {
		t.Errorf("Expected asn entries to be 500, got %v", asnGot)
	}

	durationGot := testutil.ToFloat64(dbLoadDuration)
	if durationGot != 2.0 {
		t.Errorf("Expected load duration to be 2.0, got %v", durationGot)
	}
}

func BenchmarkRecordRequest(b *testing.B) {
	for range b.N {
		recordTestRequest(StatusAllowed, config.PolicyAllow)
	}
}

func BenchmarkRecordInvalidRequest(b *testing.B) {
	for range b.N {
		RecordInvalidRequest(time.Millisecond)
	}
}

func BenchmarkConcurrentRequests(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			recordTestRequest(StatusDenied, config.PolicyDeny)
			recordTestRequest(StatusAllowed, config.PolicyAllow)
			RecordInvalidRequest(time.Millisecond)
		}
	})
}
