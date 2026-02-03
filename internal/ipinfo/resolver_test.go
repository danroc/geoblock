package ipinfo_test

import (
	"bytes"
	"io"
	"net/http"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/danroc/geoblock/internal/ipinfo"
)

// nopDBUpdateCollector is a no-op collector for testing.
type nopDBUpdateCollector struct{}

func (nopDBUpdateCollector) RecordDBUpdate(_ map[string]uint64, _ time.Duration) {}

type mockRT struct {
	respond func(req *http.Request) (*http.Response, error)
}

func (m *mockRT) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.respond(req)
}

func newRTWithDBs(dbs map[string]string) http.RoundTripper {
	return &mockRT{
		respond: func(req *http.Request) (*http.Response, error) {
			body := dbs[req.URL.String()]
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			}, nil
		},
	}
}

func newDummyRT() http.RoundTripper {
	dummyDatabases := map[string]string{
		ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n1.1.0.0,1.1.2.2,FR\n",
		ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n1:2::,1:3::,FR\n",
		ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n1.1.0.0,1.1.2.2,2,Test2\n",
		ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n1:2::,1:3::,4,Test4\n",
	}
	return newRTWithDBs(dummyDatabases)
}

func newErrRT() http.RoundTripper {
	return &mockRT{
		respond: func(_ *http.Request) (*http.Response, error) {
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

func TestUpdateError(t *testing.T) {
	withRT(newErrRT(), func() {
		r := ipinfo.NewResolver(nopDBUpdateCollector{})
		if err := r.Update(); err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}

func TestResolve(t *testing.T) {
	withRT(newDummyRT(), func() {
		tests := []struct {
			ip      string
			country string
			org     string
			asn     uint32
		}{
			{"1.0.1.1", "US", "Test1", 1},
			{"1.1.1.1", "FR", "Test2", 2},
			{"1.2.1.1", "", "", ipinfo.AS0},
			{"1:0::", "US", "Test3", 3},
			{"1:2::", "FR", "Test4", 4},
			{"1:4::", "", "", ipinfo.AS0},
		}
		r := ipinfo.NewResolver(nopDBUpdateCollector{})
		if err := r.Update(); err != nil {
			t.Fatal(err)
		}
		for _, tt := range tests {
			t.Run(tt.ip, func(t *testing.T) {
				result := r.Resolve(netip.MustParseAddr(tt.ip))
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

func TestUpdateInvalidData(t *testing.T) {
	tests := []struct {
		dbs    map[string]string
		errMsg string
	}{
		{
			map[string]string{
				ipinfo.CountryIPv4URL: "invalid,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			"unable to parse IP",
		},
		{
			map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,invalid,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			"unable to parse IP",
		},
		{
			map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "invalid,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			"unable to parse IP",
		},
		{
			map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,invalid,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			"unable to parse IP",
		},
		{
			map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "invalid,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			"unable to parse IP",
		},
		{
			map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,invalid,3,Test3\n",
			},
			"unable to parse IP",
		},
		{
			map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1,extra\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			"invalid record length",
		},
		{
			map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,missing\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			"invalid record length",
		},
		{
			map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,invalid,Test3\n",
			},
			"invalid ASN",
		},
		{
			map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			"invalid record length",
		},
		{
			map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US,FR\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			"invalid record length",
		},
	}

	for _, tt := range tests {
		withRT(newRTWithDBs(tt.dbs), func() {
			r := ipinfo.NewResolver(nopDBUpdateCollector{})
			err := r.Update()
			if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("got %v, want %v", err, tt.errMsg)
			}
		})
	}
}
