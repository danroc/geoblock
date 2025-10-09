package prometheus

import "testing"

func TestMetric_String(t *testing.T) {
	tests := []struct {
		name     string
		metric   Metric
		expected string
	}{
		{
			name: "full metric",
			metric: Metric{
				Name: "test_metric",
				Help: "This is a test metric",
				Type: "counter",
				Samples: []Sample{
					{
						Labels: map[string]string{
							"label1": "value1",
							"label2": "value2",
						},
						Value: 42,
					},
				},
			},
			expected: `# HELP test_metric This is a test metric
# TYPE test_metric counter
test_metric{label1="value1",label2="value2"} 42
`,
		},
		{
			name: "simple metric",
			metric: Metric{
				Name: "simple_metric",
				Samples: []Sample{
					{
						Value: 1,
					},
				},
			},
			expected: "simple_metric 1\n",
		},
		{
			name: "gauge metric without labels",
			metric: Metric{
				Name: "gauge_metric",
				Help: "A gauge metric",
				Type: "gauge",
				Samples: []Sample{
					{
						Value: 3.14,
					},
				},
			},
			expected: `# HELP gauge_metric A gauge metric
# TYPE gauge_metric gauge
gauge_metric 3.14
`,
		},
		{
			name: "labeled metric",
			metric: Metric{
				Name: "labeled_metric",
				Samples: []Sample{
					{
						Labels: map[string]string{
							"status": "success",
						},
						Value: 100,
					},
				},
			},
			expected: `labeled_metric{status="success"} 100
`,
		},
		{
			name: "empty help and type",
			metric: Metric{
				Name: "empty_meta_metric",
				Samples: []Sample{
					{
						Value: 1,
					},
				},
			},
			expected: "empty_meta_metric 1\n",
		},
		{
			name: "single label",
			metric: Metric{
				Name: "single_label_metric",
				Samples: []Sample{
					{
						Labels: map[string]string{
							"env": "production",
						},
						Value: 1,
					},
				},
			},
			expected: `single_label_metric{env="production"} 1
`,
		},
		{
			name: "special characters in labels",
			metric: Metric{
				Name: "special_chars_metric",
				Samples: []Sample{
					{
						Labels: map[string]string{
							"path":   "/api/v1/metrics",
							"method": "GET",
						},
						Value: 10,
					},
				},
			},
			expected: `special_chars_metric{method="GET",path="/api/v1/metrics"} 10
`,
		},
		{
			name: "empty name",
			metric: Metric{
				Samples: []Sample{
					{
						Value: 1,
					},
				},
			},
			expected: " 1\n",
		},
		{
			name: "empty labels map",
			metric: Metric{
				Name: "empty_labels_metric",
				Samples: []Sample{
					{
						Labels: map[string]string{},
						Value:  42,
					},
				},
			},
			expected: "empty_labels_metric 42\n",
		},
		{
			name: "nil labels map",
			metric: Metric{
				Name: "nil_labels_metric",
				Samples: []Sample{
					{
						Labels: nil,
						Value:  42,
					},
				},
			},
			expected: "nil_labels_metric 42\n",
		},
		{
			name: "sorted labels",
			metric: Metric{
				Name: "multi_label_metric",
				Samples: []Sample{
					{
						Labels: map[string]string{
							"c": "third",
							"a": "first",
							"b": "second",
						},
						Value: 1,
					},
				},
			},
			expected: `multi_label_metric{a="first",b="second",c="third"} 1
`,
		},
		{
			name: "escaped labels",
			metric: Metric{
				Name: "escaped_metric",
				Samples: []Sample{
					{
						Labels: map[string]string{
							"quote":     `value with " quote`,
							"backslash": `value with \ backslash`,
							"newline":   "value with \n newline",
						},
						Value: 1,
					},
				},
			},
			expected: `escaped_metric{backslash="value with \\ backslash",newline="value with \n newline",quote="value with \" quote"} 1
`,
		},
		{
			name: "float values",
			metric: Metric{
				Name: "float_metric",
				Samples: []Sample{
					{
						Value: 0.001,
					},
				},
			},
			expected: "float_metric 0.001\n",
		},
		{
			name: "large float",
			metric: Metric{
				Name: "large_metric",
				Samples: []Sample{
					{
						Value: 1234567.89,
					},
				},
			},
			expected: "large_metric 1.23456789e+06\n",
		},
		{
			name: "negative float",
			metric: Metric{
				Name: "negative_metric",
				Samples: []Sample{
					{
						Value: -42.5,
					},
				},
			},
			expected: "negative_metric -42.5\n",
		},
		{
			name: "multiple values",
			metric: Metric{
				Name: "multi_value_metric",
				Help: "A metric with multiple values",
				Type: "counter",
				Samples: []Sample{
					{
						Labels: map[string]string{
							"status": "success",
						},
						Value: 100,
					},
					{
						Labels: map[string]string{
							"status": "error",
						},
						Value: 5,
					},
				},
			},
			expected: `# HELP multi_value_metric A metric with multiple values
# TYPE multi_value_metric counter
multi_value_metric{status="success"} 100
multi_value_metric{status="error"} 5
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.metric.String(); got != tt.expected {
				t.Errorf(
					"unexpected output:\n--- expected ---\n%s--- got ---\n%s",
					tt.expected,
					got,
				)
			}
		})
	}
}
