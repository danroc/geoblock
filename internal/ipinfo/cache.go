package ipinfo

import (
	"context"
	"encoding/csv"
	"os"
	"path"
	"path/filepath"
	"time"
)

// defaultDirPermissions is the default permissions used to create cache directories.
const defaultDirPermissions = 0o750

// CacheLogger is the interface for logging cache operations.
type CacheLogger interface {
	Warn(msg, path string, err error)
}

// CachedFetcher is a Fetcher that caches fetched CSV records in a local directory. It
// checks the cache before fetching, and updates the cache after fetching.
//
// The cache entries are considered valid for a specified maximum age.
type CachedFetcher struct {
	CacheDir string
	MaxAge   time.Duration
	Fetcher  Fetcher
	Logger   CacheLogger
}

// NewCachedFetcher creates a new CachedFetcher with the given cache directory, maximum
// age for cache entries, underlying fetcher, and logger.
func NewCachedFetcher(
	cacheDir string,
	maxAge time.Duration,
	fetcher Fetcher,
	logger CacheLogger,
) *CachedFetcher {
	return &CachedFetcher{
		CacheDir: cacheDir,
		MaxAge:   maxAge,
		Fetcher:  fetcher,
		Logger:   logger,
	}
}

// Fetch fetches CSV records from the given URL, using the cache if possible.
func (c *CachedFetcher) Fetch(ctx context.Context, url string) ([][]string, error) {
	// If caching is disabled, just use the underlying fetcher directly.
	if c.CacheDir == "" {
		return c.Fetcher.Fetch(ctx, url)
	}

	// The cache file is named after the base name of the URL, and stored in the cache
	// directory.
	cachePath := filepath.Join(c.CacheDir, path.Base(url))

	// Check if the cache file exists and is still valid. If so, read from the cache
	// instead of fetching.
	if info, err := os.Stat(cachePath); err == nil {
		if time.Since(info.ModTime()) < c.MaxAge {
			records, err := readCSV(cachePath)
			if err == nil {
				return records, nil
			}
			// Cache read failed, log warning and fall through to fetch fresh data.
			c.Logger.Warn("Failed to read cache file", cachePath, err)
		}
	}

	// Otherwise, use the underlying fetcher to fetch the data.
	records, err := c.Fetcher.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	// Try to write the fetched data to the cache for future use. If this fails, we
	// log a warning and return the fetched data anyway.
	if err := writeCSV(cachePath, records); err != nil {
		c.Logger.Warn("Failed to write cache file", cachePath, err)
	}
	return records, nil
}

// readCSV reads a CSV file from the given path and returns the records.
func readCSV(path string) ([][]string, error) {
	file, err := os.Open(path) // #nosec G304 -- Path is cache dir + remote filename
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	return csv.NewReader(file).ReadAll()
}

// writeCSV writes the given records to a CSV file at the given path. It tries to create
// the parent directories if they do not exist.
func writeCSV(path string, records [][]string) error {
	dir := filepath.Dir(path)

	// Try to create the parent directories if they do not exist.
	if err := os.MkdirAll(dir, defaultDirPermissions); err != nil {
		return err
	}

	// Create a temporary file in the same directory, this will be used to write the
	// cache data before renaming it to the final path.
	//
	// This ensures that we don't end up with a partially written cache file if the
	// program is interrupted while writing.
	tmpFile, err := os.CreateTemp(dir, ".cache-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Write the records to the temporary file.
	writer := csv.NewWriter(tmpFile)
	if err := writer.WriteAll(records); err != nil {
		_ = tmpFile.Close()
		return err
	}

	// Close and rename the temporary file to the final path.
	if err := tmpFile.Close(); err != nil {
		return err
	}
	return os.Rename(tmpFile.Name(), path)
}
