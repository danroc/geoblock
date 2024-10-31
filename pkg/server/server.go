// Package server contains the HTTP authorization server.
package server

import (
	"net"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/danroc/geoblock/pkg/database"
	"github.com/danroc/geoblock/pkg/rules"
)

// HTTP headers used by reverse proxies to identify the original request.
const (
	HeaderXForwardedMethod = "X-Forwarded-Method"
	HeaderXForwardedProto  = "X-Forwarded-Proto"
	HeaderXForwardedHost   = "X-Forwarded-Host"
	HeaderXForwardedURI    = "X-Forwarded-Uri"
	HeaderXForwardedFor    = "X-Forwarded-For"
)

// Fields used in the log messages.
const (
	FieldRequestedDomain = "requested_domain"
	FieldSourceIP        = "source_ip"
	FieldSourceCountry   = "source_country"
	FieldSourceASN       = "source_asn"
	FieldSourceOrg       = "source_org"
)

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
		log.WithFields(log.Fields{
			FieldRequestedDomain: domain,
			FieldSourceIP:        origin,
		}).Warn("Missing required headers")
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
		FieldRequestedDomain: domain,
		FieldSourceIP:        sourceIP,
		FieldSourceCountry:   resolution.CountryCode,
		FieldSourceASN:       resolution.ASN,
		FieldSourceOrg:       resolution.Organization,
	}

	if engine.Authorize(&query) {
		log.WithFields(logFields).Info("Request authorized")
		writer.WriteHeader(http.StatusNoContent)
	} else {
		log.WithFields(logFields).Warn("Request denied")
		writer.WriteHeader(http.StatusForbidden)
	}
}

// getHealth returns a 204 status code to indicate that the server is running.
func getHealth(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusNoContent)
}

// NewServer creates a new HTTP server that listens on the given address.
func NewServer(
	address string,
	engine *rules.Engine,
	resolver *database.Resolver,
) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"GET /v1/forward-auth",
		func(writer http.ResponseWriter, request *http.Request) {
			getForwardAuth(writer, request, resolver, engine)
		},
	)
	mux.HandleFunc(
		"GET /v1/health",
		func(writer http.ResponseWriter, request *http.Request) {
			getHealth(writer, request)
		},
	)

	return &http.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
}
