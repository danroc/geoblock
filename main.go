package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"net/http"
	"slices"
	"sort"

	"github.com/danroc/geoblock/set"
)

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

type InvalidIPError struct {
	Address string
}

func (e *InvalidIPError) Error() string {
	return fmt.Sprintf("invalid IP address: %s", e.Address)
}

type RangeData struct {
	CountryCode string `json:"country_code"`
}

type RangeEntry struct {
	StartIP net.IP    `json:"start_ip"`
	EndIP   net.IP    `json:"end_ip"`
	Data    RangeData `json:"data"`
}

const (
	countryIpV4Url = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-country/geolite2-country-ipv4.csv"
	countryIpV6Url = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-country/geolite2-country-ipv6.csv"
)

func CompareIP(a net.IP, b net.IP) int {
	return bytes.Compare(a, b)
}

func parseRecords(records [][]string) ([]RangeEntry, error) {
	var entries []RangeEntry
	for _, record := range records {
		var (
			startIP     = net.ParseIP(record[0])
			endIP       = net.ParseIP(record[1])
			countryCode = record[2]
		)

		if startIP == nil {
			return nil, &InvalidIPError{Address: record[0]}
		}

		if endIP == nil {
			return nil, &InvalidIPError{Address: record[1]}
		}

		entry := RangeEntry{
			StartIP: startIP,
			EndIP:   endIP,
			Data: RangeData{
				CountryCode: countryCode,
			},
		}

		entries = append(entries, entry)
	}

	// Sort the entries by the start IP so that we can use binary search
	slices.SortFunc(entries, func(a, b RangeEntry) int {
		return CompareIP(a.StartIP, b.StartIP)
	})

	return entries, nil
}

func findEntry(entries []RangeEntry, ip net.IP) *RangeEntry {
	// The IP string used to create the IP object is invalid
	if ip == nil {
		return nil
	}

	// Find the first entry whose start IP is greater than the given IP
	i := sort.Search(len(entries), func(i int) bool {
		return CompareIP(entries[i].StartIP, ip) > 0
	})

	// Not found: the IP is before the first entry
	if i == 0 {
		return nil
	}

	// The last entry whose start IP is less than or equal to the given IP
	match := entries[i-1]

	// From the search, it's guaranteed that the start IP of the match is less
	// than or equal to the given IP.
	if CompareIP(ip, match.EndIP) <= 0 {
		return &match
	}

	// Not found: the IP is NOT within the range
	return nil
}

const (
	HeaderXForwardedMethod = "X-Forwarded-Method"
	HeaderXForwardedProto  = "X-Forwarded-Proto"
	HeaderXForwardedHost   = "X-Forwarded-Host"
	HeaderXForwardedURI    = "X-Forwarded-Uri"
	HeaderXForwardedFor    = "X-Forwarded-For"
)

func getAuthorize(
	entries []RangeEntry,
	allowed set.Set[string],
	w http.ResponseWriter,
	r *http.Request,
) {
	origins := r.Header[HeaderXForwardedFor]

	// Block request: missing header
	if origins == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Find the country code for the client IP (first IP in the list)
	match := findEntry(entries, net.ParseIP(origins[0]))

	// Block request: IP not found
	if match == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Allow request: country code is in the allowed set
	if allowed.Contains(match.Data.CountryCode) {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Block request: default case
	w.WriteHeader(http.StatusForbidden)
}

func main() {
	records, err := fetchCsv(countryIpV4Url)
	if err != nil {
		fmt.Println(err)
		return
	}

	entries, err := parseRecords(records)
	if err != nil {
		fmt.Println(err)
		return
	}

	match := findEntry(entries, net.ParseIP("62.35.255.255"))
	fmt.Println(match)

	allowedCountryCodes := set.NewSet[string]()
	allowedCountryCodes.Add("FR")

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/authorize", func(w http.ResponseWriter, r *http.Request) {
		getAuthorize(entries, allowedCountryCodes, w, r)
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("Starting server at %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
