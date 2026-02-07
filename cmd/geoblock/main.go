// Package main contains the main geoblock application.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/danroc/geoblock/internal/config"
	"github.com/danroc/geoblock/internal/ipinfo"
	"github.com/danroc/geoblock/internal/metrics"
	"github.com/danroc/geoblock/internal/rules"
	"github.com/danroc/geoblock/internal/server"
	"github.com/danroc/geoblock/internal/version"
)

// RFC3339Milli is the RFC3339 format with milliseconds precision.
const RFC3339Milli = "2006-01-02T15:04:05.000Z07:00"

// Auto-reload, auto-update, and shutdown intervals
const (
	autoReloadInterval = 5 * time.Second
	autoUpdateInterval = 24 * time.Hour
	maxCacheAge        = 12 * time.Hour
	shutdownTimeout    = 30 * time.Second
)

// Log levels
const (
	LogLevelTrace = "trace"
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
	LogLevelPanic = "panic"
)

// Log formats
const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

// Default options
const (
	DefaultCacheDir   = "/var/cache/geoblock"
	DefaultConfigPath = "/etc/geoblock/config.yaml"
	DefaultLogFormat  = LogFormatJSON
	DefaultLogLevel   = LogLevelInfo
	DefaultServerPort = "8080"
)

// Environment variable names
const (
	OptionCacheDir   = "GEOBLOCK_CACHE_DIR"
	OptionConfigPath = "GEOBLOCK_CONFIG_FILE"
	OptionLogFormat  = "GEOBLOCK_LOG_FORMAT"
	OptionLogLevel   = "GEOBLOCK_LOG_LEVEL"
	OptionServerPort = "GEOBLOCK_PORT"
)

// getEnv retrieves the value of the environment variable key. If it is not set, it
// returns the fallback value.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

type appOptions struct {
	configPath string
	serverPort string
	logLevel   string
	logFormat  string
}

// getOptions returns the application options from the environment variables.
func getOptions() *appOptions {
	return &appOptions{
		configPath: getEnv(OptionConfigPath, DefaultConfigPath),
		serverPort: getEnv(OptionServerPort, DefaultServerPort),
		logLevel:   getEnv(OptionLogLevel, DefaultLogLevel),
		logFormat:  getEnv(OptionLogFormat, DefaultLogFormat),
	}
}

// getCacheDir returns the database cache directory. Set env var to empty string to
// disable caching.
func getCacheDir() string {
	if val, ok := os.LookupEnv(OptionCacheDir); ok {
		return val
	}
	return DefaultCacheDir
}

// cacheLogger implements ipinfo.CacheLogger using zerolog.
type cacheLogger struct{}

func (cacheLogger) Warn(msg, path string, err error) {
	log.Warn().Err(err).Str("path", path).Msg(msg)
}

// runEvery executes fn at the given interval until the context is canceled.
func runEvery(ctx context.Context, interval time.Duration, fn func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn()
		}
	}
}

// Updater is the interface for types that can update their databases.
type Updater interface {
	Update(ctx context.Context) error
}

// autoUpdate updates the databases at regular intervals.
func autoUpdate(ctx context.Context, resolver Updater) {
	runEvery(ctx, autoUpdateInterval, func() {
		if err := resolver.Update(ctx); err != nil {
			log.Error().Err(err).Msg("Cannot update databases")
		} else {
			log.Info().Msg("Databases updated")
		}
	})
}

// loadConfig reads the configuration file from the given path and returns it.
func loadConfig(path string) (*config.Configuration, error) {
	file, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return nil, err
	}
	return config.ReadConfig(bytes.NewReader(file))
}

// ConfigUpdater is the interface for types that can update their configuration.
type ConfigUpdater interface {
	UpdateConfig(config *config.AccessControl)
}

// configReloader watches a config file for changes.
type configReloader struct {
	path     string
	prevStat os.FileInfo
	// Swappable for testing: stat retrieves file metadata for the config file, and load
	// parses and returns the configuration from the given path.
	stat func(string) (os.FileInfo, error)
	load func(string) (*config.Configuration, error)
}

// newConfigReloader creates a config reloader for the given path.
func newConfigReloader(path string) (*configReloader, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &configReloader{
		path:     path,
		prevStat: stat,
		stat:     os.Stat,
		load:     loadConfig,
	}, nil
}

// hasChanged returns true if the file metadata differs from the previous stat.
func (r *configReloader) hasChanged(stat os.FileInfo) bool {
	var (
		sizeChanged    = r.prevStat.Size() != stat.Size()
		modTimeChanged = !r.prevStat.ModTime().Equal(stat.ModTime())
	)
	return sizeChanged || modTimeChanged
}

