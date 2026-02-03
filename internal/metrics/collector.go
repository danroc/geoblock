// Package metrics provides Prometheus metrics for the application.
package metrics

import "time"

// RequestCollector collects metrics for HTTP requests.
type RequestCollector interface {
	RecordRequest(
		status string,
		country string,
		method string,
		duration time.Duration,
		ruleIndex int,
		action string,
		isDefaultPolicy bool,
	)

	RecordInvalidRequest(duration time.Duration)
}

// DBUpdateCollector collects metrics for database updates.
type DBUpdateCollector interface {
	RecordDBUpdate(entries map[string]uint64, duration time.Duration)
}

// ConfigReloadCollector collects metrics for config reloads.
type ConfigReloadCollector interface {
	RecordConfigReload(success bool, rulesCount int)
}

// Collector combines all metric collection interfaces.
type Collector interface {
	RequestCollector
	DBUpdateCollector
	ConfigReloadCollector
}

// NopCollector is a no-op collector for testing or when metrics are disabled.
type NopCollector struct{}

// RecordRequest is a no-op.
func (NopCollector) RecordRequest(
	_ string,
	_ string,
	_ string,
	_ time.Duration,
	_ int,
	_ string,
	_ bool,
) {
}

// RecordInvalidRequest is a no-op.
func (NopCollector) RecordInvalidRequest(_ time.Duration) {}

// RecordDBUpdate is a no-op.
func (NopCollector) RecordDBUpdate(_ map[string]uint64, _ time.Duration) {}

// RecordConfigReload is a no-op.
func (NopCollector) RecordConfigReload(_ bool, _ int) {}
