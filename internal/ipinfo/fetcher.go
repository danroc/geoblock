package ipinfo

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"
)

// URLs of the CSV IP location databases
const (
	CountryIPv4URL = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-country/geolite2-country-ipv4.csv"
	CountryIPv6URL = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-country/geolite2-country-ipv6.csv"
	ASNIPv4URL     = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-asn/geolite2-asn-ipv4.csv"
	ASNIPv6URL     = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-asn/geolite2-asn-ipv6.csv"
)

const (
	// The timeout for the HTTP client.
	clientTimeout = 60 * time.Second
)

// Fetcher fetches CSV records from a URL.
type Fetcher interface {
	Fetch(ctx context.Context, url string) ([][]string, error)
}

// HTTPFetcher is the default Fetcher implementation that fetches CSV records over HTTP.
type HTTPFetcher struct {
	Client *http.Client
}

// NewHTTPFetcher creates a new HTTPFetcher with a default HTTP client.
func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{
		Client: &http.Client{Timeout: clientTimeout},
	}
}

// Fetch fetches CSV records from the given URL.
func (f *HTTPFetcher) Fetch(ctx context.Context, url string) ([][]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, err
	}

	// Use an anonymous function to please the linter by not ignoring the error.
	defer func() { _ = resp.Body.Close() }()

	// We check the status code to avoid trying to parse an invalid response body.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return csv.NewReader(resp.Body).ReadAll()
}
