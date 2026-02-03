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
