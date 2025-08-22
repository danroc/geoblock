package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/danroc/geoblock/internal/config"
	"github.com/danroc/geoblock/internal/ipres"
	"github.com/danroc/geoblock/internal/metrics"
	"github.com/danroc/geoblock/internal/rules"
)

// Test helpers

// assertStatus is a test helper that checks HTTP status codes.
func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
}

// assertContentType is a test helper that checks Content-Type headers.
func assertContentType(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
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

// withTestTransport temporarily sets http.DefaultTransport to a test transport
// for the duration of fn.
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
func createTestResolver(testData map[string]string) *ipres.Resolver {
	var resolver *ipres.Resolver
	withTestTransport(testData, func() {
		resolver = ipres.NewResolver()
		resolver.Update()
	})
	return resolver
}

func TestGetForwardAuth(t *testing.T) {
	resolver := ipres.NewResolver()
	engine := newAllowEngine()
	tests := []struct {
		name    string
		headers map[string]string
		want    int
	}{
		{
			name: "missing X-Forwarded-For header",
			headers: map[string]string{
				HeaderXForwardedHost:   "example.com",
				HeaderXForwardedMethod: "GET",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "missing X-Forwarded-Host header",
			headers: map[string]string{
				HeaderXForwardedFor:    "8.8.8.8",
				HeaderXForwardedMethod: "GET",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "missing X-Forwarded-Method header",
			headers: map[string]string{
				HeaderXForwardedFor:  "8.8.8.8",
				HeaderXForwardedHost: "example.com",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "invalid IP address",
			headers: map[string]string{
				HeaderXForwardedFor:    "invalid-ip",
				HeaderXForwardedHost:   "example.com",
				HeaderXForwardedMethod: "GET",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "empty headers",
			headers: map[string]string{
				HeaderXForwardedFor:    "",
				HeaderXForwardedHost:   "",
				HeaderXForwardedMethod: "",
			},
			want: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newTestRequest("GET", "/v1/forward-auth", tt.headers)
			w := httptest.NewRecorder()

			getForwardAuth(w, req, resolver, engine)
			assertStatus(t, w.Code, tt.want)
		})
	}
}

func TestGetForwardAuthWithSpecificRules(t *testing.T) {
	testData := map[string]string{
		ipres.CountryIPv4URL: "8.8.8.8,8.8.8.8,US\n",
		ipres.CountryIPv6URL: "",
		ipres.ASNIPv4URL:     "8.8.8.8,8.8.8.8,15169,Google LLC\n",
		ipres.ASNIPv6URL:     "",
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
				HeaderXForwardedFor:    tt.ip,
				HeaderXForwardedHost:   tt.domain,
				HeaderXForwardedMethod: tt.method,
			}
			req := newTestRequest("GET", "/v1/forward-auth", headers)
			w := httptest.NewRecorder()

			getForwardAuth(w, req, resolver, engine)
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

func TestGetMetrics(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/metrics", nil)
	w := httptest.NewRecorder()
	getMetrics(w, req)

	assertStatus(t, w.Code, http.StatusOK)
	assertContentType(t, w.Header().Get("Content-Type"), "application/json")

	var response metrics.Snapshot
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}
	if response.Version == "" {
		t.Error("version should not be empty")
	}
}

func TestGetMetricsJSONError(t *testing.T) {
	brokenWriter := &brokenResponseWriter{
		header: make(http.Header),
	}
	req := httptest.NewRequest("GET", "/v1/metrics", nil)
	getMetrics(brokenWriter, req)
	assertStatus(t, brokenWriter.statusCode, http.StatusOK)
}

// brokenResponseWriter is a ResponseWriter that fails on Write
type brokenResponseWriter struct {
	header     http.Header
	statusCode int
}

func (w *brokenResponseWriter) Header() http.Header {
	return w.header
}

func (w *brokenResponseWriter) Write([]byte) (int, error) {
	return 0, &json.UnsupportedTypeError{}
}

func (w *brokenResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func TestNewServer(t *testing.T) {
	resolver := ipres.NewResolver()
	engine := newAllowEngine()
	server := NewServer(":8080", engine, resolver)

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
	resolver := ipres.NewResolver()
	engine := newAllowEngine()
	server := NewServer(":8080", engine, resolver)
	tests := []struct {
		method string
		path   string
		want   int
	}{
		{"GET", "/v1/health", http.StatusNoContent},
		{"GET", "/v1/metrics", http.StatusOK},
		{"GET", "/nonexistent", http.StatusNotFound},
		{"POST", "/v1/health", http.StatusMethodNotAllowed},
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

func TestGetForwardAuthValidRequests(t *testing.T) {
	testData := map[string]string{
		ipres.CountryIPv4URL: "8.8.8.8,8.8.8.8,US\n",
		ipres.CountryIPv6URL: "",
		ipres.ASNIPv4URL:     "8.8.8.8,8.8.8.8,15169,Google LLC\n",
		ipres.ASNIPv6URL:     "",
	}
	resolver := createTestResolver(testData)
	engine := newAllowEngine()
	headers := map[string]string{
		HeaderXForwardedFor:    "8.8.8.8",
		HeaderXForwardedHost:   "example.com",
		HeaderXForwardedMethod: "GET",
	}
	req := newTestRequest("GET", "/v1/forward-auth", headers)
	w := httptest.NewRecorder()

	getForwardAuth(w, req, resolver, engine)
	assertStatus(t, w.Code, http.StatusNoContent)
}

func TestServerHandlerSetup(t *testing.T) {
	resolver := ipres.NewResolver()
	engine := newAllowEngine()
	server := NewServer(":8080", engine, resolver)

	req := httptest.NewRequest("GET", "/v1/forward-auth", nil)
	w := httptest.NewRecorder()

	server.Handler.ServeHTTP(w, req)
	assertStatus(t, w.Code, http.StatusBadRequest)
}
