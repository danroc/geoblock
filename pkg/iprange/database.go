// Package iprange provides a database of IP ranges and their associated data.
package iprange

import (
	"encoding/csv"
	"io"
	"net"
	"slices"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/danroc/geoblock/pkg/utils/iputils"
)

// Entry represents an IP range and its associated data.
type Entry struct {
	StartIP net.IP
	EndIP   net.IP
	Data    []string
}

// sanitize trims the leading and trailing spaces from the given strings.
func sanitize(data []string) []string {
	sanitized := make([]string, len(data))
	for i, d := range data {
		sanitized[i] = strings.TrimSpace(d)
	}
	return sanitized
}

// parseRecords parses the given CSV records into database entries.
func parseRecords(records [][]string) ([]Entry, error) {
	var entries []Entry
	for _, record := range records {
		var (
			startIP = net.ParseIP(record[0])
			endIP   = net.ParseIP(record[1])
			data    = sanitize(record[2:])
		)

		if startIP == nil {
			return nil, &iputils.ErrInvalidIP{Address: record[0]}
		}

		if endIP == nil {
			return nil, &iputils.ErrInvalidIP{Address: record[1]}
		}

		entries = append(entries, Entry{
			StartIP: startIP,
			EndIP:   endIP,
			Data:    data,
		})
	}
	return entries, nil
}

// Database represents a database of IP ranges.
type Database struct {
	entries atomic.Value // []Entry
}

// NewDatabase creates a new database from the given URL.
func NewDatabase() *Database {
	db := &Database{}
	db.entries.Store([]Entry{})
	return db
}

// Update updates the database with the data from the given reader.
func (db *Database) Update(reader io.Reader) error {
	// Records are the raw data from the CSV file.
	records, err := csv.NewReader(reader).ReadAll()
	if err != nil {
		return err
	}

	// Entries are the parsed data from the records, it is composed by a start
	// IP, an end IP, and the string data associated with the range.
	entries, err := parseRecords(records)
	if err != nil {
		return err
	}

	// The entries must be sorted by their start IP to allow binary search. The
	// sort is done in-place.
	slices.SortFunc(entries, func(a, b Entry) int {
		return iputils.CompareIP(a.StartIP, b.StartIP)
	})

	// This atomically updates the database entries.
	db.entries.Store(entries)

	return nil
}

// Find returns the data associated with the entry that contains the given IP.
// If the IP is not found, nil is returned.
func (db *Database) Find(ip net.IP) []string {
	// If the given IP address is invalid, we return nil to indicate that the
	// IP cannot be found in the database. It is up to the caller to validate
	// the IP address before calling this method.
	if ip == nil {
		return nil
	}

	// Atomically load the database entries.
	entries := db.entries.Load().([]Entry)

	// Find the first entry whose start-IP is greater than the given IP. The
	// search cannot be done the other way around (i.e., search for the first
	// entry whose start-IP is less than or equal to the given IP) because it
	// would return the first entry of the list in most of the cases.
	i := sort.Search(len(entries), func(i int) bool {
		return iputils.CompareIP(entries[i].StartIP, ip) > 0
	})

	// Not found: the start-IP of the first entry is greater than the given IP.
	if i == 0 {
		return nil
	}

	// The last entry whose start-IP is less than or equal to the given IP.
	match := entries[i-1]

	// From the search, it's guaranteed that the start-IP of the match is less
	// than or equal to the given IP. So, the IP only needs to be compared to
	// the end-IP of the match.
	if iputils.CompareIP(ip, match.EndIP) <= 0 {
		return match.Data
	}

	// Not found: the IP is NOT within the range.
	return nil
}
