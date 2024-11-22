package database_test

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/danroc/geoblock/pkg/database"
)

func TestStrIndex(t *testing.T) {
	tests := []struct {
		data     []string
		index    int
		expected string
	}{
		{[]string{"a", "b", "c"}, 0, "a"},
		{[]string{"a", "b", "c"}, 1, "b"},
		{[]string{"a", "b", "c"}, 2, "c"},
		{[]string{"a", "b", "c"}, 3, ""},
		{[]string{"a", "b", "c"}, -1, ""},
		{[]string{}, 0, ""},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := database.StrIndex(tt.data, tt.index)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestStrToASN(t *testing.T) {
	tests := []struct {
		input    string
		expected uint32
	}{
		{"12345", 12345},
		{"0", 0},
		{"4294967295", 4294967295},
		{"invalid", database.AS0},
		{"", database.AS0},
		{"-1", database.AS0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := database.StrToASN(tt.input)
			if result != tt.expected {
				t.Errorf("got %d, want %d", result, tt.expected)
			}
		})
	}
}

type mockRT struct {
	respond func(req *http.Request) (*http.Response, error)
}

func (m *mockRT) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.respond(req)
}

func newDummyRT() http.RoundTripper {
	return &mockRT{
		respond: func(req *http.Request) (*http.Response, error) {
			body := map[string]string{
				database.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n1.1.0.0,1.1.2.2,FR\n",
				database.CountryIPv6URL: "1:0::,1:1::,US\n1:2::,1:3::,FR\n",
				database.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n1.1.0.0,1.1.2.2,2,Test2\n",
				database.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n1:2::,1:3::,4,Test4\n",
			}[req.URL.String()]

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			}, nil
		},
	}
}

func newErrRT() http.RoundTripper {
	return &mockRT{
		respond: func(req *http.Request) (*http.Response, error) {
			return nil, io.ErrUnexpectedEOF
		},
	}
}

func withRT(rt http.RoundTripper, f func()) {
	original := http.DefaultTransport
	http.DefaultTransport = rt
	defer func() { http.DefaultTransport = original }()
	f()
}

func TestNewResolverError(t *testing.T) {
	withRT(newErrRT(), func() {
		_, err := database.NewResolver()
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}

func TestResolverResolve(t *testing.T) {
	withRT(newDummyRT(), func() {
		tests := []struct {
			ip      string
			country string
			org     string
			asn     uint32
		}{
			{"1.0.1.1", "US", "Test1", 1},
			{"1.1.1.1", "FR", "Test2", 2},
			{"1.2.1.1", "", "", database.AS0},
			{"1:0::", "US", "Test3", 3},
			{"1:2::", "FR", "Test4", 4},
			{"1:4::", "", "", database.AS0},
		}
		r, _ := database.NewResolver()
		for _, tt := range tests {
			t.Run(tt.ip, func(t *testing.T) {
				result := r.Resolve(net.ParseIP(tt.ip))
				if result.CountryCode != tt.country {
					t.Errorf("got %q, want %q", result.CountryCode, tt.country)
				}
				if result.ASN != tt.asn {
					t.Errorf("got %q, want %q", result.ASN, tt.asn)
				}
				if result.Organization != tt.org {
					t.Errorf("got %q, want %q", result.Organization, tt.org)
				}
			})
		}
	})
}
