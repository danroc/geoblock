// Package server contains the HTTP authorization server.
package server

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/danroc/geoblock/internal/ipres"
	"github.com/danroc/geoblock/internal/metrics"
	"github.com/danroc/geoblock/internal/rules"
)

// HTTP server timeout constants
const (
	httpTimeoutRead  = 10 * time.Second
	httpTimeoutWrite = 30 * time.Second
	httpTimeoutIdle  = 30 * time.Second
)

// HTTP headers used by reverse proxies to identify the original request.
const (
	headerForwardedMethod = "X-Forwarded-Method"
	headerForwardedHost   = "X-Forwarded-Host"
	headerForwardedFor    = "X-Forwarded-For"
)

// Fields used in the log messages.
const (
	fieldRequestDomain = "request_domain"
	fieldRequestMethod = "request_method"
	fieldRequestStatus = "request_status"
	fieldSourceIP      = "source_ip"
	fieldSourceIsLocal = "source_is_local"
	fieldSourceCountry = "source_country"
	fieldSourceASN     = "source_asn"
	fieldSourceOrg     = "source_org"
)

// Possible request statuses.
const (
	requestStatusInvalid = "invalid"
	requestStatusAllowed = "allowed"
	requestStatusDenied  = "denied"
)

// isAllowedStatus maps the boolean authorization result to a string status.
var isAllowedStatus = map[bool]string{
	true:  requestStatusAllowed,
	false: requestStatusDenied,
}

// localNetworkCIDRs contains the list of local networks CIDRs.
var localNetworkCIDRs = []netip.Prefix{
	netip.MustParsePrefix("10.0.0.0/8"),     // (RFC 1918) Class A private
	netip.MustParsePrefix("172.16.0.0/12"),  // (RFC 1918) Class B private
	netip.MustParsePrefix("192.168.0.0/16"), // (RFC 1918) Class C private
	netip.MustParsePrefix("127.0.0.0/8"),    // (RFC 1122) Loopback
	netip.MustParsePrefix("169.254.0.0/16"), // (RFC 3927) Link‑local
	netip.MustParsePrefix("::1/128"),        // (RFC 4291) IPv6 loopback
	netip.MustParsePrefix("fc00::/7"),       // (RFC 4193) IPv6 unique local
	netip.MustParsePrefix("fe80::/10"),      // (RFC 4291) IPv6 link‑local
}

// isLocalIP checks if the given IP address is a local IP address.
func isLocalIP(ip netip.Addr) bool {
	for _, cidr := range localNetworkCIDRs {
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
		origin = request.Header.Get(headerForwardedFor)
		domain = request.Header.Get(headerForwardedHost)
		method = request.Header.Get(headerForwardedMethod)
	)

	// Block the request if one or more of the required headers are missing. It
	// probably means that the request didn't come from the reverse proxy.
	if origin == "" || domain == "" || method == "" {
		log.WithFields(log.Fields{
			fieldRequestDomain: domain,
			fieldRequestMethod: method,
			fieldRequestStatus: requestStatusInvalid,
			fieldSourceIP:      origin,
		}).Error("Missing required headers")
		writer.WriteHeader(http.StatusBadRequest)
		metrics.IncInvalid()
		return
	}

	// For sanity, we check if the source IP is a valid IP address. If the IP
	// is invalid, we deny the request regardless of the default policy.
	sourceIP, err := netip.ParseAddr(origin)
	if err != nil {
		log.WithFields(log.Fields{
			fieldRequestDomain: domain,
			fieldRequestMethod: method,
			fieldRequestStatus: requestStatusInvalid,
			fieldSourceIP:      origin,
		}).Error("Invalid source IP")
		writer.WriteHeader(http.StatusBadRequest)
		metrics.IncInvalid()
		return
	}

	resolved := resolver.Resolve(sourceIP)
	isAllowed := engine.Authorize(&rules.Query{
		RequestedDomain: domain,
		RequestedMethod: method,
		SourceIP:        sourceIP,
		SourceCountry:   resolved.CountryCode,
		SourceASN:       resolved.ASN,
	})

	logFields := log.Fields{
		fieldRequestDomain: domain,
		fieldRequestMethod: method,
		fieldRequestStatus: isAllowedStatus[isAllowed],
		fieldSourceIP:      sourceIP,
		fieldSourceIsLocal: isLocalIP(sourceIP),
		fieldSourceCountry: resolved.CountryCode,
		fieldSourceASN:     resolved.ASN,
		fieldSourceOrg:     resolved.Organization,
	}

	if isAllowed {
		log.WithFields(logFields).Info("Request allowed")
		writer.WriteHeader(http.StatusNoContent)
		metrics.IncAllowed()
	} else {
		log.WithFields(logFields).Warn("Request denied")
		writer.WriteHeader(http.StatusForbidden)
		metrics.IncDenied()
	}
}

// getHealth returns a 204 status code to indicate that the server is running.
func getHealth(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusNoContent)
}

// getMetrics returns the metrics in JSON format.
func getMetrics(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(metrics.Get()); err != nil {
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
		ReadTimeout:  httpTimeoutRead,
		WriteTimeout: httpTimeoutWrite,
		IdleTimeout:  httpTimeoutIdle,
	}
}
