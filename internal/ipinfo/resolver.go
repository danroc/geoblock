// Package ipinfo provides an IP resolver that returns information about an IP address.
package ipinfo

import (
	"context"
	"errors"
	"net/netip"
	"sync/atomic"
	"time"

	"github.com/danroc/geoblock/internal/itree"
)

// ResTree is a type alias for an interval tree that maps IP addresses to resolutions.
type ResTree = itree.ITree[netip.Addr, Resolution]

// Resolution contains the result of resolving an IP address.
type Resolution struct {
	CountryCode  string // ISO 3166-1 alpha-2 country code
	Organization string // Organization name
	ASN          uint32 // Autonomous System Number
}

// mergeResolutions combines multiple Resolution objects by taking the last non-zero
// value for each field. This implements a "last-write-wins" strategy where later
// resolutions override earlier ones.
func mergeResolutions(resolutions []Resolution) Resolution {
	var merged Resolution
	for _, r := range resolutions {
		if r.CountryCode != "" {
			merged.CountryCode = r.CountryCode
		}
		if r.Organization != "" {
			merged.Organization = r.Organization
		}
		if r.ASN != 0 {
			merged.ASN = r.ASN
		}
	}
	return merged
}

// DBUpdateCollector collects metrics for database updates.
type DBUpdateCollector interface {
	RecordDBUpdate(entries map[DBSource]uint64, duration time.Duration)
}

// Database type constants for metrics.
const (
	DBTypeCountry = "country"
	DBTypeASN     = "asn"
)

// IP version constants for metrics.
const (
	IPVersion4 = "4"
	IPVersion6 = "6"
)

// DBSource identifies a database entry by type and IP version.
type DBSource struct {
	DBType    string
	IPVersion string
}

// DBSourceSpec defines a database source with its URL and parser.
type DBSourceSpec struct {
	Source DBSource
	URL    string
	Parser ParserFn
}

// defaultSources defines all database sources to fetch during updates.
var defaultSources = []DBSourceSpec{
	{
		Source: DBSource{DBTypeCountry, IPVersion4},
		URL:    CountryIPv4URL,
		Parser: ParseCountryRecord,
	},
	{
		Source: DBSource{DBTypeCountry, IPVersion6},
		URL:    CountryIPv6URL,
		Parser: ParseCountryRecord,
	},
	{
		Source: DBSource{DBTypeASN, IPVersion4},
		URL:    ASNIPv4URL,
		Parser: ParseASNRecord,
	},
	{
		Source: DBSource{DBTypeASN, IPVersion6},
		URL:    ASNIPv6URL,
		Parser: ParseASNRecord,
	},
}

// Resolver is an IP resolver that returns information about an IP address.
type Resolver struct {
	db        atomic.Pointer[ResTree]
	collector DBUpdateCollector
	fetcher   Fetcher
}

// NewResolver creates a new IP resolver with the given metrics collector
// and fetcher.
func NewResolver(
	collector DBUpdateCollector,
	fetcher Fetcher,
) *Resolver {
	return &Resolver{
		collector: collector,
		fetcher:   fetcher,
	}
}

// Update updates the databases used by the resolver. The context can be used to cancel
// the update operation.
//
// If an error occurs while updating a database, the function proceeds to update the
// next database and returns all the errors at the end.
func (r *Resolver) Update(ctx context.Context) error {
	start := time.Now()
	db := itree.NewITree[netip.Addr, Resolution]()
	loader := NewLoader(r.fetcher)

	var errs []error
	entries := make(map[DBSource]uint64)
	for _, src := range defaultSources {
		count, err := loader.Load(ctx, db, src)
		if err != nil {
			errs = append(errs, err)
		}
		entries[src.Source] = count
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	// Combine identical intervals into single entries.
	r.db.Store(db.Compacted(mergeResolutions))
	r.collector.RecordDBUpdate(entries, time.Since(start))
	return nil
}

// Resolve resolves the given IP address to a country code and an ASN.
//
// It is the caller's responsibility to check if the IP is valid.
//
// If the country of the IP is not found, the CountryCode field of the result will be an
// empty string. If the ASN of the IP is not found, the ASN field of the result will be
// zero.
//
// The Organization field is present for informational purposes only. It is not used by
// the rules engine.
func (r *Resolver) Resolve(ip netip.Addr) Resolution {
	return mergeResolutions(r.db.Load().Query(ip))
}
