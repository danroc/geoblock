package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/danroc/geoblock/pkg/database"
	"github.com/danroc/geoblock/pkg/rules"
	"github.com/danroc/geoblock/pkg/schema"
	"github.com/danroc/geoblock/pkg/server"
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
}

func getOptions() *appOptions {
	return &appOptions{
		configPath: getEnv("GEOBLOCK_CONFIG", "config/geoblock.yaml"),
		serverPort: getEnv("GEOBLOCK_PORT", "8080"),
	}
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	options := getOptions()

	log.Info("Loading configuration file")
	config, err := schema.ReadFile(options.configPath)
	if err != nil {
		log.Fatalf("Failed to read configuration file: %v", err)
	}

	log.Info("Initializing database resolver")
	resolver, err := database.NewResolver()
	if err != nil {
		log.Fatalf("Cannot initialize database resolver: %v", err)
	}

	var (
		address = ":" + options.serverPort
		engine  = rules.NewEngine(&config.AccessControl)
		server  = server.NewServer(address, engine, resolver)
	)

	log.Infof("Starting server at %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
