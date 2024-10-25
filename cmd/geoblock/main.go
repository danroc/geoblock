package main

import (
	"flag"
	"net"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/danroc/geoblock/pkg/database"
	"github.com/danroc/geoblock/pkg/rules"
	"github.com/danroc/geoblock/pkg/schema"
)

const (
	HeaderXForwardedMethod = "X-Forwarded-Method"
	HeaderXForwardedProto  = "X-Forwarded-Proto"
	HeaderXForwardedHost   = "X-Forwarded-Host"
	HeaderXForwardedURI    = "X-Forwarded-Uri"
	HeaderXForwardedFor    = "X-Forwarded-For"
)

type Rule struct {
	Policy    string
	Networks  []net.IPNet
	Domains   []string
	Countries []string
}

type Service struct {
	DefaultPolicy string
	Rules         []Rule
}

// getForwardAuth checks if the request is authorized to access the requested
// resource. It uses the reverse proxy headers to determine the source IP and
// requested domain.
func getForwardAuth(
	writer http.ResponseWriter,
	request *http.Request,
	resolver *database.Resolver,
	engine *rules.Engine,
) {
	origin := request.Header.Get(HeaderXForwardedFor)
	domain := request.Header.Get(HeaderXForwardedHost)

	// Block the request if one or more of the required headers are missing. It
	// probably means that the request didn't come from the reverse proxy.
	if origin == "" || domain == "" {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	sourceIP := net.ParseIP(origin)
	resolution := resolver.Resolve(sourceIP)

	query := rules.Query{
		RequestedDomain: domain,
		SourceIP:        sourceIP,
		SourceCountry:   resolution.CountryCode,
		SourceASN:       resolution.ASN,
	}

	logFields := log.Fields{
		"requested_domain": domain,
		"source_ip":        sourceIP,
		"source_country":   resolution.CountryCode,
		"source_asn":       resolution.ASN,
		"source_org":       resolution.Organization,
	}

	if engine.Authorize(&query) {
		log.WithFields(logFields).Info("Request authorized")
		writer.WriteHeader(http.StatusNoContent)
	} else {
		log.WithFields(logFields).Warn("Request denied")
		writer.WriteHeader(http.StatusForbidden)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

type commandFlags struct {
	configPath string
	serverPort string
}

const (
	DefaultConfigPath = "config/geoblock.yaml"
	DefaultServerPort = "8080"
)

const usage = `Usage: geoblock [options]

Options:
  -c, --config CONFIG  Path to the configuration file (default "config/geoblock.yaml")
  -p, --port   PORT    Port to run the server on (default "8080")`

func parseFlags() *commandFlags {
	flags := &commandFlags{}

	flag.StringVar(&flags.configPath, "c", DefaultConfigPath, "")
	flag.StringVar(&flags.configPath, "config", DefaultConfigPath, "")

	flag.StringVar(&flags.serverPort, "p", DefaultServerPort, "")
	flag.StringVar(&flags.serverPort, "port", DefaultServerPort, "")

	flag.Usage = func() {
		println(usage)
	}

	flag.Parse()
	return flags
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	flags := parseFlags()

	// var (
	// 	configPath = getEnv("GEOBLOCK_CONFIG", "config/geoblock.yaml")
	// 	serverPort = getEnv("GEOBLOCK_PORT", "8080")
	// )

	log.Info("Loading configuration file")
	config, err := schema.ReadFile(flags.configPath)
	if err != nil {
		log.Fatalf("Failed to read configuration file: %v", err)
	}

	log.Info("Initializing database resolver")
	resolver, err := database.NewResolver()
	if err != nil {
		log.Fatalf("Cannot initialize database resolver: %v", err)
	}

	engine := rules.NewEngine(&config.AccessControl)
	mux := http.NewServeMux()
	mux.HandleFunc(
		"/v1/forward-auth",
		func(writer http.ResponseWriter, request *http.Request) {
			getForwardAuth(writer, request, resolver, engine)
		},
	)

	server := http.Server{
		Addr:         ":" + flags.serverPort,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Infof("Starting server at %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
