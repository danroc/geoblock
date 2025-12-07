// Package prometheus provides Prometheus format support for metrics.
package prometheus

import (
	"fmt"
	"strings"

	"github.com/danroc/geoblock/internal/utils/maps"
	"github.com/danroc/geoblock/internal/utils/stats"
)

// Metric types
const (
	TypeCounter   = "counter"
	TypeGauge     = "gauge"
	TypeHistogram = "histogram"
)

// Sample represents a single sample of a Prometheus metric.
type Sample struct {
	Name   string
	Labels map[string]string
	Value  float64
}

// Metric represents a single Prometheus metric with its metadata.
type Metric struct {
	Comment string
	Name    string
	Help    string
	Type    string
	Samples []Sample
}

// String formats the metric in Prometheus exposition format.
func (m Metric) String() string {
	var b strings.Builder

	// Comment text
	if m.Comment != "" {
		for _, line := range strings.Split(m.Comment, "\n") {
			fmt.Fprintf(&b, "# %s\n", line)
		}
	}

	// Help text
	if m.Help != "" {
		fmt.Fprintf(&b, "# HELP %s %s\n", m.Name, m.Help)
	}

	// Type information
	if m.Type != "" {
		fmt.Fprintf(&b, "# TYPE %s %s\n", m.Name, m.Type)
	}

	// Write each metric value
	for _, s := range m.Samples {
		// Metric name
		if s.Name != "" {
			b.WriteString(s.Name)
		} else {
			b.WriteString(m.Name)
		}

		// Labels
		if len(s.Labels) > 0 {
			b.WriteString("{")
			for i, k := range maps.SortedKeys(s.Labels) {
				if i > 0 {
					b.WriteString(",")
				}
				fmt.Fprintf(&b, `%s="%s"`, k, escapeLabel(s.Labels[k]))
			}
			b.WriteString("}")
		}

		// Value
		fmt.Fprintf(&b, " %v\n", s.Value)
	}

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

// Format takes multiple metrics and returns them formatted in Prometheus
// exposition format.
func Format(metrics []Metric) string {
	var b strings.Builder
	for i, metric := range metrics {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(metric.String())
	}
	return b.String()
}

// FromHistogram converts a stats.Histogram into a Prometheus Metric.
func FromHistogram(
	name, help, comment string,
	labels map[string]string,
	histogram stats.HistogramSummary,
) Metric {
	buckets := histogram.Buckets
	samples := make([]Sample, 0, len(buckets)+2)

	for _, bucket := range buckets {
		sampleLabels := maps.Merge(labels, map[string]string{
			"le": fmt.Sprintf("%g", bucket.UpperBound),
		})
		samples = append(samples, Sample{
			Name:   name + "_bucket",
			Labels: sampleLabels,
			Value:  float64(bucket.Count),
		})
	}

	samples = append(samples, Sample{
		Name:   name + "_sum",
		Labels: labels,
		Value:  histogram.Sum,
	})

	samples = append(samples, Sample{
		Name:   name + "_count",
		Labels: labels,
		Value:  float64(histogram.Count),
	})

	return Metric{
		Comment: comment,
		Name:    name,
		Help:    help,
		Type:    TypeHistogram,
		Samples: samples,
	}
}
