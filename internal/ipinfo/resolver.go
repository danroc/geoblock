// Package ipinfo provides an IP resolver that returns information about an IP address.
package ipinfo

import (
	"encoding/csv"
	"errors"
	"net/http"
	"net/netip"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/danroc/geoblock/internal/itree"
)

// URLs of the CSV IP location databases
const (
	CountryIPv4URL = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-country/geolite2-country-ipv4.csv"
	CountryIPv6URL = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-country/geolite2-country-ipv6.csv"
	ASNIPv4URL     = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-asn/geolite2-asn-ipv4.csv"
	ASNIPv6URL     = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-asn/geolite2-asn-ipv6.csv"
)

// Length of the CSV records (number of fields)
const (
	asnRecordLength     = 4
	countryRecordLength = 3
)

const (
	// The timeout for the HTTP client.
	clientTimeout = 30 * time.Second
)

// ErrRecordLength is returned when a CSV record has an unexpected length.
var (
	ErrRecordLength = errors.New("invalid record length")
	ErrInvalidASN   = errors.New("invalid ASN")
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

// Resolver is an IP resolver that returns information about an IP address.
type Resolver struct {
	db atomic.Pointer[ResTree]
}

// NewResolver creates a new IP resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// Update updates the databases used by the resolver.
//
// If an error occurs while updating a database, the function proceeds to update the
// next database and returns all the errors at the end.
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

	// A new database is created for each update so that it can be atomically swapped
	// with the current database.
	db := itree.NewITree[netip.Addr, Resolution]()

	var errs []error
	for _, item := range items {
		if err := update(db, item.parser, item.url); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	// Atomically swap the current database with the new one.
	r.db.Store(db)
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

// update adds the records fetched from the given URL to the database.
func update(db *ResTree, parser ParserFn, url string) error {
	records, err := fetchCSV(url)
	if err != nil {
		return err
	}

	var errs []error
	for _, record := range records {
		entry, err := parser(record)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		db.Insert(
			itree.NewInterval(entry.StartIP, entry.EndIP),
			entry.Resolution,
		)
	}
	return errors.Join(errs...)
}

// fetchCSV returns the CSV records fetched from the given URL.
func fetchCSV(url string) ([][]string, error) {
	// It's important to set a timeout to avoid hanging the program if the remote server
	// doesn't respond.
	client := &http.Client{
		Timeout: clientTimeout,
	}

	resp, err := client.Get(url) // #nosec G107
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return csv.NewReader(resp.Body).ReadAll()
}

// parseIPRange parses the start and end IP addresses from a record.
func parseIPRange(record []string) (netip.Addr, netip.Addr, error) {
	startIP, err := netip.ParseAddr(record[0])
	if err != nil {
		return netip.Addr{}, netip.Addr{}, err
	}

	endIP, err := netip.ParseAddr(record[1])
	if err != nil {
		return netip.Addr{}, netip.Addr{}, err
	}

	return startIP, endIP, nil
}

// parseCountryRecord parses a country database record.
func parseCountryRecord(record []string) (*DBRecord, error) {
	if len(record) != countryRecordLength {
		return nil, ErrRecordLength
	}

	startIP, endIP, err := parseIPRange(record)
	if err != nil {
		return nil, err
	}

	return &DBRecord{
		StartIP: startIP,
		EndIP:   endIP,
		Resolution: Resolution{
			CountryCode: record[2],
		},
	}, nil
}

// parseASNRecord parses an ASN database record.
func parseASNRecord(record []string) (*DBRecord, error) {
	if len(record) != asnRecordLength {
		return nil, ErrRecordLength
	}

	startIP, endIP, err := parseIPRange(record)
	if err != nil {
		return nil, err
	}

	asn, err := strconv.ParseUint(record[2], 10, 32)
	if err != nil {
		return nil, ErrInvalidASN
	}

	return &DBRecord{
		StartIP: startIP,
		EndIP:   endIP,
		Resolution: Resolution{
			ASN:          uint32(asn),
			Organization: record[3],
		},
	}, nil
}
