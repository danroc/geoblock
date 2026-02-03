package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/danroc/geoblock/internal/config"
	"github.com/danroc/geoblock/internal/ipinfo"
	"github.com/danroc/geoblock/internal/metrics"
	"github.com/danroc/geoblock/internal/rules"
)

// nopDBUpdateCollector is a no-op collector for ipinfo.Resolver in tests.
type nopDBUpdateCollector struct{}

func (nopDBUpdateCollector) RecordDBUpdate(
	_ map[ipinfo.DBSource]uint64,
	_ time.Duration,
) {
}

// nopRequestCollector is a no-op collector for server tests.
type nopRequestCollector struct{}

func (nopRequestCollector) RecordRequest(
	_, _, _ string, _ time.Duration, _ int, _ string, _ bool,
) {
}

func (nopRequestCollector) RecordInvalidRequest(_ time.Duration) {}

// Test helpers

// assertStatus is a test helper that checks HTTP status codes.
func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
}

// newTestRequest creates a new HTTP request with the given headers.
func newTestRequest(
	method, path string,
	headers map[string]string,
) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req
}

// newAllowEngine creates a rules engine that allows all requests by default.
func newAllowEngine() *rules.Engine {
	return rules.NewEngine(&config.AccessControl{
		DefaultPolicy: config.PolicyAllow,
		Rules:         []config.AccessControlRule{},
	})
}

// testRoundTripper allows mocking HTTP responses for resolver testing.
type testRoundTripper struct {
	responses map[string]string
}

func (rt *testRoundTripper) RoundTrip(
	req *http.Request,
) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(
			bytes.NewBufferString(rt.responses[req.URL.String()]),
		),
	}, nil
}

// withTestTransport temporarily sets http.DefaultTransport to a test transport for the
// duration of fn.
func withTestTransport(testData map[string]string, fn func()) {
	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	http.DefaultTransport = &testRoundTripper{
		responses: testData,
	}
	fn()
}

// createTestResolver creates a resolver with mocked HTTP responses.
func createTestResolver(testData map[string]string) *ipinfo.Resolver {
	var resolver *ipinfo.Resolver
	withTestTransport(testData, func() {
		resolver = ipinfo.NewResolver(nopDBUpdateCollector{})
		_ = resolver.Update()
	})
	return resolver
}

