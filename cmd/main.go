package main

import (
	"fmt"
	"net"
	"net/http"
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

	fmt.Printf("Query: %+v\n", query)

	if engine.Authorize(query) {
		writer.WriteHeader(http.StatusNoContent)
	} else {
		writer.WriteHeader(http.StatusForbidden)
	}
}

func main() {
	config, err := schema.ReadFile("examples/configuration.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	engine := rules.NewEngine(config.AccessControl)

	resolver, err := database.NewResolver()
	if err != nil {
		fmt.Println(err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/forward-auth", func(writer http.ResponseWriter, request *http.Request) {
		getForwardAuth(writer, request, resolver, engine)
	})

	server := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Infof("Starting server at %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
