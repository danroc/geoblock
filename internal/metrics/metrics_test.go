package metrics

import (
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/danroc/geoblock/internal/config"
	"github.com/danroc/geoblock/internal/ipinfo"
	"github.com/danroc/geoblock/internal/version"
)

// collector is a package-level test collector.
var collector = NewCollector()

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
	collector.RecordRequest(status, "US", "GET", time.Millisecond, 0, action, false)
}

// assertCounterValue asserts that a counter metric has the expected value.
func assertCounterValue(
	t *testing.T,
	collector prometheus.Collector,
	expected float64,
	name string,
) {
	t.Helper()
	if got := testutil.ToFloat64(collector); got != expected {
		t.Errorf("Expected %s to be %v, got %v", name, expected, got)
	}
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

func TestRecordRequestEmptyCountry(t *testing.T) {
	setupTest(t)

	collector.RecordRequest(
		StatusAllowed,
		"",
		"GET",
		time.Millisecond,
		0,
		config.PolicyAllow,
		false,
	)

	// Request should still be counted
	if got := getCounterValue(StatusAllowed); got != 1 {
		t.Errorf("Expected allowed count to be 1, got %v", got)
	}

	// Country counter should not be incremented for empty country
	if got := testutil.ToFloat64(requestsByCountry.WithLabelValues("")); got != 0 {
		t.Errorf("Expected empty country count to be 0, got %v", got)
	}
}

func TestRecordRequestEmptyMethod(t *testing.T) {
	setupTest(t)

	collector.RecordRequest(
		StatusAllowed,
		"US",
		"",
		time.Millisecond,
		0,
		config.PolicyAllow,
		false,
	)

	// Request should still be counted
	if got := getCounterValue(StatusAllowed); got != 1 {
		t.Errorf("Expected allowed count to be 1, got %v", got)
	}

	// Method counter should not be incremented for empty method
	if got := testutil.ToFloat64(requestsByMethod.WithLabelValues("")); got != 0 {
		t.Errorf("Expected empty method count to be 0, got %v", got)
	}
}

func TestRecordRequestDefaultPolicy(t *testing.T) {
	setupTest(t)

	collector.RecordRequest(
		StatusAllowed,
		"US",
		"GET",
		time.Millisecond,
		0,
		config.PolicyAllow,
		true,
	)

	// Request should be counted
	if got := getCounterValue(StatusAllowed); got != 1 {
		t.Errorf("Expected allowed count to be 1, got %v", got)
	}

	// Default policy counter should be incremented
	if got := testutil.ToFloat64(
		defaultPolicyMatches.WithLabelValues(config.PolicyAllow),
	); got != 1 {
		t.Errorf("Expected default policy allow count to be 1, got %v", got)
	}

	// Rule matches should NOT be incremented
	if got := testutil.ToFloat64(
		ruleMatches.WithLabelValues("0", config.PolicyAllow),
	); got != 0 {
		t.Errorf("Expected rule matches to be 0 for default policy, got %v", got)
	}
}

func TestRecordRequestByCountry(t *testing.T) {
	setupTest(t)

	collector.RecordRequest(
		StatusAllowed,
		"US",
		"GET",
		time.Millisecond,
		0,
		config.PolicyAllow,
		false,
	)
	collector.RecordRequest(
		StatusAllowed,
		"US",
		"GET",
		time.Millisecond,
		0,
		config.PolicyAllow,
		false,
	)
	collector.RecordRequest(
		StatusAllowed,
		"DE",
		"GET",
		time.Millisecond,
		0,
		config.PolicyAllow,
		false,
	)

	assertCounterValue(t, requestsByCountry.WithLabelValues("US"), 2, "US count")
	assertCounterValue(t, requestsByCountry.WithLabelValues("DE"), 1, "DE count")
}

func TestRecordRequestByMethod(t *testing.T) {
	setupTest(t)

	collector.RecordRequest(
		StatusAllowed,
		"US",
		"GET",
		time.Millisecond,
		0,
		config.PolicyAllow,
		false,
	)
	collector.RecordRequest(
		StatusAllowed,
		"US",
		"POST",
		time.Millisecond,
		0,
		config.PolicyAllow,
		false,
	)
	collector.RecordRequest(
		StatusAllowed,
		"US",
		"GET",
		time.Millisecond,
		0,
		config.PolicyAllow,
		false,
	)

	assertCounterValue(t, requestsByMethod.WithLabelValues("GET"), 2, "GET count")
	assertCounterValue(t, requestsByMethod.WithLabelValues("POST"), 1, "POST count")
}

func TestRecordRequestRuleMatches(t *testing.T) {
	setupTest(t)

	collector.RecordRequest(
		StatusAllowed,
		"US",
		"GET",
		time.Millisecond,
		0,
		config.PolicyAllow,
		false,
	)
	collector.RecordRequest(
		StatusDenied,
		"US",
		"GET",
		time.Millisecond,
		1,
		config.PolicyDeny,
		false,
	)
	collector.RecordRequest(
		StatusDenied,
		"US",
		"GET",
		time.Millisecond,
		1,
		config.PolicyDeny,
		false,
	)
	collector.RecordRequest(
		StatusAllowed,
		"US",
		"GET",
		time.Millisecond,
		2,
		config.PolicyAllow,
		false,
	)

	assertCounterValue(
		t,
		ruleMatches.WithLabelValues("0", config.PolicyAllow),
		1,
		"rule 0 allow",
	)
	assertCounterValue(
		t,
		ruleMatches.WithLabelValues("1", config.PolicyDeny),
		2,
		"rule 1 deny",
	)
	assertCounterValue(
		t,
		ruleMatches.WithLabelValues("2", config.PolicyAllow),
		1,
		"rule 2 allow",
	)
}

func TestRecordRequestDuration(t *testing.T) {
	setupTest(t)

	collector.RecordRequest(
		StatusAllowed,
		"US",
		"GET",
		50*time.Millisecond,
		0,
		config.PolicyAllow,
		false,
	)
	collector.RecordRequest(
		StatusDenied,
		"US",
		"GET",
		100*time.Millisecond,
		0,
		config.PolicyDeny,
		false,
	)

	// Verify histogram has recorded observations by checking the collector count
	// Each histogram label creates multiple metrics (one per bucket + sum + count)
	count := testutil.CollectAndCount(requestDuration)
	if count == 0 {
		t.Error("Expected request duration histogram to have recorded metrics")
	}
}

func TestRecordInvalidRequest(t *testing.T) {
	setupTest(t)
	collector.RecordInvalidRequest(time.Millisecond)
	if got := getCounterValue(StatusInvalid); got != 1 {
		t.Errorf("Expected invalid count to be 1, got %v", got)
	}
}

func TestMultipleRequests(t *testing.T) {
	setupTest(t)

	for range 5 {
		recordTestRequest(StatusDenied, config.PolicyDeny)
	}
	for range 3 {
		recordTestRequest(StatusAllowed, config.PolicyAllow)
	}
	for range 2 {
		collector.RecordInvalidRequest(time.Millisecond)
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
				collector.RecordInvalidRequest(time.Millisecond)
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
	collector.RecordInvalidRequest(time.Millisecond)

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
	collector.RecordInvalidRequest(time.Millisecond)

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

			before := time.Now().Unix()
			collector.RecordConfigReload(tt.success, tt.rulesCount)
			after := time.Now().Unix()

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

				// Verify timestamp was recorded
				timestamp := testutil.ToFloat64(configLastReload)
				if timestamp < float64(before) || timestamp > float64(after) {
					t.Errorf(
						"Expected configLastReload timestamp between %d and %d, got %v",
						before,
						after,
						timestamp,
					)
				}
			} else {
				// Verify timestamp was NOT updated on failure
				timestamp := testutil.ToFloat64(configLastReload)
				if timestamp != 0 {
					t.Errorf(
						"Expected configLastReload to be 0 on failure, got %v",
						timestamp,
					)
				}
			}
		})
	}
}

