// Package metrics provides the metrics for the application.
package metrics

import (
	"fmt"
	"sync/atomic"

	"github.com/danroc/geoblock/internal/prometheus"
	"github.com/danroc/geoblock/internal/utils/maps"
	"github.com/danroc/geoblock/internal/utils/stats"
	"github.com/danroc/geoblock/internal/version"
)

// RequestCount contains the request count.
type RequestCount struct {
	Denied  atomic.Uint64
	Allowed atomic.Uint64
	Invalid atomic.Uint64
}

type HistogramKey struct {
	Method  string
	Handler string
	Status  int
}

var (
	durationsHistogram = maps.NewSyncMap[HistogramKey, *stats.Histogram]()
	requestCount       = RequestCount{}
)

// IncDenied increments the request denied count.
func IncDenied() {
	requestCount.Denied.Add(1)
}

// IncAllowed increments the request allowed count.
func IncAllowed() {
	requestCount.Allowed.Add(1)
}

// IncInvalid increments the request invalid count.
func IncInvalid() {
	requestCount.Invalid.Add(1)
}

// newHistogram creates a new histogram with default buckets.
func newHistogram() *stats.Histogram {
	return stats.NewHistogram([]float64{
		0.001, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0, 10.0,
	})
}

// ObserveDuration records a new observation in the durations histogram.
func ObserveDuration(method, handler string, status int, duration float64) {
	histogram, _ := durationsHistogram.GetOrSet(
		HistogramKey{
			Method:  method,
			Handler: handler,
			Status:  status,
		}, newHistogram(),
	)
	histogram.Observe(duration)
}

// Prometheus returns metrics formatted in Prometheus exposition format.
func Prometheus() string {
	metrics := []prometheus.Metric{
		{
			Name: "geoblock_version_info",
			Help: "Version information",
			Type: prometheus.TypeGauge,
			Samples: []prometheus.Sample{
				{
					Labels: map[string]string{
						"version": version.Get(),
					},
					Value: 1,
				},
			},
		},
		{
			Name: "geoblock_requests_total",
			Help: "Total number of requests by status",
			Type: prometheus.TypeCounter,
			Samples: []prometheus.Sample{
				{
					Labels: map[string]string{
						"status": "allowed",
					},
					Value: float64(requestCount.Allowed.Load()),
				},
				{
					Labels: map[string]string{
						"status": "denied",
					},
					Value: float64(requestCount.Denied.Load()),
				},
				{
					Labels: map[string]string{
						"status": "invalid",
					},
					Value: float64(requestCount.Invalid.Load()),
				},
			},
		},
	}

	durationsHistogram.Range(func(key HistogramKey, histogram *stats.Histogram) bool {
		metrics = append(metrics, prometheus.FromHistogram(
			"geoblock_request_duration_seconds",
			"Geoblock request duration in seconds",
			"",
			map[string]string{
				"method":  key.Method,
				"handler": key.Handler,
				"status":  fmt.Sprintf("%d", key.Status),
			},
			histogram.Summary(),
		))
		return true
	})

	return prometheus.Format(metrics)
}
