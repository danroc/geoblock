// Package main contains the main geoblock application.
package main

import (
	"bytes"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/danroc/geoblock/internal/config"
	"github.com/danroc/geoblock/internal/ipres"
	"github.com/danroc/geoblock/internal/rules"
	"github.com/danroc/geoblock/internal/server"
	"github.com/danroc/geoblock/internal/version"
)

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
func autoUpdate(resolver *ipres.Resolver) {
	for range time.Tick(autoUpdateInterval) {
		if err := resolver.Update(); err != nil {
			log.WithError(err).Error("Cannot update databases")
			continue
		}
		log.Info("Databases updated")
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
		log.WithError(err).WithField("path", path).Error(
			"Cannot watch configuration file",
		)
		return
	}

	for range time.Tick(autoReloadInterval) {
		stat, err := os.Stat(path)
		if err != nil {
			log.WithError(err).WithField("path", path).Error(
				"Cannot watch configuration file",
			)
			continue
		}

		if !hasChanged(prevStat, stat) {
			continue
		}

		// Since the file has changed, we update the previous stat.
		prevStat = stat

		cfg, err := loadConfig(path)
		if err != nil {
			log.WithError(err).WithField("path", path).Error(
				"Cannot read configuration file",
			)
			continue
		}

		engine.UpdateConfig(&cfg.AccessControl)
		log.Info("Configuration reloaded")
	}
}

// configureLogger configures the logger with the given log format and level.
func configureLogger(logFormat string, level string) {
	// This should be done first, before any log message is emitted to avoid
	// inconsistent log messages.
	switch logFormat {
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	case "text":
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	default:
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
		log.WithField("format", logFormat).Warn("Invalid log format")
	}

	if parsedLevel, err := log.ParseLevel(level); err != nil {
		log.WithField("level", level).Warn("Invalid log level")
	} else {
		log.SetLevel(parsedLevel)
	}
}

func main() {
	options := getOptions()
	configureLogger(options.logFormat, options.logLevel)

	log.Infof("Starting Geoblock version %s", version.Get())
	log.Info("Loading configuration file")
	cfg, err := loadConfig(options.configPath)
	if err != nil {
		log.WithError(err).WithField("path", options.configPath).Fatal(
			"Cannot read configuration file",
		)
	}

	log.Info("Initializing database resolver")
	resolver := ipres.NewResolver()
	if err := resolver.Update(); err != nil {
		log.WithError(err).Fatal("Cannot initialize database resolver")
	}

	var (
		address = ":" + options.serverPort
		engine  = rules.NewEngine(&cfg.AccessControl)
		server  = server.NewServer(address, engine, resolver)
	)

	go autoUpdate(resolver)
	go autoReload(engine, options.configPath)

	log.Infof("Starting server at %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
