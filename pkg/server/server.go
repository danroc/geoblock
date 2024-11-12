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
	FieldRequestedMethod = "requested_method"
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
	var (
		origin = request.Header.Get(HeaderXForwardedFor)
		domain = request.Header.Get(HeaderXForwardedHost)
		method = request.Header.Get(HeaderXForwardedMethod)
	)

	// Block the request if one or more of the required headers are missing. It
	// probably means that the request didn't come from the reverse proxy.
	if origin == "" || domain == "" || method == "" {
		log.WithFields(log.Fields{
			FieldRequestedDomain: domain,
			FieldRequestedMethod: method,
			FieldSourceIP:        origin,
		}).Error("Missing required headers")
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	// For sanity, we check if the source IP is a valid IP address. If the IP
	// is invalid, we deny the request regardless of the default policy.
	sourceIP := net.ParseIP(origin)
	if sourceIP == nil {
		log.WithFields(log.Fields{
			FieldRequestedDomain: domain,
			FieldRequestedMethod: method,
			FieldSourceIP:        origin,
		}).Error("Invalid source IP")
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	resolved := resolver.Resolve(sourceIP)

	query := &rules.Query{
		RequestedDomain: domain,
		RequestedMethod: method,
		SourceIP:        sourceIP,
		SourceCountry:   resolved.CountryCode,
		SourceASN:       resolved.ASN,
	}

	logFields := log.Fields{
		FieldRequestedDomain: domain,
		FieldRequestedMethod: method,
		FieldSourceIP:        sourceIP,
		FieldSourceCountry:   resolved.CountryCode,
		FieldSourceASN:       resolved.ASN,
		FieldSourceOrg:       resolved.Organization,
	}

	if engine.Authorize(query) {
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
