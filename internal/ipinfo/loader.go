package ipinfo

import (
	"context"
	"errors"
	"net/netip"
	"strconv"

	"github.com/danroc/geoblock/internal/itree"
)

// Length of the CSV records (number of fields)
const (
	asnRecordLength     = 4
	countryRecordLength = 3
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

// ParserFunc is a function that parses a CSV record into a database record.
type ParserFunc func([]string) (*DBRecord, error)

// Loader loads database records from a source into an interval tree.
type Loader struct {
	fetcher Fetcher
}

// NewLoader creates a new Loader with the given fetcher.
func NewLoader(fetcher Fetcher) *Loader {
	return &Loader{fetcher: fetcher}
}

// Load fetches records from the source and inserts them into the database.
func (l *Loader) Load(
	ctx context.Context,
	db *ResTree,
	src DBSourceSpec,
) (uint64, error) {
	records, err := l.fetcher.Fetch(ctx, src.URL)
	if err != nil {
		return 0, err
	}

	var (
		count uint64
		errs  []error
	)
	for _, rec := range records {
		entry, err := src.Parser(rec)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		db.Insert(
			itree.NewInterval(entry.StartIP, entry.EndIP),
			entry.Resolution,
		)
		count++
	}
	return count, errors.Join(errs...)
}

// parseIPRange parses the start and end IP addresses from a record. Callers must ensure
// the record has at least 2 elements.
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

// ParseCountryRecord parses a country database record.
func ParseCountryRecord(record []string) (*DBRecord, error) {
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

// ParseASNRecord parses an ASN database record.
func ParseASNRecord(record []string) (*DBRecord, error) {
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
