package ipinfo_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/danroc/geoblock/internal/ipinfo"
	"github.com/danroc/geoblock/internal/itree"
)

type fakeFetcher struct {
	records [][]string
	err     error
}

func (f *fakeFetcher) Fetch(_ context.Context, _ string) ([][]string, error) {
	return f.records, f.err
}

func TestLoader_LoadCountry(t *testing.T) {
	db := itree.NewTree[netip.Addr, ipinfo.Resolution]()
	fetcher := &fakeFetcher{
		records: [][]string{
			{"1.1.1.0", "1.1.1.255", "AU"},
		},
	}

	loader := ipinfo.NewLoader(fetcher)
	count, err := loader.Load(context.Background(), db, ipinfo.DBSourceSpec{
		Parser: ipinfo.ParseCountryRecord,
	})
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if count != 1 {
		t.Errorf("Load() count = %d, want 1", count)
	}

	res := db.Query(netip.MustParseAddr("1.1.1.100"))
	if len(res) != 1 {
		t.Fatalf("Query() got %d results, want 1", len(res))
	}
	if res[0].CountryCode != "AU" {
		t.Errorf("CountryCode = %q, want %q", res[0].CountryCode, "AU")
	}
}

func TestLoader_LoadASN(t *testing.T) {
	db := itree.NewTree[netip.Addr, ipinfo.Resolution]()
	fetcher := &fakeFetcher{
		records: [][]string{
			{"8.8.8.0", "8.8.8.255", "15169", "Google LLC"},
		},
	}

	loader := ipinfo.NewLoader(fetcher)
	count, err := loader.Load(context.Background(), db, ipinfo.DBSourceSpec{
		Parser: ipinfo.ParseASNRecord,
	})
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if count != 1 {
		t.Errorf("Load() count = %d, want 1", count)
	}

	res := db.Query(netip.MustParseAddr("8.8.8.8"))
	if len(res) != 1 {
		t.Fatalf("Query() got %d results, want 1", len(res))
	}
	if res[0].ASN != 15169 {
		t.Errorf("ASN = %d, want 15169", res[0].ASN)
	}
	if res[0].Organization != "Google LLC" {
		t.Errorf("Organization = %q, want %q", res[0].Organization, "Google LLC")
	}
}

func TestLoader_LoadFetchError(t *testing.T) {
	db := itree.NewTree[netip.Addr, ipinfo.Resolution]()
	fetchErr := errors.New("network error")
	fetcher := &fakeFetcher{err: fetchErr}

	loader := ipinfo.NewLoader(fetcher)
	count, err := loader.Load(context.Background(), db, ipinfo.DBSourceSpec{
		Parser: ipinfo.ParseCountryRecord,
	})
	if !errors.Is(err, fetchErr) {
		t.Errorf("Load() error = %v, want %v", err, fetchErr)
	}
	if count != 0 {
		t.Errorf("Load() count = %d, want 0", count)
	}
}

func TestLoader_LoadParseError(t *testing.T) {
	db := itree.NewTree[netip.Addr, ipinfo.Resolution]()
	fetcher := &fakeFetcher{
		records: [][]string{
			{"1.1.1.0", "1.1.1.255", "AU"}, // valid
			{"invalid", "1.1.1.255", "US"}, // invalid start IP
			{"2.2.2.0", "2.2.2.255", "FR"}, // valid
			{"3.3.3.0", "3.3.3.255"},       // missing country (wrong length)
		},
	}

	loader := ipinfo.NewLoader(fetcher)
	count, err := loader.Load(context.Background(), db, ipinfo.DBSourceSpec{
		Parser: ipinfo.ParseCountryRecord,
	})
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
	if count != 2 {
		t.Errorf("Load() count = %d, want 2", count)
	}
}
