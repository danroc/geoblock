// Package metrics provides Prometheus metrics for the application.
package metrics //nolint:revive // Package name acceptable for internal package

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/danroc/geoblock/internal/ipinfo"
	"github.com/danroc/geoblock/internal/version"
)

// Status label values for the requests counter.
const (
	StatusAllowed = "allowed"
	StatusDenied  = "denied"
	StatusInvalid = "invalid"
)

// requestDurationBuckets are the bucket boundaries for request duration histogram.
// Buckets range from 1ms to 1s to capture both fast requests and occasional slow ones.
var requestDurationBuckets = []float64{
	0.001,
	0.005,
	0.01,
	0.025,
	0.05,
	0.1,
	0.25,
	0.5,
	1.0,
}

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

	// startTime records when the process started.
	startTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "geoblock_start_time_seconds",
		Help: "Unix timestamp of when the process started",
	})

	// requestsTotal tracks the total number of requests by status.
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "geoblock_requests_total",
			Help: "Total number of requests by status",
		},
		[]string{"status"},
	)

	// requestsByCountry tracks requests by country.
	requestsByCountry = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "geoblock_requests_by_country_total",
			Help: "Total number of requests by country",
		},
		[]string{"country"},
	)

	// requestsByMethod tracks requests by HTTP method.
	requestsByMethod = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "geoblock_requests_by_method_total",
			Help: "Total number of requests by HTTP method",
		},
		[]string{"method"},
	)

	// requestDuration tracks request processing latency.
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "geoblock_request_duration_seconds",
			Help:    "Request processing duration in seconds",
			Buckets: requestDurationBuckets,
		},
		[]string{"status"},
	)

	// ruleMatches tracks which rules are being matched.
	ruleMatches = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "geoblock_rule_matches_total",
			Help: "Total number of requests matched by each rule",
		},
		[]string{"rule_index", "action"},
	)

	// defaultPolicyMatches tracks requests handled by default policy.
	defaultPolicyMatches = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "geoblock_default_policy_matches_total",
			Help: "Total number of requests matched by default policy",
		},
		[]string{"action"},
	)

	// configRulesTotal tracks the number of configured rules.
	configRulesTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "geoblock_config_rules_total",
		Help: "Number of configured access control rules",
	})

	// configReloadTotal tracks config reload attempts.
	configReloadTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "geoblock_config_reload_total",
			Help: "Total number of configuration reload attempts",
		},
		[]string{"result"},
	)

	// configLastReload records the timestamp of last successful config reload.
	configLastReload = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "geoblock_config_last_reload_timestamp",
		Help: "Unix timestamp of last successful config reload",
	})

	// dbEntries tracks the number of entries in IP databases.
	dbEntries = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "geoblock_db_entries",
			Help: "Number of entries in IP database",
		},
		[]string{"database", "ip_version"},
	)

	// dbLastUpdate records the timestamp of last successful database update.
	dbLastUpdate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "geoblock_db_last_update_timestamp",
		Help: "Unix timestamp of last successful database update",
	})

	// dbLoadDuration records the duration of the last database load.
	dbLoadDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "geoblock_db_load_duration_seconds",
		Help: "Duration of last database load in seconds",
	})
)

func init() {
	registry.MustRegister(
		versionInfo,
		startTime,
		requestsTotal,
		requestsByCountry,
		requestsByMethod,
		requestDuration,
		ruleMatches,
		defaultPolicyMatches,
		configRulesTotal,
		configReloadTotal,
		configLastReload,
		dbEntries,
		dbLastUpdate,
		dbLoadDuration,
	)

	versionInfo.WithLabelValues(version.Get()).Set(1)
	startTime.Set(float64(time.Now().Unix()))
}

// PrometheusCollector implements the Collector interface using Prometheus metrics.
type PrometheusCollector struct{}

// NewCollector creates a new PrometheusCollector.
func NewCollector() *PrometheusCollector {
	return &PrometheusCollector{}
}

// RecordRequest records comprehensive metrics for a request.
func (c *PrometheusCollector) RecordRequest(
	status string,
	country string,
	method string,
	duration time.Duration,
	ruleIndex int,
	action string,
	isDefaultPolicy bool,
) {
	requestsTotal.WithLabelValues(status).Inc()
	requestDuration.WithLabelValues(status).Observe(duration.Seconds())

	if country != "" {
		requestsByCountry.WithLabelValues(country).Inc()
	}
	if method != "" {
		requestsByMethod.WithLabelValues(method).Inc()
	}

	if isDefaultPolicy {
		defaultPolicyMatches.WithLabelValues(action).Inc()
	} else {
		ruleMatches.WithLabelValues(strconv.Itoa(ruleIndex), action).Inc()
	}
}

// RecordInvalidRequest records metrics for an invalid request.
func (c *PrometheusCollector) RecordInvalidRequest(duration time.Duration) {
	requestsTotal.WithLabelValues(StatusInvalid).Inc()
	requestDuration.WithLabelValues(StatusInvalid).Observe(duration.Seconds())
}

// RecordConfigReload records a config reload attempt.
func (c *PrometheusCollector) RecordConfigReload(success bool, rulesCount int) {
	if success {
		configReloadTotal.WithLabelValues("success").Inc()
		configLastReload.Set(float64(time.Now().Unix()))
		configRulesTotal.Set(float64(rulesCount))
	} else {
		configReloadTotal.WithLabelValues("failure").Inc()
	}
}

// RecordDBUpdate records an IP database update.
func (c *PrometheusCollector) RecordDBUpdate(
	entries map[ipinfo.DBSource]uint64,
	duration time.Duration,
) {
	for key, count := range entries {
		dbEntries.WithLabelValues(key.DBType, key.IPVersion).Set(float64(count))
	}

	dbLastUpdate.Set(float64(time.Now().Unix()))
	dbLoadDuration.Set(duration.Seconds())
}

// Handler returns an HTTP handler for the metrics endpoint.
func Handler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

// Reset resets all metrics. This is intended for use in tests only.
func Reset() {
	requestsTotal.Reset()
	requestsByCountry.Reset()
	requestsByMethod.Reset()
	requestDuration.Reset()
	ruleMatches.Reset()
	defaultPolicyMatches.Reset()
	configRulesTotal.Set(0)
	configReloadTotal.Reset()
	configLastReload.Set(0)
	dbEntries.Reset()
	dbLastUpdate.Set(0)
	dbLoadDuration.Set(0)
	versionInfo.Reset()
	versionInfo.WithLabelValues(version.Get()).Set(1)
}
