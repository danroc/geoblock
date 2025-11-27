// Package main contains the main geoblock application.
package main

import (
	"bytes"
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
const RFC3339Milli = "2006-01-02T15:04:05.999Z07:00"

// Auto-update and auto-reload intervals.
const (
	autoUpdateInterval = 24 * time.Hour
	autoReloadInterval = 5 * time.Second
)

// Log levels.
const (
	LogLevelInfo  = "info"
	LogLevelDebug = "debug"
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
	OptionConfigPath = "GEOBLOCK_CONFIG"
	OptionServerPort = "GEOBLOCK_PORT"
	OptionLogLevel   = "GEOBLOCK_LOG_LEVEL"
	OptionLogFormat  = "GEOBLOCK_LOG_FORMAT"
)

// getEnv retrieves the value of the environment variable `key`. If it is not
// set, it returns the `fallback` value.
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

// autoUpdate updates the databases at regular intervals.
func autoUpdate(resolver *ipinfo.Resolver) {
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

// hasChanged returns true if the two file infos are different. It only checks
// the size and the modification time.
func hasChanged(a, b os.FileInfo) bool {
	return a.Size() != b.Size() || a.ModTime() != b.ModTime()
}

// autoReload watches the configuration file for changes and updates the engine
// when it happens.
func autoReload(engine *rules.Engine, path string) {
	prevStat, err := os.Stat(path)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", path).
			Msg("Cannot watch configuration file")
		return
	}

	for range time.Tick(autoReloadInterval) {
		stat, err := os.Stat(path)
		if err != nil {
			log.Error().
				Err(err).
				Str("path", path).
				Msg("Cannot watch configuration file")
			continue
		}

		if !hasChanged(prevStat, stat) {
			continue
		}

		// Since the file has changed, we update the previous stat.
		prevStat = stat

		cfg, err := loadConfig(path)
		if err != nil {
			log.Error().
				Err(err).
				Str("path", path).
				Msg("Cannot read configuration file")
			continue
		}

		engine.UpdateConfig(&cfg.AccessControl)
		log.Info().Msg("Configuration reloaded")
	}
}

// configureLogger configures the logger with the given log format and level.
func configureLogger(logFormat, level string) {
	// Configure log format
	switch logFormat {
	case "json":
		zerolog.TimeFieldFormat = RFC3339Milli
	case "text":
		log.Logger = log.Output(
			zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
		)
	default:
		log.Logger = log.Output(
			zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
		)
		log.Warn().Str("format", logFormat).Msg("Invalid log format")
	}

	// Configure log level
	switch level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Warn().Str("level", level).Msg("Invalid log level")
	}
}

func main() {
	options := getOptions()
	configureLogger(options.logFormat, options.logLevel)

	log.Info().Str("version", version.Get()).Msg("Starting Geoblock")
	log.Debug().Msg("Loading configuration file")
	cfg, err := loadConfig(options.configPath)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("path", options.configPath).
			Msg("Cannot read configuration file")
	}

	log.Debug().Msg("Initializing database resolver")
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
