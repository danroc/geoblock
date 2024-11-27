// Package ipres provider an IP resolver that returns information about an IP
// address.
package ipres

import (
	"encoding/csv"
	"errors"
	"net/http"
	"net/netip"
	"strconv"

	"github.com/danroc/geoblock/internal/itree"
)

// URLs of the CSV IP location databases.
const (
	CountryIPv4URL = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-country/geolite2-country-ipv4.csv"
	CountryIPv6URL = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-country/geolite2-country-ipv6.csv"
	ASNIPv4URL     = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-asn/geolite2-asn-ipv4.csv"
	ASNIPv6URL     = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-asn/geolite2-asn-ipv6.csv"
)

// AS0 represents the default ASN value for unknown addresses.
const AS0 uint32 = 0

// DBRecord contains the information of a database record.
type DBRecord struct {
	StartIP    netip.Addr
	EndIP      netip.Addr
	Resolution Resolution
}

// ParserFn is a function that parses a CSV record into a database record.
type ParserFn func([]string) (*DBRecord, error)

// Resolution contains the result of resolving an IP address.
type Resolution struct {
	CountryCode  string
	Organization string
	ASN          uint32
}

// Or returns a new resolution that combines the fields of the receiver and the
// other resolution.
func (r *Resolution) Or(other Resolution) Resolution {
	return Resolution{
		CountryCode:  ifEmpty(r.CountryCode, other.CountryCode),
		Organization: ifEmpty(r.Organization, other.Organization),
		ASN:          ifZero(r.ASN, other.ASN),
	}
}

func ifEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func ifZero(value, fallback uint32) uint32 {
	if value == 0 {
		return fallback
	}
	return value
}

// Resolver is an IP resolver that returns information about an IP address.
type Resolver struct {
	db *itree.ITree[netip.Addr, Resolution]
}

// NewResolver creates a new IP resolver.
func NewResolver() *Resolver {
	return &Resolver{
		db: itree.NewITree[netip.Addr, Resolution](),
	}
}

// Update updates the databases used by the resolver.
//
// If an error occurs while updating a database, the function proceeds to
// update the next database and returns all the errors at the end.
func (r *Resolver) Update() error {
	items := []struct {
		parser ParserFn
		url    string
	}{
		{parseCountryRecord, CountryIPv4URL},
		{parseCountryRecord, CountryIPv6URL},
		{parseASNRecord, ASNIPv4URL},
		{parseASNRecord, ASNIPv6URL},
	}

	var errs []error

	for _, item := range items {
		if err := r.updateDB(item.parser, item.url); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Resolve resolves the given IP address to a country code and an ASN.
//
// It is the caller's responsibility to check if the IP is valid.
//
// If the country of the IP is not found, the CountryCode field of the result
// will be an empty string. If the ASN of the IP is not found, the ASN field of
// the result will be zero.
//
// The Organization field is present for informational purposes only. It is not
// used by the rules engine.
func (r *Resolver) Resolve(ip netip.Addr) Resolution {
	results := r.db.Query(ip)

	return reduce(results, func(acc, r Resolution) Resolution {
		return acc.Or(r)
	}, Resolution{})
}

// updateDB updates the given database with the data from the given URL.
func (r *Resolver) updateDB(parser ParserFn, url string) error {
	records, err := fetchCSV(url)
	if err != nil {
		return err
	}

	for _, record := range records {
		entry, err := parser(record)
		if err != nil {
			return err
		}
		r.db.Insert(
			itree.NewInterval(entry.StartIP, entry.EndIP),
			entry.Resolution,
		)
	}
	return nil
}

func fetchCSV(url string) ([][]string, error) {
	resp, err := http.Get(url) // #nosec G107
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return csv.NewReader(resp.Body).ReadAll()
}

func parseCountryRecord(record []string) (*DBRecord, error) {
	startIP, err := netip.ParseAddr(record[0])
	if err != nil {
		return nil, err
	}

	endIP, err := netip.ParseAddr(record[1])
	if err != nil {
		return nil, err
	}

	return &DBRecord{
		StartIP: startIP,
		EndIP:   endIP,
		Resolution: Resolution{
			CountryCode: strIndex(record, 2),
		},
	}, nil
}

func parseASNRecord(record []string) (*DBRecord, error) {
	startIP, err := netip.ParseAddr(record[0])
	if err != nil {
		return nil, err
	}

	endIP, err := netip.ParseAddr(record[1])
	if err != nil {
		return nil, err
	}

	return &DBRecord{
		StartIP: startIP,
		EndIP:   endIP,
		Resolution: Resolution{
			ASN:          strToASN(strIndex(record, 2)),
			Organization: strIndex(record, 3),
		},
	}, nil
}

// strToASN converts a string to an ASN. If the string is not a valid ASN, the
// function returns ReservedAS0.
func strToASN(s string) uint32 {
	asn, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return AS0
	}
	return uint32(asn)
}

// strIndex returns the element at the given index of the data slice. If the
// index is out of bounds, the function returns an empty string.
func strIndex(data []string, index int) string {
	if index < 0 || index >= len(data) {
		return ""
	}
	return data[index]
}

func reduce[T any, R any](
	collection []T,
	accumulator func(R, T) R,
	initial R,
) R {
	for _, item := range collection {
		initial = accumulator(initial, item)
	}
	return initial
}
