// Package metrics provides the metrics for the application.
package metrics

import (
	"strings"
	"sync/atomic"

	"github.com/danroc/geoblock/internal/prometheus"
	"github.com/danroc/geoblock/internal/version"
)

// RequestCountSnapshot contains the snapshot of the request count.
type RequestCountSnapshot struct {
	Allowed uint64 `json:"allowed"`
	Denied  uint64 `json:"denied"`
	Invalid uint64 `json:"invalid"`
	Total   uint64 `json:"total"`
}

// Snapshot contains the snapshot of the metrics.
type Snapshot struct {
	Version  string               `json:"version"`
	Requests RequestCountSnapshot `json:"requests"`
}

// RequestCount contains the request count.
type RequestCount struct {
	Denied  atomic.Uint64
	Allowed atomic.Uint64
	Invalid atomic.Uint64
}

var metrics = RequestCount{}

// IncDenied increments the request denied count.
func IncDenied() {
	metrics.Denied.Add(1)
}

// IncAllowed increments the request allowed count.
func IncAllowed() {
	metrics.Allowed.Add(1)
}

// IncInvalid increments the request invalid count.
func IncInvalid() {
	metrics.Invalid.Add(1)
}

// Get returns a snapshot of the metrics.
func Get() *Snapshot {
	var (
		allowed = metrics.Allowed.Load()
		denied  = metrics.Denied.Load()
		invalid = metrics.Invalid.Load()
	)

	return &Snapshot{
		Version: version.Get(),
		Requests: RequestCountSnapshot{
			Denied:  denied,
			Allowed: allowed,
			Invalid: invalid,
			Total:   allowed + denied + invalid,
		},
	}
}

// Prometheus returns metrics formatted in Prometheus exposition format.
func Prometheus() string {
	snapshot := Get()
	metrics := []prometheus.Metric{
		{
			Name: "geoblock_version_info",
			Help: "Version information",
			Type: prometheus.TypeGauge,
			Samples: []prometheus.Sample{
				{
					Labels: map[string]string{
						"version": snapshot.Version,
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
					Value: float64(snapshot.Requests.Allowed),
				},
				{
					Labels: map[string]string{
						"status": "denied",
					},
					Value: float64(snapshot.Requests.Denied),
				},
				{
					Labels: map[string]string{
						"status": "invalid",
					},
					Value: float64(snapshot.Requests.Invalid),
				},
			},
		},
	}

	var output strings.Builder
	for _, metric := range metrics {
		output.WriteString(metric.String())
	}

	return output.String()
}
