// Package main contains the main geoblock application.
package main

import (
	"bytes"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/danroc/geoblock/pkg/config"
	"github.com/danroc/geoblock/pkg/iprange"
	"github.com/danroc/geoblock/pkg/rules"
	"github.com/danroc/geoblock/pkg/server"
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
func autoUpdate(resolver *iprange.Resolver) {
	for range time.Tick(autoUpdateInterval) {
		if err := resolver.Update(); err != nil {
			log.Errorf("Cannot update databases: %v", err)
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
		log.Errorf("Cannot watch configuration file: %v", err)
		return
	}

	for range time.Tick(autoReloadInterval) {
		stat, err := os.Stat(path)
		if err != nil {
			log.Errorf("Cannot watch configuration file: %v", err)
			continue
		}

		if !hasChanged(prevStat, stat) {
			continue
		}
		prevStat = stat

		cfg, err := loadConfig(path)
		if err != nil {
			log.Errorf("Cannot read configuration file: %v", err)
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

	if lvl, err := log.ParseLevel(level); err != nil {
		log.Warnf("Invalid log level: %s", level)
	} else {
		log.SetLevel(lvl)
	}
}

func main() {
	options := getOptions()
	configureLogger(options.logLevel)

	log.Info("Loading configuration file")
	cfg, err := loadConfig(options.configPath)
	if err != nil {
		log.Fatalf("Cannot read configuration file: %v", err)
	}

	log.Info("Initializing database resolver")
	resolver, err := iprange.NewResolver()
	if err != nil {
		log.Fatalf("Cannot initialize database resolver: %v", err)
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
