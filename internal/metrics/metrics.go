// Package metrics provides Prometheus metrics for the application.
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/danroc/geoblock/internal/ipinfo"
	"github.com/danroc/geoblock/internal/version"
)

// RequestRecord contains all data needed to record metrics for a single request.
type RequestRecord struct {
	Status          string
	Country         string
	Method          string
	Duration        time.Duration
	RuleIndex       int
	Action          string
	IsDefaultPolicy bool
}

// RequestCollector collects metrics for HTTP requests.
type RequestCollector interface {
	RecordRequest(record RequestRecord)
	RecordInvalidRequest(duration time.Duration)
}

// Status label values for the requests counter.
const (
	StatusAllowed = "allowed"
	StatusDenied  = "denied"
	StatusInvalid = "invalid"
)

// Result label values for the config reload counter.
const (
	ResultSuccess = "success"
	ResultFailure = "failure"
)

// requestDurationBuckets defines the histogram bucket boundaries for request duration.
var requestDurationBuckets = []float64{
	0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0,
}

// PrometheusCollector implements the Collector interface using Prometheus metrics.
type PrometheusCollector struct {
	registry *prometheus.Registry

	// Process info
	versionInfo *prometheus.GaugeVec
	startTime   prometheus.Gauge

	// Request metrics
	requestsTotal     *prometheus.CounterVec
	requestsByCountry *prometheus.CounterVec
	requestsByMethod  *prometheus.CounterVec
	requestDuration   *prometheus.HistogramVec

	// Rule matching
	ruleMatches          *prometheus.CounterVec
	defaultPolicyMatches *prometheus.CounterVec

	// Configuration
	configRulesTotal  prometheus.Gauge
	configReloadTotal *prometheus.CounterVec
	configLastReload  prometheus.Gauge

	// IP database
	dbEntries      *prometheus.GaugeVec
	dbLastUpdate   prometheus.Gauge
	dbLoadDuration prometheus.Gauge
}

// NewCollector creates a new PrometheusCollector with its own registry.
func NewCollector() *PrometheusCollector {
	c := &PrometheusCollector{
		registry: prometheus.NewRegistry(),

		// Process info
		versionInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "geoblock_version_info",
				Help: "Version information",
			},
			[]string{"version", "commit"},
		),
		startTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "geoblock_start_time_seconds",
			Help: "Unix timestamp of when the process started",
		}),

		// Request metrics
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geoblock_requests_total",
				Help: "Total number of requests by status",
			},
			[]string{"status"},
		),
		requestsByCountry: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geoblock_requests_by_country_total",
				Help: "Total number of requests by country",
			},
			[]string{"country"},
		),
		requestsByMethod: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geoblock_requests_by_method_total",
				Help: "Total number of requests by HTTP method",
			},
			[]string{"method"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "geoblock_request_duration_seconds",
				Help:    "Request processing duration in seconds",
				Buckets: requestDurationBuckets,
			},
			[]string{"status"},
		),

		// Rule matching
		ruleMatches: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geoblock_rule_matches_total",
				Help: "Total number of requests matched by each rule",
			},
			[]string{"rule_index", "action"},
		),
		defaultPolicyMatches: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geoblock_default_policy_matches_total",
				Help: "Total number of requests matched by default policy",
			},
			[]string{"action"},
		),

		// Configuration
		configRulesTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "geoblock_config_rules_total",
			Help: "Number of configured access control rules",
		}),
		configReloadTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geoblock_config_reload_total",
				Help: "Total number of configuration reload attempts",
			},
			[]string{"result"},
		),
		configLastReload: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "geoblock_config_last_reload_timestamp",
			Help: "Unix timestamp of last successful config reload",
		}),

		// IP database
		dbEntries: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "geoblock_db_entries",
				Help: "Number of entries in IP database",
			},
			[]string{"database", "ip_version"},
		),
		dbLastUpdate: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "geoblock_db_last_update_timestamp",
			Help: "Unix timestamp of last successful database update",
		}),
		dbLoadDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "geoblock_db_load_duration_seconds",
			Help: "Duration of last database load in seconds",
		}),
	}

	c.registry.MustRegister(
		// Process info
		c.versionInfo,
		c.startTime,

		// Request metrics
		c.requestsTotal,
		c.requestsByCountry,
		c.requestsByMethod,
		c.requestDuration,

		// Rule matching
		c.ruleMatches,
		c.defaultPolicyMatches,

		// Configuration
		c.configRulesTotal,
		c.configReloadTotal,
		c.configLastReload,

		// IP database
		c.dbEntries,
		c.dbLastUpdate,
		c.dbLoadDuration,
	)

	c.versionInfo.WithLabelValues(version.Version, version.Commit).Set(1)
	c.startTime.Set(float64(time.Now().Unix()))

	return c
}

// RecordRequest records comprehensive metrics for a request.
func (c *PrometheusCollector) RecordRequest(r RequestRecord) {
	c.requestsTotal.WithLabelValues(r.Status).Inc()
	c.requestDuration.WithLabelValues(r.Status).Observe(r.Duration.Seconds())

	if r.Country != "" {
		c.requestsByCountry.WithLabelValues(r.Country).Inc()
	}
	if r.Method != "" {
		c.requestsByMethod.WithLabelValues(r.Method).Inc()
	}

	if r.IsDefaultPolicy {
		c.defaultPolicyMatches.WithLabelValues(r.Action).Inc()
	} else {
		c.ruleMatches.WithLabelValues(
			strconv.Itoa(r.RuleIndex), r.Action,
		).Inc()
	}
}

// RecordInvalidRequest records metrics for an invalid request.
func (c *PrometheusCollector) RecordInvalidRequest(duration time.Duration) {
	c.requestsTotal.WithLabelValues(StatusInvalid).Inc()
	c.requestDuration.WithLabelValues(StatusInvalid).Observe(duration.Seconds())
}

// RecordConfigReload records a config reload attempt.
func (c *PrometheusCollector) RecordConfigReload(success bool, rulesCount int) {
	if success {
		c.configReloadTotal.WithLabelValues(ResultSuccess).Inc()
		c.configLastReload.Set(float64(time.Now().Unix()))
		c.configRulesTotal.Set(float64(rulesCount))
	} else {
		c.configReloadTotal.WithLabelValues(ResultFailure).Inc()
	}
}

// RecordDBUpdate records an IP database update.
func (c *PrometheusCollector) RecordDBUpdate(
	entries map[ipinfo.DBSource]uint64,
	duration time.Duration,
) {
	for key, count := range entries {
		c.dbEntries.WithLabelValues(key.DBType, key.IPVersion).Set(float64(count))
	}

	c.dbLastUpdate.Set(float64(time.Now().Unix()))
	c.dbLoadDuration.Set(duration.Seconds())
}

// Handler returns an HTTP handler for the metrics endpoint.
func (c *PrometheusCollector) Handler() http.Handler {
	return promhttp.HandlerFor(c.registry, promhttp.HandlerOpts{})
}
