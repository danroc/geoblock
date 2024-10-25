package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

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

func getAuthorize(
	w http.ResponseWriter,
	r *http.Request,
	resolver *database.Resolver,
	engine *rules.Engine,
) {
	origin := r.Header.Get(HeaderXForwardedFor)
	domain := r.Header.Get(HeaderXForwardedHost)

	// Block the request if one or more of the required headers are missing. It
	// probably means that the request didn't come from a reverse proxy.
	if origin == "" || domain == "" {
		w.WriteHeader(http.StatusForbidden)
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
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusForbidden)
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
		getAuthorize(writer, request, resolver, engine)
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("Starting server at %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