// reloadIfChanged checks if the config file changed and updates the engine if so.
// Returns (reloaded, rulesCount, error).
func (r *configReloader) reloadIfChanged(engine ConfigUpdater) (bool, int, error) {
	stat, err := r.stat(r.path)
	if err != nil {
		return false, 0, fmt.Errorf("stat config file: %w", err)
	}

	if !r.hasChanged(stat) {
		return false, 0, nil
	}

	cfg, err := r.load(r.path)
	if err != nil {
		return false, 0, fmt.Errorf("load config file: %w", err)
	}

	engine.UpdateConfig(&cfg.AccessControl)

	// Update prevStat only after successful reload
	r.prevStat = stat
	return true, len(cfg.AccessControl.Rules), nil
}

// ConfigReloadCollector collects metrics for config reloads.
type ConfigReloadCollector interface {
	RecordConfigReload(success bool, rulesCount int)
}

// autoReload watches the configuration file for changes and updates the engine when it
// happens. It stops when the context is canceled.
func autoReload(
	ctx context.Context,
	engine ConfigUpdater,
	path string,
	collector ConfigReloadCollector,
) {
	reloader, err := newConfigReloader(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("Cannot watch configuration file")
		return
	}

	runEvery(ctx, autoReloadInterval, func() {
		reloaded, rulesCount, err := reloader.reloadIfChanged(engine)
		if err != nil {
			log.Error().Err(err).Str("path", path).Msg("Cannot reload configuration")
			collector.RecordConfigReload(false, 0)
			return
		}
		if reloaded {
			log.Info().Msg("Configuration reloaded")
			collector.RecordConfigReload(true, rulesCount)
		}
	})
}

// Shutdowner is the interface for types that can be shut down.
type Shutdowner interface {
	Shutdown(context.Context) error
}

// stopServer waits for the context to be canceled and then shuts down the server.
func stopServer(ctx context.Context, srv Shutdowner) {
	<-ctx.Done()
	log.Info().Msg("Shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server shutdown error")
	}
}

// parseLogLevel parses the log level from string to zerolog.Level. It defaults to info
// level if the provided level is invalid.
func parseLogLevel(level string) (zerolog.Level, error) {
	switch level {
	case LogLevelTrace:
		return zerolog.TraceLevel, nil
	case LogLevelDebug:
		return zerolog.DebugLevel, nil
	case LogLevelInfo:
		return zerolog.InfoLevel, nil
	case LogLevelWarn:
		return zerolog.WarnLevel, nil
	case LogLevelError:
		return zerolog.ErrorLevel, nil
	case LogLevelFatal:
		return zerolog.FatalLevel, nil
	case LogLevelPanic:
		return zerolog.PanicLevel, nil
	default:
		return zerolog.InfoLevel, errors.New("invalid log level")
	}
}

// configureLogger configures the logger with the given log format and level.
func configureLogger(logFormat, level string) {
	// Configure log format before emitting any log messages.
	switch logFormat {
	case LogFormatJSON:
		zerolog.TimeFieldFormat = RFC3339Milli
	case LogFormatText:
		log.Logger = log.Output(
			zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
		)
	default:
		log.Logger = log.Output(
			zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
		)
		log.Warn().Str("format", logFormat).Msg("Invalid log format")
	}

	// If the log level is invalid, default to info level.
	if parsedLevel, err := parseLogLevel(level); err != nil {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Warn().Str("level", level).Msg("Invalid log level")
	} else {
		zerolog.SetGlobalLevel(parsedLevel)
	}
}

func main() {
	options := getOptions()
	configureLogger(options.logFormat, options.logLevel)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	log.Info().Str("version", version.Get()).Msg("Starting Geoblock")
	log.Info().Msg("Loading configuration file")
	cfg, err := loadConfig(options.configPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", options.configPath).Msg(
			"Cannot read configuration file",
		)
	}

	collector := metrics.NewCollector()
	collector.RecordConfigReload(true, len(cfg.AccessControl.Rules))

	log.Info().Msg("Initializing database")
	resolver := ipinfo.NewResolver(
		collector,
		ipinfo.NewCachedFetcher(
			getCacheDir(),
			maxCacheAge,
			ipinfo.NewHTTPFetcher(),
			cacheLogger{},
		),
	)
	if err := resolver.Update(ctx); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize database")
	}

	var (
		address = ":" + options.serverPort
		engine  = rules.NewEngine(&cfg.AccessControl)
		srv     = server.New(address, engine, resolver, collector, metrics.Handler())
	)

	// Start background tasks to update databases, reload configuration, and gracefully
	// shut down the server.
	go autoUpdate(ctx, resolver)
	go autoReload(ctx, engine, options.configPath, collector)
	go stopServer(ctx, srv)

	log.Info().Str("address", srv.Addr).Msg("Starting server")

	// ErrServerClosed is expected on graceful shutdown, so only log fatal for
	// unexpected errors.
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("Server stopped")
	}
}
