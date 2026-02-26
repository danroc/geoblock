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

// PrometheusCollector collects application metrics using Prometheus.
type PrometheusCollector struct {
	Registry             *prometheus.Registry
	VersionInfo          *prometheus.GaugeVec
	StartTime            prometheus.Gauge
	RequestsTotal        *prometheus.CounterVec
	RequestsByCountry    *prometheus.CounterVec
	RequestsByMethod     *prometheus.CounterVec
	RequestDuration      *prometheus.HistogramVec
	RuleMatches          *prometheus.CounterVec
	DefaultPolicyMatches *prometheus.CounterVec
	ConfigRulesTotal     prometheus.Gauge
	ConfigReloadTotal    *prometheus.CounterVec
	ConfigLastReload     prometheus.Gauge
	DBEntries            *prometheus.GaugeVec
	DBLastUpdate         prometheus.Gauge
	DBLoadDuration       prometheus.Gauge
}

// NewCollector creates a new PrometheusCollector with its own registry.
func NewCollector() *PrometheusCollector {
	c := &PrometheusCollector{
		Registry: prometheus.NewRegistry(),

		VersionInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "geoblock_version_info",
				Help: "Version information",
			},
			[]string{"version", "commit"},
		),

		StartTime: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "geoblock_start_time_seconds",
			Help: "Unix timestamp of when the process started",
		}),

		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geoblock_requests_total",
				Help: "Total number of requests by status",
			},
			[]string{"status"},
		),

		RequestsByCountry: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geoblock_requests_by_country_total",
				Help: "Total number of requests by country",
			},
			[]string{"country"},
		),

		RequestsByMethod: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geoblock_requests_by_method_total",
				Help: "Total number of requests by HTTP method",
			},
			[]string{"method"},
		),

		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "geoblock_request_duration_seconds",
				Help:    "Request processing duration in seconds",
				Buckets: requestDurationBuckets,
			},
			[]string{"status"},
		),

		RuleMatches: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geoblock_rule_matches_total",
				Help: "Total number of requests matched by each rule",
			},
			[]string{"rule_index", "action"},
		),

		DefaultPolicyMatches: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geoblock_default_policy_matches_total",
				Help: "Total number of requests matched by default policy",
			},
			[]string{"action"},
		),

		ConfigRulesTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "geoblock_config_rules_total",
			Help: "Number of configured access control rules",
		}),

		ConfigReloadTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "geoblock_config_reload_total",
				Help: "Total number of configuration reload attempts",
			},
			[]string{"result"},
		),

		ConfigLastReload: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "geoblock_config_last_reload_timestamp",
			Help: "Unix timestamp of last successful config reload",
		}),

		DBEntries: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "geoblock_db_entries",
				Help: "Number of entries in IP database",
			},
			[]string{"database", "ip_version"},
		),

		DBLastUpdate: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "geoblock_db_last_update_timestamp",
			Help: "Unix timestamp of last successful database update",
		}),

		DBLoadDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "geoblock_db_load_duration_seconds",
			Help: "Duration of last database load in seconds",
		}),
	}

	c.Registry.MustRegister(
		c.VersionInfo,
		c.StartTime,
		c.RequestsTotal,
		c.RequestsByCountry,
		c.RequestsByMethod,
		c.RequestDuration,
		c.RuleMatches,
		c.DefaultPolicyMatches,
		c.ConfigRulesTotal,
		c.ConfigReloadTotal,
		c.ConfigLastReload,
		c.DBEntries,
		c.DBLastUpdate,
		c.DBLoadDuration,
	)

	c.VersionInfo.WithLabelValues(version.Version, version.Commit).Set(1)
	c.StartTime.Set(float64(time.Now().Unix()))

	return c
}

// RecordRequest records comprehensive metrics for a request.
func (c *PrometheusCollector) RecordRequest(r RequestRecord) {
	c.RequestsTotal.WithLabelValues(r.Status).Inc()
	c.RequestDuration.WithLabelValues(r.Status).Observe(r.Duration.Seconds())

	if r.Country != "" {
		c.RequestsByCountry.WithLabelValues(r.Country).Inc()
	}
	if r.Method != "" {
		c.RequestsByMethod.WithLabelValues(r.Method).Inc()
	}

	if r.IsDefaultPolicy {
		c.DefaultPolicyMatches.WithLabelValues(r.Action).Inc()
	} else {
		c.RuleMatches.WithLabelValues(
			strconv.Itoa(r.RuleIndex), r.Action,
		).Inc()
	}
}

// RecordInvalidRequest records metrics for an invalid request.
func (c *PrometheusCollector) RecordInvalidRequest(duration time.Duration) {
	c.RequestsTotal.WithLabelValues(StatusInvalid).Inc()
	c.RequestDuration.WithLabelValues(StatusInvalid).Observe(duration.Seconds())
}

// RecordConfigReload records a config reload attempt.
func (c *PrometheusCollector) RecordConfigReload(success bool, rulesCount int) {
	if success {
		c.ConfigReloadTotal.WithLabelValues("success").Inc()
		c.ConfigLastReload.Set(float64(time.Now().Unix()))
		c.ConfigRulesTotal.Set(float64(rulesCount))
	} else {
		c.ConfigReloadTotal.WithLabelValues("failure").Inc()
	}
}

// RecordDBUpdate records an IP database update.
func (c *PrometheusCollector) RecordDBUpdate(
	entries map[ipinfo.DBSource]uint64,
	duration time.Duration,
) {
	for key, count := range entries {
		c.DBEntries.WithLabelValues(key.DBType, key.IPVersion).Set(float64(count))
	}

	c.DBLastUpdate.Set(float64(time.Now().Unix()))
	c.DBLoadDuration.Set(duration.Seconds())
}

// Handler returns an HTTP handler for the metrics endpoint.
func (c *PrometheusCollector) Handler() http.Handler {
	return promhttp.HandlerFor(c.Registry, promhttp.HandlerOpts{})
}
