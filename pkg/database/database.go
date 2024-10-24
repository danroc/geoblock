package database

import (
	"encoding/csv"
	"net"
	"net/http"
	"slices"
	"sort"
	"strings"

	"github.com/danroc/geoblock/pkg/utils"
)

// Entry represents an IP range and its associated data.
type Entry struct {
	StartIP net.IP
	EndIP   net.IP
	Data    []string
}

// fetchCsv fetches a CSV file from the given URL and returns its records.
func fetchCsv(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

// sanatizeData trims the leading and trailing spaces from the given strings.
func sanatizeData(data []string) []string {
	var sanitized []string
	for _, s := range data {
		sanitized = append(sanitized, strings.TrimSpace(s))
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
			data    = sanatizeData(record[2:])
		)

		if startIP == nil {
			return nil, &utils.InvalidIPError{Address: record[0]}
		}

		if endIP == nil {
			return nil, &utils.InvalidIPError{Address: record[1]}
		}

		entries = append(entries, Entry{
			StartIP: startIP,
			EndIP:   endIP,
			Data:    data,
		})
	}
	return entries, nil
}

// sortEntries sorts the entries by their start IP.
func sortEntries(entries []Entry) {
	slices.SortFunc(entries, func(a, b Entry) int {
		return utils.CompareIP(a.StartIP, b.StartIP)
	})
}

// Database represents a database of IP ranges.
type Database struct {
	entries []Entry
}

// NewDatabase creates a new database from the given URL.
func NewDatabase(url string) (*Database, error) {
	// Records are the raw data from the CSV file.
	records, err := fetchCsv(url)
	if err != nil {
		return nil, err
	}

	// Entries are the parsed data from the records, it is composed by a start
	// IP, an end IP, and the string data associated with the range.
	entries, err := parseRecords(records)
	if err != nil {
		return nil, err
	}

	// The entries must be sorted by their start IP to allow binary search. The
	// sort is done in-place.
	sortEntries(entries)
	return &Database{entries: entries}, nil
}

// Find returns the data associated with the entry that contains the given IP.
// If the IP is not found, nil is returned.
func (db *Database) Find(ip net.IP) []string {
	// If the given IP address is invalid, we return nil to indidate that the
	// IP cannot be found in the database. It is up to the caller to validate
	// the IP address before calling this method.
	if ip == nil {
		return nil
	}

	// Find the first entry whose start-IP is greater than the given IP. The
	// search cannot be done the other way around (i.e., search for the first
	// entry whose start-IP is less than or equal to the given IP) because it
	// would return the first entry in most of the cases.
	i := sort.Search(len(db.entries), func(i int) bool {
		return utils.CompareIP(db.entries[i].StartIP, ip) > 0
	})

	// Not found: the start-IP of the first entry is greater than the given IP.
	if i == 0 {
		return nil
	}

	// The last entry whose start-IP is less than or equal to the given IP.
	match := db.entries[i-1]

	// From the search, it's guaranteed that the start-IP of the match is less
	// than or equal to the given IP. So, the IP only needs to be compared to
	// the end-IP of the match.
	if utils.CompareIP(ip, match.EndIP) <= 0 {
		return match.Data
	}

	// Not found: the IP is NOT within the range
	return nil
}
