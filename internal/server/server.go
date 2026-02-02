// Package server contains the HTTP authorization server.
package server

import (
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/danroc/geoblock/internal/ipinfo"
	"github.com/danroc/geoblock/internal/metrics"
	"github.com/danroc/geoblock/internal/rules"
)

// HTTP server timeout constants
const (
	httpTimeoutIdle  = 30 * time.Second
	httpTimeoutRead  = 10 * time.Second
	httpTimeoutWrite = 30 * time.Second
)

// HTTP headers used by reverse proxies to identify the original request.
const (
	headerForwardedFor    = "X-Forwarded-For"
	headerForwardedHost   = "X-Forwarded-Host"
	headerForwardedMethod = "X-Forwarded-Method"
)

// Fields used in the log messages
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

// Possible request statuses
const (
	requestStatusAllowed = "allowed"
	requestStatusDenied  = "denied"
	requestStatusInvalid = "invalid"
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

// parseForwardedFor extracts the client IP from the X-Forwarded-For header.
//
// The header can contain a comma-separated list of IPs, where the first IP is typically
// the original client IP.
func parseForwardedFor(header string) string {
	ips := strings.Split(header, ",")
	return strings.TrimSpace(ips[0])
}

// getLogEvent returns a zerolog event based on the authorization result.
func getLogEvent(isAllowed bool) *zerolog.Event {
	if isAllowed {
		return log.Info()
	}
	return log.Warn()
}

// getForwardAuth checks if the request is authorized to access the requested resource.
// It uses the reverse proxy headers to determine the source IP and requested domain.
func getForwardAuth(
	writer http.ResponseWriter,
	request *http.Request,
	resolver *ipinfo.Resolver,
	engine *rules.Engine,
) {
	start := time.Now()

	var (
		origin = parseForwardedFor(request.Header.Get(headerForwardedFor))
		domain = request.Header.Get(headerForwardedHost)
		method = request.Header.Get(headerForwardedMethod)
	)

	// Block the request if one or more of the required headers are missing. It
	// probably means that the request didn't come from the reverse proxy.
	if origin == "" || domain == "" || method == "" {
		log.Error().
			Str(fieldRequestDomain, domain).
			Str(fieldRequestMethod, method).
			Str(fieldRequestStatus, requestStatusInvalid).
			Str(fieldSourceIP, origin).
			Msg("Missing required headers")
		writer.WriteHeader(http.StatusBadRequest)
		metrics.RecordInvalidRequest(time.Since(start))
		return
	}

	// For sanity, we check if the source IP is a valid IP address. If the IP
	// is invalid, we deny the request regardless of the default policy.
	sourceIP, err := netip.ParseAddr(origin)
	if err != nil {
		log.Error().
			Str(fieldRequestDomain, domain).
			Str(fieldRequestMethod, method).
			Str(fieldRequestStatus, requestStatusInvalid).
			Str(fieldSourceIP, origin).
			Msg("Invalid source IP")
		writer.WriteHeader(http.StatusBadRequest)
		metrics.RecordInvalidRequest(time.Since(start))
		return
	}

	resolved := resolver.Resolve(sourceIP)
	result := engine.Authorize(&rules.Query{
		RequestedDomain: domain,
		RequestedMethod: method,
		SourceIP:        sourceIP,
		SourceCountry:   resolved.CountryCode,
		SourceASN:       resolved.ASN,
	})

	duration := time.Since(start)
	status := isAllowedStatus[result.Allowed]

	// Prepare a zerolog event for structured logging
	event := getLogEvent(result.Allowed).
		Str(fieldRequestDomain, domain).
		Str(fieldRequestMethod, method).
		Str(fieldRequestStatus, status).
		Str(fieldSourceIP, sourceIP.String()).
		Bool(fieldSourceIsLocal, isLocalIP(sourceIP)).
		Str(fieldSourceCountry, resolved.CountryCode).
		Uint32(fieldSourceASN, resolved.ASN).
		Str(fieldSourceOrg, resolved.Organization)

	if result.Allowed {
		event.Msg("Request allowed")
		writer.WriteHeader(http.StatusNoContent)
	} else {
		event.Msg("Request denied")
		writer.WriteHeader(http.StatusForbidden)
	}

	metrics.RecordRequest(
		status,
		resolved.CountryCode,
		method,
		duration,
		result.RuleIndex,
		result.Action,
		result.IsDefaultPolicy,
	)
}

// getHealth returns a 204 status code to indicate that the server is running.
func getHealth(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusNoContent)
}

// New creates a new HTTP server that listens on the given address.
func New(address string, engine *rules.Engine, resolver *ipinfo.Resolver) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(
		"GET /v1/forward-auth",
		func(writer http.ResponseWriter, request *http.Request) {
			getForwardAuth(writer, request, resolver, engine)
		},
	)
	mux.HandleFunc("GET /v1/health", getHealth)
	mux.Handle("GET /metrics", metrics.Handler())

	return &http.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  httpTimeoutRead,
		WriteTimeout: httpTimeoutWrite,
		IdleTimeout:  httpTimeoutIdle,
	}
}
