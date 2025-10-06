// Package prometheus provides Prometheus format support for metrics.
package prometheus

import (
	"fmt"
	"strings"

	"github.com/danroc/geoblock/internal/utils/maps"
)

// Metric represents a single Prometheus metric with its metadata.
type Metric struct {
	Name   string
	Help   string
	Type   string
	Labels map[string]string
	Value  float64
}

// String formats the metric in Prometheus exposition format.
func (m Metric) String() string {
	var b strings.Builder

	// Help text
	if m.Help != "" {
		fmt.Fprintf(&b, "# HELP %s %s\n", m.Name, m.Help)
	}

	// Type information
	if m.Type != "" {
		fmt.Fprintf(&b, "# TYPE %s %s\n", m.Name, m.Type)
	}

	// Metric name
	b.WriteString(m.Name)

	// Labels
	if len(m.Labels) > 0 {
		b.WriteString("{")
		for i, k := range maps.SortedKeys(m.Labels) {
			if i > 0 {
				b.WriteString(",")
			}
			fmt.Fprintf(&b, `%s="%s"`, k, escapeLabel(m.Labels[k]))
		}
		b.WriteString("}")
	}

	// Metric value
	fmt.Fprintf(&b, " %v\n", m.Value)
	return b.String()
}

// escapeLabel escapes a label accordingly to Prometheus format spec.
// See: https://prometheus.io/docs/instrumenting/exposition_formats/#text-format-details
func escapeLabel(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	v = strings.ReplaceAll(v, "\n", `\n`)
	return v
}