func TestGetForwardAuth(t *testing.T) {
	resolver := ipinfo.NewResolver(nopDBUpdateCollector{})
	engine := newAllowEngine()
	tests := []struct {
		name    string
		headers map[string]string
		want    int
	}{
		{
			name: "missing X-Forwarded-For header",
			headers: map[string]string{
				headerForwardedHost:   "example.com",
				headerForwardedMethod: "GET",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "missing X-Forwarded-Host header",
			headers: map[string]string{
				headerForwardedFor:    "8.8.8.8",
				headerForwardedMethod: "GET",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "missing X-Forwarded-Method header",
			headers: map[string]string{
				headerForwardedFor:  "8.8.8.8",
				headerForwardedHost: "example.com",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "invalid IP address",
			headers: map[string]string{
				headerForwardedFor:    "invalid-ip",
				headerForwardedHost:   "example.com",
				headerForwardedMethod: "GET",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "empty headers",
			headers: map[string]string{
				headerForwardedFor:    "",
				headerForwardedHost:   "",
				headerForwardedMethod: "",
			},
			want: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newTestRequest("GET", "/v1/forward-auth", tt.headers)
			w := httptest.NewRecorder()

			getForwardAuth(w, req, resolver, engine, nopRequestCollector{})
			assertStatus(t, w.Code, tt.want)
		})
	}
}

func TestGetForwardAuthWithSpecificRules(t *testing.T) {
	testData := map[string]string{
		ipinfo.CountryIPv4URL: "8.8.8.8,8.8.8.8,US\n",
		ipinfo.CountryIPv6URL: "",
		ipinfo.ASNIPv4URL:     "8.8.8.8,8.8.8.8,15169,Google LLC\n",
		ipinfo.ASNIPv6URL:     "",
	}
	engine := rules.NewEngine(&config.AccessControl{
		DefaultPolicy: config.PolicyDeny,
		Rules: []config.AccessControlRule{
			{
				Policy:    config.PolicyAllow,
				Domains:   []string{"allowed.example.com"},
				Methods:   []string{"GET", "POST"},
				Countries: []string{"US"},
			},
		},
	})
	resolver := createTestResolver(testData)

	tests := []struct {
		name   string
		ip     string
		domain string
		method string
		want   int
	}{
		{
			name:   "allowed by rules",
			ip:     "8.8.8.8",
			domain: "allowed.example.com",
			method: "GET",
			want:   http.StatusNoContent,
		},
		{
			name:   "denied by domain",
			ip:     "8.8.8.8",
			domain: "blocked.example.com",
			method: "GET",
			want:   http.StatusForbidden,
		},
		{
			name:   "denied by method",
			ip:     "8.8.8.8",
			domain: "allowed.example.com",
			method: "DELETE",
			want:   http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				headerForwardedFor:    tt.ip,
				headerForwardedHost:   tt.domain,
				headerForwardedMethod: tt.method,
			}
			req := newTestRequest("GET", "/v1/forward-auth", headers)
			w := httptest.NewRecorder()

			getForwardAuth(w, req, resolver, engine, nopRequestCollector{})
			assertStatus(t, w.Code, tt.want)
		})
	}
}

func TestGetHealth(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/health", nil)
	w := httptest.NewRecorder()
	getHealth(w, req)
	assertStatus(t, w.Code, http.StatusNoContent)
}

func TestGetPrometheusMetrics(t *testing.T) {
	collector := metrics.NewCollector()
	resolver := ipinfo.NewResolver(nopDBUpdateCollector{})
	engine := newAllowEngine()
	server := New(":8080", engine, resolver, collector, metrics.Handler())

	// Record a request so that the requests_total metric appears in output
	collector.RecordRequest("allowed", "US", "GET", 0, 0, "allow", false)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	server.Handler.ServeHTTP(w, req)

	assertStatus(t, w.Code, http.StatusOK)

	body := w.Body.Bytes()
	if !bytes.Contains(body, []byte("geoblock_version_info")) {
		t.Error(
			"Prometheus output should contain geoblock_version_info metric",
		)
	}
	if !bytes.Contains(body, []byte("geoblock_requests_total")) {
		t.Error(
			"Prometheus output should contain geoblock_requests_total metric",
		)
	}
	if !bytes.Contains(body, []byte("# HELP")) {
		t.Error("Prometheus output should contain HELP comments")
	}
	if !bytes.Contains(body, []byte("# TYPE")) {
		t.Error("Prometheus output should contain TYPE comments")
	}
}

func TestNewServer(t *testing.T) {
	resolver := ipinfo.NewResolver(nopDBUpdateCollector{})
	engine := newAllowEngine()
	server := New(":8080", engine, resolver, nopRequestCollector{}, metrics.Handler())

	if got, want := server.Addr, ":8080"; got != want {
		t.Errorf("Addr = %q, want %q", got, want)
	}
	if server.Handler == nil {
		t.Error("Handler should not be nil")
	}
	if server.ReadTimeout <= 0 {
		t.Errorf("ReadTimeout = %v, want > 0", server.ReadTimeout)
	}
	if server.WriteTimeout <= 0 {
		t.Errorf("WriteTimeout = %v, want > 0", server.WriteTimeout)
	}
	if server.IdleTimeout <= 0 {
		t.Errorf("IdleTimeout = %v, want > 0", server.IdleTimeout)
	}
}

func TestServerEndpoints(t *testing.T) {
	resolver := ipinfo.NewResolver(nopDBUpdateCollector{})
	engine := newAllowEngine()
	server := New(":8080", engine, resolver, nopRequestCollector{}, metrics.Handler())
	tests := []struct {
		method string
		path   string
		want   int
	}{
		{"GET", "/v1/health", http.StatusNoContent},
		{"GET", "/metrics", http.StatusOK},
		{"GET", "/nonexistent", http.StatusNotFound},
		{"POST", "/v1/health", http.StatusMethodNotAllowed},
		{"POST", "/metrics", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			server.Handler.ServeHTTP(w, req)
			assertStatus(t, w.Code, tt.want)
		})
	}
}

func TestIsLocalIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
		desc string
	}{
		// RFC 1918 Class A private
		{"10.0.0.0", true, "start of 10.0.0.0/8"},
		{"10.255.255.255", true, "end of 10.0.0.0/8"},
		{"9.255.255.255", false, "before 10.0.0.0/8"},
		{"11.0.0.0", false, "after 10.0.0.0/8"},

		// RFC 1918 Class B private
		{"172.16.0.0", true, "start of 172.16.0.0/12"},
		{"172.31.255.255", true, "end of 172.16.0.0/12"},
		{"172.15.255.255", false, "before 172.16.0.0/12"},
		{"172.32.0.0", false, "after 172.16.0.0/12"},

		// RFC 1918 Class C private
		{"192.168.0.0", true, "start of 192.168.0.0/16"},
		{"192.168.255.255", true, "end of 192.168.0.0/16"},
		{"192.167.255.255", false, "before 192.168.0.0/16"},
		{"192.169.0.0", false, "after 192.168.0.0/16"},

		// RFC 1122 Loopback
		{"127.0.0.0", true, "start of loopback"},
		{"127.255.255.255", true, "end of loopback"},

		// RFC 3927 Link-local
		{"169.254.0.0", true, "start of link-local"},
		{"169.254.255.255", true, "end of link-local"},

		// IPv6 addresses
		{"::1", true, "IPv6 loopback"},
		{"fc00::", true, "start of IPv6 unique local"},
		{
			"fdff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			true,
			"end of IPv6 unique local",
		},
		{"fe80::", true, "start of IPv6 link-local"},
		{
			"febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			true,
			"end of IPv6 link-local",
		},

		// Public addresses
		{"8.8.8.8", false, "Google DNS"},
		{"1.1.1.1", false, "Cloudflare DNS"},
		{"2001:4860:4860::8888", false, "Google IPv6 DNS"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ip, err := netip.ParseAddr(tt.ip)
			if err != nil {
				t.Fatalf("ParseAddr(%q): %v", tt.ip, err)
			}
			if got := isLocalIP(ip); got != tt.want {
				t.Errorf("isLocalIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestParseForwardedFor(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "empty header",
			header: "",
			want:   "",
		},
		{
			name:   "single IP",
			header: "192.168.1.1",
			want:   "192.168.1.1",
		},
		{
			name:   "single IP with spaces",
			header: "  192.168.1.1  ",
			want:   "192.168.1.1",
		},
		{
			name:   "multiple IPs",
			header: "203.0.113.195,70.41.3.18,150.172.238.178",
			want:   "203.0.113.195",
		},
		{
			name:   "multiple IPs with spaces",
			header: "203.0.113.195, 70.41.3.18, 150.172.238.178",
			want:   "203.0.113.195",
		},
		{
			name:   "multiple IPs with extra spaces",
			header: "  203.0.113.195  ,  70.41.3.18  ,  150.172.238.178  ",
			want:   "203.0.113.195",
		},
		{
			name:   "IPv6 single",
			header: "2001:db8::1",
			want:   "2001:db8::1",
		},
		{
			name:   "IPv6 multiple",
			header: "2001:db8::1, 192.168.1.1",
			want:   "2001:db8::1",
		},
		{
			name:   "real-world example",
			header: "129.78.138.66, 129.78.64.103",
			want:   "129.78.138.66",
		},
		{
			name:   "only commas",
			header: ",,,",
			want:   "",
		},
		{
			name:   "single comma",
			header: ",",
			want:   "",
		},
		{
			name:   "empty string after split",
			header: ",192.168.1.1",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseForwardedFor(tt.header)
			if got != tt.want {
				t.Errorf(
					"parseForwardedFor(%q) = %q, want %q",
					tt.header,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestGetForwardAuthValidRequests(t *testing.T) {
	testData := map[string]string{
		ipinfo.CountryIPv4URL: "8.8.8.8,8.8.8.8,US\n",
		ipinfo.CountryIPv6URL: "",
		ipinfo.ASNIPv4URL:     "8.8.8.8,8.8.8.8,15169,Google LLC\n",
		ipinfo.ASNIPv6URL:     "",
	}
	resolver := createTestResolver(testData)
	engine := newAllowEngine()
	headers := map[string]string{
		headerForwardedFor:    "8.8.8.8",
		headerForwardedHost:   "example.com",
		headerForwardedMethod: "GET",
	}
	req := newTestRequest("GET", "/v1/forward-auth", headers)
	w := httptest.NewRecorder()

	getForwardAuth(w, req, resolver, engine, nopRequestCollector{})
	assertStatus(t, w.Code, http.StatusNoContent)
}

func TestGetForwardAuthMultipleForwardedIPs(t *testing.T) {
	testData := map[string]string{
		ipinfo.CountryIPv4URL: "8.8.8.8,8.8.8.8,US\n",
		ipinfo.CountryIPv6URL: "",
		ipinfo.ASNIPv4URL:     "8.8.8.8,8.8.8.8,15169,Google LLC\n",
		ipinfo.ASNIPv6URL:     "",
	}
	resolver := createTestResolver(testData)
	engine := newAllowEngine()

	tests := []struct {
		name           string
		forwardedFor   string
		expectedStatus int
	}{
		{
			name:           "multiple IPs - client IP first",
			forwardedFor:   "8.8.8.8, 192.168.1.1, 10.0.0.1",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "multiple IPs with extra spaces",
			forwardedFor:   "  8.8.8.8  ,  192.168.1.1  ",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "single IP with spaces",
			forwardedFor:   "  8.8.8.8  ",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				headerForwardedFor:    tt.forwardedFor,
				headerForwardedHost:   "example.com",
				headerForwardedMethod: "GET",
			}
			req := newTestRequest("GET", "/v1/forward-auth", headers)
			w := httptest.NewRecorder()

			getForwardAuth(w, req, resolver, engine, nopRequestCollector{})
			assertStatus(t, w.Code, tt.expectedStatus)
		})
	}
}

func TestServerHandlerSetup(t *testing.T) {
	resolver := ipinfo.NewResolver(nopDBUpdateCollector{})
	engine := newAllowEngine()
	server := New(":8080", engine, resolver, nopRequestCollector{}, metrics.Handler())

	req := httptest.NewRequest("GET", "/v1/forward-auth", nil)
	w := httptest.NewRecorder()

	server.Handler.ServeHTTP(w, req)
	assertStatus(t, w.Code, http.StatusBadRequest)
}
