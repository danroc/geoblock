package ipinfo_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/danroc/geoblock/internal/ipinfo"
)

// nopLogger is a no-op implementation of CacheLogger for testing.
type nopLogger struct{}

func (nopLogger) Warn(string, string, error) {}

// mockFetcher is a test double for the Fetcher interface.
type mockFetcher struct {
	records [][]string
	err     error
	calls   int
}

func (m *mockFetcher) Fetch(_ context.Context, _ string) ([][]string, error) {
	m.calls++
	return m.records, m.err
}

func writeCache(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write cache file: %v", err)
	}
}

func setModTime(t *testing.T, path string, modTime time.Time) {
	t.Helper()
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("failed to set mod time: %v", err)
	}
}

func TestCachedFetcher_Fetch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cacheDir    func(t *testing.T) string
		maxAge      time.Duration
		records     [][]string
		fetchErr    error
		setupCache  func(t *testing.T, cacheDir string)
		wantCalls   int
		wantRecords [][]string
		wantErr     bool
	}{
		{
			name:        "bypasses cache when dir empty",
			cacheDir:    func(*testing.T) string { return "" },
			maxAge:      time.Hour,
			records:     [][]string{{"fetched", "data"}},
			wantCalls:   1,
			wantRecords: [][]string{{"fetched", "data"}},
		},
		{
			name:   "uses valid cache",
			maxAge: time.Hour,
			setupCache: func(t *testing.T, cacheDir string) {
				writeCache(t, filepath.Join(cacheDir, "data.csv"), "cached,data\n")
			},
			wantCalls:   0,
			wantRecords: [][]string{{"cached", "data"}},
		},
		{
			name:    "fetches when cache expired",
			maxAge:  time.Hour,
			records: [][]string{{"fresh", "data"}},
			setupCache: func(t *testing.T, cacheDir string) {
				path := filepath.Join(cacheDir, "data.csv")
				writeCache(t, path, "old,data\n")
				setModTime(t, path, time.Now().Add(-2*time.Hour))
			},
			wantCalls:   1,
			wantRecords: [][]string{{"fresh", "data"}},
		},
		{
			name:      "returns error from underlying fetcher",
			maxAge:    time.Hour,
			fetchErr:  errors.New("network error"),
			wantCalls: 1,
			wantErr:   true,
		},
		{
			name:    "falls back to fetch when cache corrupted",
			maxAge:  time.Hour,
			records: [][]string{{"fresh", "data"}},
			setupCache: func(t *testing.T, cacheDir string) {
				// Write malformed CSV (unclosed quote)
				writeCache(t, filepath.Join(cacheDir, "data.csv"), "\"unclosed\n")
			},
			wantCalls:   1,
			wantRecords: [][]string{{"fresh", "data"}},
		},
		{
			name: "creates cache dir when missing",
			cacheDir: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "sub")
			},
			maxAge:      time.Hour,
			records:     [][]string{{"a", "b"}},
			wantCalls:   1,
			wantRecords: [][]string{{"a", "b"}},
		},
		{
			name: "cache write failure does not affect return",
			cacheDir: func(t *testing.T) string {
				dir := t.TempDir()
				err := os.Chmod(dir, 0o500) // #nosec G302
				if err != nil {
					t.Fatalf("failed to chmod: %v", err)
				}
				t.Cleanup(func() {
					_ = os.Chmod(dir, 0o700) // #nosec G302
				})
				return dir
			},
			maxAge:      time.Hour,
			records:     [][]string{{"a", "b"}},
			wantCalls:   1,
			wantRecords: [][]string{{"a", "b"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cacheDir := t.TempDir()
			if tt.cacheDir != nil {
				cacheDir = tt.cacheDir(t)
			}

			if tt.setupCache != nil {
				tt.setupCache(t, cacheDir)
			}

			mock := &mockFetcher{records: tt.records, err: tt.fetchErr}
			cached := ipinfo.NewCachedFetcher(cacheDir, tt.maxAge, mock, nopLogger{})

			got, err := cached.Fetch(
				context.Background(),
				"http://example.com/data.csv",
			)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Fetch() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Fetch() error = %v, want nil", err)
			}
			if mock.calls != tt.wantCalls {
				t.Errorf("fetcher calls = %d, want %d", mock.calls, tt.wantCalls)
			}
			if !reflect.DeepEqual(got, tt.wantRecords) {
				t.Errorf("got %v, want %v", got, tt.wantRecords)
			}
		})
	}
}

func TestCachedFetcher_Fetch_CachePersistence(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	wantRecords := [][]string{{"data", "here"}}
	mock := &mockFetcher{records: wantRecords}
	cached := ipinfo.NewCachedFetcher(cacheDir, time.Hour, mock, nopLogger{})

	// First call fetches, second uses cache
	for i := range 2 {
		got, err := cached.Fetch(context.Background(), "http://example.com/data.csv")
		if err != nil {
			t.Fatalf("call %d: error = %v", i, err)
		}
		if !reflect.DeepEqual(got, wantRecords) {
			t.Errorf("call %d: got %v, want %v", i, got, wantRecords)
		}
	}

	if mock.calls != 1 {
		t.Errorf("fetcher calls = %d, want 1", mock.calls)
	}
}
