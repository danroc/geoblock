// Package server contains the HTTP authorization server.
package server

import (
	"fmt"
	"net/http"
	"net/netip"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/danroc/geoblock/internal/ipres"
	"github.com/danroc/geoblock/internal/rules"
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
	FieldRequestDomain   = "request_domain"
	FieldRequestMethod   = "request_method"
	FieldSourceIP        = "source_ip"
	FieldSourceIPPrivate = "source_ip_private"
	FieldSourceCountry   = "source_country"
	FieldSourceASN       = "source_asn"
	FieldSourceOrg       = "source_org"
)

// Metrics contains the metric values of the server.
type Metrics struct {
	Denied  atomic.Uint64
	Allowed atomic.Uint64
	Invalid atomic.Uint64
}

var metrics = Metrics{}

// privateCIDRs contains the list of private networks.
var privateCIDRs = []netip.Prefix{
	netip.MustParsePrefix("10.0.0.0/8"),     // (RFC 1918) Class A private
	netip.MustParsePrefix("172.16.0.0/12"),  // (RFC 1918) Class B private
	netip.MustParsePrefix("192.168.0.0/16"), // (RFC 1918) Class C private
	netip.MustParsePrefix("127.0.0.0/8"),    // (RFC 1122) Loopback
	netip.MustParsePrefix("169.254.0.0/16"), // (RFC 3927) Link‑local
	netip.MustParsePrefix("::1/128"),        // (RFC 4291) IPv6 loopback
	netip.MustParsePrefix("fc00::/7"),       // (RFC 4193) IPv6 unique local
	netip.MustParsePrefix("fe80::/10"),      // (RFC 4291) IPv6 link‑local
}

// isPrivateIP checks if the given IP address is a private IP address.
func isPrivateIP(ip netip.Addr) bool {
	for _, cidr := range privateCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// getForwardAuth checks if the request is authorized to access the requested
// resource. It uses the reverse proxy headers to determine the source IP and
// requested domain.
func getForwardAuth(
	writer http.ResponseWriter,
	request *http.Request,
	resolver *ipres.Resolver,
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
			FieldRequestDomain: domain,
			FieldRequestMethod: method,
			FieldSourceIP:      origin,
		}).Error("Missing required headers")
		writer.WriteHeader(http.StatusBadRequest)
		metrics.Invalid.Add(1)
		return
	}

	// For sanity, we check if the source IP is a valid IP address. If the IP
	// is invalid, we deny the request regardless of the default policy.
	sourceIP, err := netip.ParseAddr(origin)
	if err != nil {
		log.WithFields(log.Fields{
			FieldRequestDomain: domain,
			FieldRequestMethod: method,
			FieldSourceIP:      origin,
		}).Error("Invalid source IP")
		writer.WriteHeader(http.StatusBadRequest)
		metrics.Invalid.Add(1)
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
		FieldRequestDomain:   domain,
		FieldRequestMethod:   method,
		FieldSourceIP:        sourceIP,
		FieldSourceIPPrivate: isPrivateIP(sourceIP),
		FieldSourceCountry:   resolved.CountryCode,
		FieldSourceASN:       resolved.ASN,
		FieldSourceOrg:       resolved.Organization,
	}

	if engine.Authorize(query) {
		log.WithFields(logFields).Info("Request authorized")
		writer.WriteHeader(http.StatusNoContent)
		metrics.Allowed.Add(1)
	} else {
		log.WithFields(logFields).Warn("Request denied")
		writer.WriteHeader(http.StatusForbidden)
		metrics.Denied.Add(1)
	}
}

// getHealth returns a 204 status code to indicate that the server is running.
func getHealth(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusNoContent)
}

// getMetrics returns the metrics in JSON format.
func getMetrics(writer http.ResponseWriter, _ *http.Request) {
	var (
		denied  = metrics.Denied.Load()
		allowed = metrics.Allowed.Load()
		invalid = metrics.Invalid.Load()
		total   = denied + allowed + invalid
	)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(
		[]byte(
			fmt.Sprintf(
				`{"denied": %d, "allowed": %d, "invalid": %d, "total": %d}`,
				denied, allowed, invalid, total,
			),
		),
	); err != nil {
		log.WithError(err).Error("Cannot write metrics response")
	}
}

// NewServer creates a new HTTP server that listens on the given address.
func NewServer(
	address string,
	engine *rules.Engine,
	resolver *ipres.Resolver,
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
	mux.HandleFunc(
		"GET /v1/metrics",
		func(writer http.ResponseWriter, request *http.Request) {
			getMetrics(writer, request)
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
