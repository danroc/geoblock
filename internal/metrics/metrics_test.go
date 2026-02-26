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
)

// Record a request with sensible defaults for tests that don't care about specific
// country, method, duration, or rule values.
func record(c *PrometheusCollector, status string) {
	action := config.PolicyAllow
	if status == StatusDenied {
		action = config.PolicyDeny
	}
	c.RecordRequest(RequestRecord{
		Status:   status,
		Country:  "FR",
		Method:   "GET",
		Duration: time.Millisecond,
		Action:   action,
	})
}

// getCounterValue returns the current value of a counter with the given label.
func getCounterValue(c *PrometheusCollector, status string) float64 {
	return testutil.ToFloat64(c.requestsTotal.WithLabelValues(status))
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
			c := NewCollector()

			c.RecordRequest(RequestRecord{
				Status:   tt.status,
				Country:  "FR",
				Method:   "GET",
				Duration: time.Millisecond,
				Action:   tt.action,
			})

			if got := getCounterValue(c, tt.status); got != 1 {
				t.Errorf("Expected %s count to be 1, got %v", tt.status, got)
			}
		})
	}
}

func TestRecordRequest_EmptyCountry(t *testing.T) {
	c := NewCollector()

	c.RecordRequest(RequestRecord{
		Status:   StatusAllowed,
		Method:   "GET",
		Duration: time.Millisecond,
		Action:   config.PolicyAllow,
	})

	// Request should still be counted
	if got := getCounterValue(c, StatusAllowed); got != 1 {
		t.Errorf("Expected allowed count to be 1, got %v", got)
	}

	// Country counter should not be incremented for empty country
	if got := testutil.ToFloat64(c.requestsByCountry.WithLabelValues("")); got != 0 {
		t.Errorf("Expected empty country count to be 0, got %v", got)
	}
}

func TestRecordRequest_EmptyMethod(t *testing.T) {
	c := NewCollector()

	c.RecordRequest(RequestRecord{
		Status:   StatusAllowed,
		Country:  "FR",
		Duration: time.Millisecond,
		Action:   config.PolicyAllow,
	})

	// Request should still be counted
	if got := getCounterValue(c, StatusAllowed); got != 1 {
		t.Errorf("Expected allowed count to be 1, got %v", got)
	}

	// Method counter should not be incremented for empty method
	if got := testutil.ToFloat64(c.requestsByMethod.WithLabelValues("")); got != 0 {
		t.Errorf("Expected empty method count to be 0, got %v", got)
	}
}

func TestRecordRequest_DefaultPolicy(t *testing.T) {
	c := NewCollector()

	c.RecordRequest(RequestRecord{
		Status:          StatusAllowed,
		Country:         "FR",
		Method:          "GET",
		Duration:        time.Millisecond,
		Action:          config.PolicyAllow,
		IsDefaultPolicy: true,
	})

	// Request should be counted
	if got := getCounterValue(c, StatusAllowed); got != 1 {
		t.Errorf("Expected allowed count to be 1, got %v", got)
	}

	// Default policy counter should be incremented
	if got := testutil.ToFloat64(
		c.defaultPolicyMatches.WithLabelValues(config.PolicyAllow),
	); got != 1 {
		t.Errorf("Expected default policy allow count to be 1, got %v", got)
	}

	// Rule matches should NOT be incremented
	if got := testutil.ToFloat64(
		c.ruleMatches.WithLabelValues("0", config.PolicyAllow),
	); got != 0 {
		t.Errorf("Expected rule matches to be 0 for default policy, got %v", got)
	}
}

func TestRecordRequest_ByCountry(t *testing.T) {
	c := NewCollector()

	c.RecordRequest(RequestRecord{
		Status: StatusAllowed, Country: "FR", Method: "GET",
		Duration: time.Millisecond, Action: config.PolicyAllow,
	})
	c.RecordRequest(RequestRecord{
		Status: StatusAllowed, Country: "FR", Method: "GET",
		Duration: time.Millisecond, Action: config.PolicyAllow,
	})
	c.RecordRequest(RequestRecord{
		Status: StatusAllowed, Country: "BR", Method: "GET",
		Duration: time.Millisecond, Action: config.PolicyAllow,
	})

	assertCounterValue(t, c.requestsByCountry.WithLabelValues("FR"), 2, "FR count")
	assertCounterValue(t, c.requestsByCountry.WithLabelValues("BR"), 1, "BR count")
}

