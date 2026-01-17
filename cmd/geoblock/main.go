// Package main contains the main geoblock application.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/danroc/geoblock/internal/config"
	"github.com/danroc/geoblock/internal/ipinfo"
	"github.com/danroc/geoblock/internal/rules"
	"github.com/danroc/geoblock/internal/server"
	"github.com/danroc/geoblock/internal/version"
)

// RFC3339Milli is the RFC3339 format with milliseconds precision.
const RFC3339Milli = "2006-01-02T15:04:05.000Z07:00"

// Auto-update and auto-reload intervals.
const (
	autoUpdateInterval = 24 * time.Hour
	autoReloadInterval = 5 * time.Second
)

// Log levels.
const (
	LogLevelTrace = "trace"
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
	LogLevelPanic = "panic"
)

// Log formats.
const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

// Default options.
const (
	DefaultConfigPath = "/etc/geoblock/config.yaml"
	DefaultServerPort = "8080"
	DefaultLogLevel   = LogLevelInfo
	DefaultLogFormat  = LogFormatJSON
)

// Environment variable names.
const (
	OptionConfigPath = "GEOBLOCK_CONFIG_FILE"
	OptionServerPort = "GEOBLOCK_PORT"
	OptionLogLevel   = "GEOBLOCK_LOG_LEVEL"
	OptionLogFormat  = "GEOBLOCK_LOG_FORMAT"
)

// getEnv retrieves the value of the environment variable `key`. If it is not set, it returns the
// `fallback` value.
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

// Updater is the interface for types that can update their databases.
type Updater interface {
	Update() error
}

// autoUpdate updates the databases at regular intervals.
func autoUpdate(resolver Updater) {
	for range time.Tick(autoUpdateInterval) {
		if err := resolver.Update(); err != nil {
			log.Error().Err(err).Msg("Cannot update databases")
			continue
		}
		log.Info().Msg("Databases updated")
	}
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

// statError wraps an error from stat operation.
type statError struct {
	err error
}

func (e *statError) Error() string {
	return fmt.Sprintf("stat error: %v", e.err)
}

func (e *statError) Unwrap() error {
	return e.err
}

// loadError wraps an error from config load operation.
type loadError struct {
	err error
}

func (e *loadError) Error() string {
	return fmt.Sprintf("load error: %v", e.err)
}

func (e *loadError) Unwrap() error {
	return e.err
}

// configReloader watches a config file for changes.
type configReloader struct {
	path     string
	prevStat os.FileInfo
	// Swappable for testing
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

// reloadIfChanged checks if the config file changed and updates the engine if so.
// Returns (true, nil) if reloaded, (false, nil) if unchanged, (false, err) on error.
func (r *configReloader) reloadIfChanged(engine ConfigUpdater) (bool, error) {
	stat, err := r.stat(r.path)
	if err != nil {
		return false, &statError{err: err}
	}

	if r.prevStat.Size() == stat.Size() && r.prevStat.ModTime().Equal(stat.ModTime()) {
		return false, nil
	}

	// Update prevStat to avoid re-triggering on the same change
	r.prevStat = stat

	cfg, err := r.load(r.path)
	if err != nil {
		return false, &loadError{err: err}
	}

	engine.UpdateConfig(&cfg.AccessControl)
	return true, nil
}

// autoReload watches the configuration file for changes and updates the engine when it happens.
func autoReload(engine ConfigUpdater, path string) {
	reloader, err := newConfigReloader(path)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("Cannot watch configuration file")
		return
	}

	for range time.Tick(autoReloadInterval) {
		reloaded, err := reloader.reloadIfChanged(engine)
		if err != nil {
			var statErr *statError
			var loadErr *loadError
			switch {
			case errors.As(err, &statErr):
				log.Error().
					Err(statErr.Unwrap()).
					Str("path", path).
					Msg("Cannot watch configuration file")
			case errors.As(err, &loadErr):
				log.Error().
					Err(loadErr.Unwrap()).
					Str("path", path).
					Msg("Cannot read configuration file")
			default:
				log.Error().Err(err).Str("path", path).Msg("Cannot reload configuration")
			}
			continue
		}
		if reloaded {
			log.Info().Msg("Configuration reloaded")
		}
	}
}

// parseLogLevel parses the log level from string to zerolog.Level. It defaults to info level if the
// provided level is invalid.
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
	default:
		log.Logger = log.Output(
			zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
		)
		if logFormat != LogFormatText {
			log.Warn().Str("format", logFormat).Msg("Invalid log format")
		}
	}

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

	log.Info().Str("version", version.Get()).Msg("Starting Geoblock")
	log.Info().Msg("Loading configuration file")
	cfg, err := loadConfig(options.configPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", options.configPath).Msg(
			"Cannot read configuration file",
		)
	}

	log.Info().Msg("Initializing database resolver")
	resolver := ipinfo.NewResolver()
	if err := resolver.Update(); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize database resolver")
	}

	var (
		address = ":" + options.serverPort
		engine  = rules.NewEngine(&cfg.AccessControl)
		server  = server.NewServer(address, engine, resolver)
	)

	go autoUpdate(resolver)
	go autoReload(engine, options.configPath)

	log.Info().Str("address", server.Addr).Msg("Starting server")
	log.Fatal().Err(server.ListenAndServe()).Msg("Server stopped")
}
