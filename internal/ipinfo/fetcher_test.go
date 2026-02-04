package ipinfo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danroc/geoblock/internal/ipinfo"
)

func TestHTTPFetcher_InvalidURL(t *testing.T) {
	fetcher := &ipinfo.HTTPFetcher{Client: http.DefaultClient}
	url := "http://example.com/\x00invalid"
	_, err := fetcher.Fetch(context.Background(), url)
	if err == nil {
		t.Errorf("HTTPFetcher.Fetch(%q) error = nil, want error", url)
	}
}

func TestHTTPFetcher_Non200Status(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)
	defer srv.Close()

	fetcher := &ipinfo.HTTPFetcher{Client: srv.Client()}
	_, err := fetcher.Fetch(context.Background(), srv.URL)
	if err == nil || !strings.Contains(err.Error(), "unexpected status") {
		t.Errorf(
			"HTTPFetcher.Fetch() error = %v, want error containing 'unexpected status'",
			err,
		)
	}
}