func TestRecordDBUpdate(t *testing.T) {
	setupTest(t)

	entries := map[ipinfo.DBSource]uint64{
		{DBType: ipinfo.DBTypeCountry, IPVersion: ipinfo.IPVersion4}: 1000,
		{DBType: ipinfo.DBTypeCountry, IPVersion: ipinfo.IPVersion6}: 800,
		{DBType: ipinfo.DBTypeASN, IPVersion: ipinfo.IPVersion4}:     500,
		{DBType: ipinfo.DBTypeASN, IPVersion: ipinfo.IPVersion6}:     400,
	}
	duration := 2 * time.Second

	before := time.Now().Unix()
	collector.RecordDBUpdate(entries, duration)
	after := time.Now().Unix()

	countryV4Got := testutil.ToFloat64(dbEntries.WithLabelValues("country", "4"))
	if countryV4Got != 1000 {
		t.Errorf("Expected country IPv4 entries to be 1000, got %v", countryV4Got)
	}

	countryV6Got := testutil.ToFloat64(dbEntries.WithLabelValues("country", "6"))
	if countryV6Got != 800 {
		t.Errorf("Expected country IPv6 entries to be 800, got %v", countryV6Got)
	}

	asnV4Got := testutil.ToFloat64(dbEntries.WithLabelValues("asn", "4"))
	if asnV4Got != 500 {
		t.Errorf("Expected asn IPv4 entries to be 500, got %v", asnV4Got)
	}

	asnV6Got := testutil.ToFloat64(dbEntries.WithLabelValues("asn", "6"))
	if asnV6Got != 400 {
		t.Errorf("Expected asn IPv6 entries to be 400, got %v", asnV6Got)
	}

	durationGot := testutil.ToFloat64(dbLoadDuration)
	if durationGot != 2.0 {
		t.Errorf("Expected load duration to be 2.0, got %v", durationGot)
	}

	// Verify timestamp was recorded
	timestamp := testutil.ToFloat64(dbLastUpdate)
	if timestamp < float64(before) || timestamp > float64(after) {
		t.Errorf(
			"Expected dbLastUpdate timestamp between %d and %d, got %v",
			before,
			after,
			timestamp,
		)
	}
}

func BenchmarkRecordRequest(b *testing.B) {
	for range b.N {
		recordTestRequest(StatusAllowed, config.PolicyAllow)
	}
}

func BenchmarkRecordInvalidRequest(b *testing.B) {
	for range b.N {
		collector.RecordInvalidRequest(time.Millisecond)
	}
}

func BenchmarkConcurrentRequests(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			recordTestRequest(StatusDenied, config.PolicyDeny)
			recordTestRequest(StatusAllowed, config.PolicyAllow)
			collector.RecordInvalidRequest(time.Millisecond)
		}
	})
}
