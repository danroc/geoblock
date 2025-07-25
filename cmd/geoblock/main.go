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
)

const (
	autoUpdateInterval = 24 * time.Hour
	autoReloadInterval = 5 * time.Second
)

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
}

// getOptions returns the application options from the environment variables.
func getOptions() *appOptions {
	return &appOptions{
		configPath: getEnv("GEOBLOCK_CONFIG", "/etc/geoblock/config.yaml"),
		serverPort: getEnv("GEOBLOCK_PORT", "8080"),
		logLevel:   getEnv("GEOBLOCK_LOG_LEVEL", "info"),
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
		log.WithError(err).Error("Cannot watch configuration file")
		return
	}

	for range time.Tick(autoReloadInterval) {
		stat, err := os.Stat(path)
		if err != nil {
			log.WithError(err).Error("Cannot watch configuration file")
			continue
		}

		if !hasChanged(prevStat, stat) {
			continue
		}
		prevStat = stat

		cfg, err := loadConfig(path)
		if err != nil {
			log.WithError(err).Error("Cannot read configuration file")
			continue
		}

		engine.UpdateConfig(&cfg.AccessControl)
		log.Info("Configuration reloaded")
	}
}

// configureLogger configures the logger with the given log level and sets the
// formatter.
func configureLogger(level string) {
	// This should be done first, before any log message is emitted to avoid
	// inconsistent log messages.
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if parsedLevel, err := log.ParseLevel(level); err != nil {
		log.WithField("level", level).Warn("Invalid log level")
	} else {
		log.SetLevel(parsedLevel)
	}
}

func main() {
	options := getOptions()
	configureLogger(options.logLevel)

	log.Info("Loading configuration file")
	cfg, err := loadConfig(options.configPath)
	if err != nil {
		log.WithError(err).Fatal("Cannot read configuration file")
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
