package ipinfo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danroc/geoblock/internal/ipinfo"
)

func TestHTTPFetcher_Success(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("a,b\nc,d\n"))
		}),
	)
	defer srv.Close()

	fetcher := ipinfo.NewHTTPFetcher()
	records, err := fetcher.Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Fetch() error = %v, want nil", err)
	}
	if len(records) != 2 {
		t.Errorf("Fetch() returned %d records, want 2", len(records))
	}
}

func TestHTTPFetcher_InvalidURL(t *testing.T) {
	fetcher := ipinfo.NewHTTPFetcher()
	_, err := fetcher.Fetch(context.Background(), "http://example.com/\x00invalid")
	if err == nil {
		t.Error("Fetch() error = nil, want error")
	}
}

func TestHTTPFetcher_RequestError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fetcher := ipinfo.NewHTTPFetcher()
	_, err := fetcher.Fetch(ctx, "http://example.com")
	if err == nil {
		t.Error("Fetch() error = nil, want error")
	}
}

func TestHTTPFetcher_Non200Status(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)
	defer srv.Close()

	fetcher := ipinfo.NewHTTPFetcher()
	_, err := fetcher.Fetch(context.Background(), srv.URL)
	if err == nil || !strings.Contains(err.Error(), "unexpected status") {
		t.Errorf("Fetch() error = %v, want error containing 'unexpected status'", err)
	}
}
