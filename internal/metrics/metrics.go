// Package metrics provides Prometheus metrics for the application.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/danroc/geoblock/internal/version"
)

// Status label values for the requests counter.
const (
	StatusAllowed = "allowed"
	StatusDenied  = "denied"
	StatusInvalid = "invalid"
)

var (
	// registry is a custom registry to avoid exposing Go runtime metrics.
	registry = prometheus.NewRegistry()

	// versionInfo exposes version information as a gauge.
	versionInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "geoblock_version_info",
			Help: "Version information",
		},
		[]string{"version"},
	)

	// requestsTotal tracks the total number of requests by status.
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "geoblock_requests_total",
			Help: "Total number of requests by status",
		},
		[]string{"status"},
	)
)

func init() {
	registry.MustRegister(versionInfo, requestsTotal)
	versionInfo.WithLabelValues(version.Get()).Set(1)
}

// IncAllowed increments the allowed requests counter.
func IncAllowed() {
	requestsTotal.WithLabelValues(StatusAllowed).Inc()
}

// IncDenied increments the denied requests counter.
func IncDenied() {
	requestsTotal.WithLabelValues(StatusDenied).Inc()
}

// IncInvalid increments the invalid requests counter.
func IncInvalid() {
	requestsTotal.WithLabelValues(StatusInvalid).Inc()
}

// Handler returns an HTTP handler for the metrics endpoint.
func Handler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

// Reset resets all metrics. This is intended for use in tests only.
func Reset() {
	requestsTotal.Reset()
	versionInfo.Reset()
	versionInfo.WithLabelValues(version.Get()).Set(1)
}