func TestRecordRequest_ByMethod(t *testing.T) {
	c := NewCollector()

	c.RecordRequest(RequestRecord{
		Status: StatusAllowed, Country: "FR", Method: "GET",
		Duration: time.Millisecond, Action: config.PolicyAllow,
	})
	c.RecordRequest(RequestRecord{
		Status: StatusAllowed, Country: "FR", Method: "POST",
		Duration: time.Millisecond, Action: config.PolicyAllow,
	})
	c.RecordRequest(RequestRecord{
		Status: StatusAllowed, Country: "FR", Method: "GET",
		Duration: time.Millisecond, Action: config.PolicyAllow,
	})

	assertCounterValue(t, c.requestsByMethod.WithLabelValues("GET"), 2, "GET count")
	assertCounterValue(t, c.requestsByMethod.WithLabelValues("POST"), 1, "POST count")
}

func TestRecordRequest_RuleMatches(t *testing.T) {
	c := NewCollector()

	c.RecordRequest(RequestRecord{
		Status: StatusAllowed, Country: "FR", Method: "GET",
		Duration: time.Millisecond, Action: config.PolicyAllow,
	})
	c.RecordRequest(RequestRecord{
		Status: StatusDenied, Country: "FR", Method: "GET",
		Duration: time.Millisecond, RuleIndex: 1,
		Action: config.PolicyDeny,
	})
	c.RecordRequest(RequestRecord{
		Status: StatusDenied, Country: "FR", Method: "GET",
		Duration: time.Millisecond, RuleIndex: 1,
		Action: config.PolicyDeny,
	})
	c.RecordRequest(RequestRecord{
		Status: StatusAllowed, Country: "FR", Method: "GET",
		Duration: time.Millisecond, RuleIndex: 2,
		Action: config.PolicyAllow,
	})

	assertCounterValue(
		t, c.ruleMatches.WithLabelValues("0", config.PolicyAllow), 1, "rule 0 allow",
	)
	assertCounterValue(
		t, c.ruleMatches.WithLabelValues("1", config.PolicyDeny), 2, "rule 1 deny",
	)
	assertCounterValue(
		t, c.ruleMatches.WithLabelValues("2", config.PolicyAllow), 1, "rule 2 allow",
	)
}

func TestRecordRequest_Duration(t *testing.T) {
	c := NewCollector()

	c.RecordRequest(RequestRecord{
		Status: StatusAllowed, Country: "FR", Method: "GET",
		Duration: 50 * time.Millisecond, Action: config.PolicyAllow,
	})
	c.RecordRequest(RequestRecord{
		Status: StatusDenied, Country: "FR", Method: "GET",
		Duration: 100 * time.Millisecond, Action: config.PolicyDeny,
	})

	// Verify histogram has recorded observations by checking the collector count. Each
	// histogram label creates multiple metrics (one per bucket + sum + count).
	count := testutil.CollectAndCount(c.requestDuration)
	if count == 0 {
		t.Error("Expected request duration histogram to have recorded metrics")
	}
}

func TestRecordInvalidRequest(t *testing.T) {
	c := NewCollector()

	c.RecordInvalidRequest(time.Millisecond)

	if got := getCounterValue(c, StatusInvalid); got != 1 {
		t.Errorf("Expected invalid count to be 1, got %v", got)
	}
}

func TestMultipleRequests(t *testing.T) {
	c := NewCollector()

	for range 5 {
		record(c, StatusDenied)
	}
	for range 3 {
		record(c, StatusAllowed)
	}
	for range 2 {
		c.RecordInvalidRequest(time.Millisecond)
	}

	if got := getCounterValue(c, "denied"); got != 5 {
		t.Errorf("Expected denied count to be 5, got %v", got)
	}
	if got := getCounterValue(c, "allowed"); got != 3 {
		t.Errorf("Expected allowed count to be 3, got %v", got)
	}
	if got := getCounterValue(c, "invalid"); got != 2 {
		t.Errorf("Expected invalid count to be 2, got %v", got)
	}
}

func TestConcurrentRequests(t *testing.T) {
	c := NewCollector()

	const numGoroutines = 100
	const requestsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3)

	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range requestsPerGoroutine {
				record(c, StatusDenied)
			}
		}()
	}
	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range requestsPerGoroutine {
				record(c, StatusAllowed)
			}
		}()
	}
	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range requestsPerGoroutine {
				c.RecordInvalidRequest(time.Millisecond)
			}
		}()
	}

	wg.Wait()

	expected := float64(numGoroutines * requestsPerGoroutine)
	for _, status := range []string{StatusDenied, StatusAllowed, StatusInvalid} {
		if got := getCounterValue(c, status); got != expected {
			t.Errorf("Expected %s count to be %v, got %v", status, expected, got)
		}
	}
}

