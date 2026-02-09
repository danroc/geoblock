package ipinfo_test

import (
	"context"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/danroc/geoblock/internal/ipinfo"
)

// nopDBUpdateCollector is a no-op collector for testing.
type nopDBUpdateCollector struct{}

func (nopDBUpdateCollector) RecordDBUpdate(
	_ map[ipinfo.DBSource]uint64,
	_ time.Duration,
) {
}

// mapFetcher returns CSV records from a URL-keyed map.
type mapFetcher struct {
	dbs map[string]string
}

func (m *mapFetcher) Fetch(
	_ context.Context,
	url string,
) ([][]string, error) {
	return parseCSVString(m.dbs[url]), nil
}

// parseCSVString splits a raw CSV string into records.
func parseCSVString(s string) [][]string {
	if s == "" {
		return nil
	}
	var records [][]string
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		records = append(records, strings.Split(line, ","))
	}
	return records
}

// errFetcher always returns an error.
type errFetcher struct {
	err error
}

func (e *errFetcher) Fetch(
	_ context.Context,
	_ string,
) ([][]string, error) {
	return nil, e.err
}

func newDummyFetcher() ipinfo.Fetcher {
	return &mapFetcher{
		dbs: map[string]string{
			ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n1.1.0.0,1.1.2.2,FR\n",
			ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n1:2::,1:3::,FR\n",
			ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n1.1.0.0,1.1.2.2,2,Test2\n",
			ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n1:2::,1:3::,4,Test4\n",
		},
	}
}

func TestUpdateError(t *testing.T) {
	r := ipinfo.NewResolver(
		nopDBUpdateCollector{},
		&errFetcher{err: context.DeadlineExceeded},
	)
	if err := r.Update(context.Background()); err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		country string
		org     string
		asn     uint32
	}{
		{"IPv4 in US range", "1.0.1.1", "US", "Test1", 1},
		{"IPv4 in FR range", "1.1.1.1", "FR", "Test2", 2},
		{"IPv4 not found", "1.2.1.1", "", "", ipinfo.AS0},
		{"IPv6 in US range", "1:0::", "US", "Test3", 3},
		{"IPv6 in FR range", "1:2::", "FR", "Test4", 4},
		{"IPv6 not found", "1:4::", "", "", ipinfo.AS0},
	}

	r := ipinfo.NewResolver(nopDBUpdateCollector{}, newDummyFetcher())
	if err := r.Update(context.Background()); err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.Resolve(netip.MustParseAddr(tt.ip))
			if result.CountryCode != tt.country {
				t.Errorf("got %q, want %q", result.CountryCode, tt.country)
			}
			if result.ASN != tt.asn {
				t.Errorf("got %d, want %d", result.ASN, tt.asn)
			}
			if result.Organization != tt.org {
				t.Errorf("got %q, want %q", result.Organization, tt.org)
			}
		})
	}
}

func TestUpdateInvalidData(t *testing.T) {
	tests := []struct {
		name   string
		dbs    map[string]string
		errMsg string
	}{
		{
			name: "invalid start IP in country IPv4",
			dbs: map[string]string{
				ipinfo.CountryIPv4URL: "invalid,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			errMsg: "unable to parse IP",
		},
		{
			name: "invalid end IP in country IPv4",
			dbs: map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,invalid,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			errMsg: "unable to parse IP",
		},
		{
			name: "invalid start IP in country IPv6",
			dbs: map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "invalid,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			errMsg: "unable to parse IP",
		},
		{
			name: "invalid end IP in country IPv6",
			dbs: map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,invalid,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			errMsg: "unable to parse IP",
		},
		{
			name: "invalid start IP in ASN IPv4",
			dbs: map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "invalid,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			errMsg: "unable to parse IP",
		},
		{
			name: "invalid end IP in ASN IPv6",
			dbs: map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,invalid,3,Test3\n",
			},
			errMsg: "unable to parse IP",
		},
		{
			name: "extra field in ASN IPv4",
			dbs: map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1,extra\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			errMsg: "invalid record length",
		},
		{
			name: "missing field in ASN IPv4",
			dbs: map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,missing\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			errMsg: "invalid record length",
		},
		{
			name: "non-numeric ASN in ASN IPv6",
			dbs: map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,invalid,Test3\n",
			},
			errMsg: "invalid ASN",
		},
		{
			name: "missing country code in country IPv4",
			dbs: map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			errMsg: "invalid record length",
		},
		{
			name: "extra field in country IPv6",
			dbs: map[string]string{
				ipinfo.CountryIPv4URL: "1.0.0.0,1.0.2.2,US\n",
				ipinfo.CountryIPv6URL: "1:0::,1:1::,US,FR\n",
				ipinfo.ASNIPv4URL:     "1.0.0.0,1.0.2.2,1,Test1\n",
				ipinfo.ASNIPv6URL:     "1:0::,1:1::,3,Test3\n",
			},
			errMsg: "invalid record length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ipinfo.NewResolver(
				nopDBUpdateCollector{},
				&mapFetcher{dbs: tt.dbs},
			)
			err := r.Update(context.Background())
			if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf(
					"Update() error = %v, want substring %q", err, tt.errMsg,
				)
			}
		})
	}
}