func TestHandler(t *testing.T) {
	c := NewCollector()

	record(c, StatusAllowed)
	record(c, StatusAllowed)
	record(c, StatusDenied)
	c.RecordInvalidRequest(time.Millisecond)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	c.Handler().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()

	expectedStrings := []string{
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
			c := NewCollector()

			before := time.Now().Unix()
			c.RecordConfigReload(tt.success, tt.rulesCount)
			after := time.Now().Unix()

			if got := testutil.ToFloat64(
				c.configReloadTotal.WithLabelValues(tt.wantResult),
			); got != 1 {
				t.Errorf("Expected %s count to be 1, got %v", tt.wantResult, got)
			}

			if tt.success {
				rulesGot := testutil.ToFloat64(c.configRulesTotal)
				if rulesGot != float64(tt.rulesCount) {
					t.Errorf(
						"Expected rules count to be %d, got %v",
						tt.rulesCount, rulesGot,
					)
				}

				// Verify timestamp was recorded
				timestamp := testutil.ToFloat64(c.configLastReload)
				if timestamp < float64(before) || timestamp > float64(after) {
					t.Errorf(
						"Expected configLastReload timestamp between %d and %d, got %v",
						before, after, timestamp,
					)
				}
			} else {
				// Verify timestamp was NOT updated on failure
				timestamp := testutil.ToFloat64(c.configLastReload)
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
	c := NewCollector()

	entries := map[ipinfo.DBSource]uint64{
		{DBType: ipinfo.DBTypeCountry, IPVersion: ipinfo.IPVersion4}: 1000,
		{DBType: ipinfo.DBTypeCountry, IPVersion: ipinfo.IPVersion6}: 800,
		{DBType: ipinfo.DBTypeASN, IPVersion: ipinfo.IPVersion4}:     500,
		{DBType: ipinfo.DBTypeASN, IPVersion: ipinfo.IPVersion6}:     400,
	}
	duration := 2 * time.Second

	before := time.Now().Unix()
	c.RecordDBUpdate(entries, duration)
	after := time.Now().Unix()

	countryV4Got := testutil.ToFloat64(c.dbEntries.WithLabelValues("country", "4"))
	if countryV4Got != 1000 {
		t.Errorf("Expected country IPv4 entries to be 1000, got %v", countryV4Got)
	}

	countryV6Got := testutil.ToFloat64(c.dbEntries.WithLabelValues("country", "6"))
	if countryV6Got != 800 {
		t.Errorf("Expected country IPv6 entries to be 800, got %v", countryV6Got)
	}

	asnV4Got := testutil.ToFloat64(c.dbEntries.WithLabelValues("asn", "4"))
	if asnV4Got != 500 {
		t.Errorf("Expected asn IPv4 entries to be 500, got %v", asnV4Got)
	}

	asnV6Got := testutil.ToFloat64(c.dbEntries.WithLabelValues("asn", "6"))
	if asnV6Got != 400 {
		t.Errorf("Expected asn IPv6 entries to be 400, got %v", asnV6Got)
	}

	durationGot := testutil.ToFloat64(c.dbLoadDuration)
	if durationGot != 2.0 {
		t.Errorf("Expected load duration to be 2.0, got %v", durationGot)
	}

	// Verify timestamp was recorded
	timestamp := testutil.ToFloat64(c.dbLastUpdate)
	if timestamp < float64(before) || timestamp > float64(after) {
		t.Errorf(
			"Expected dbLastUpdate timestamp between %d and %d, got %v",
			before, after, timestamp,
		)
	}
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

func BenchmarkRecordRequest(b *testing.B) {
	c := NewCollector()
	for range b.N {
		record(c, StatusAllowed)
	}
}

func BenchmarkRecordInvalidRequest(b *testing.B) {
	c := NewCollector()
	for range b.N {
		c.RecordInvalidRequest(time.Millisecond)
	}
}

func BenchmarkConcurrentRequests(b *testing.B) {
	c := NewCollector()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			record(c, StatusDenied)
			record(c, StatusAllowed)
			c.RecordInvalidRequest(time.Millisecond)
		}
	})
}
