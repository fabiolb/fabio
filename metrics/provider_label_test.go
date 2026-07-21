package metrics

import (
	"regexp"
	"strings"
	"testing"
)

func TestLabelProvider(t *testing.T) {
	tests := []struct {
		name           string
		prefix         string
		metricName     string
		labels         []string
		labelValues    []string
		expectedName   string
		expectedLabels string
		counterValue   float64
		gaugeValue     float64
		histoValue     float64
	}{
		{
			name:           "simple_metrics_no_labels",
			prefix:         "test",
			metricName:     "requests",
			labels:         []string{},
			labelValues:    []string{},
			expectedName:   "test.requests",
			expectedLabels: "",
			counterValue:   5.0,
			gaugeValue:     10.0,
			histoValue:     0.123,
		},
		{
			name:           "metrics_with_single_label",
			prefix:         "app",
			metricName:     "http",
			labels:         []string{"method"},
			labelValues:    []string{"method", "GET"},
			expectedName:   "app.http",
			expectedLabels: "|#method:GET",
			counterValue:   15.0,
			gaugeValue:     25.0,
			histoValue:     0.456,
		},
		{
			name:           "metrics_with_multiple_labels",
			prefix:         "service",
			metricName:     "requests",
			labels:         []string{"method", "status", "path"},
			labelValues:    []string{"method", "POST", "status", "200", "path", "/api/users"},
			expectedName:   "service.requests",
			expectedLabels: "|#method:POST,status:200,path:/api/users",
			counterValue:   100.0,
			gaugeValue:     200.0,
			histoValue:     0.789,
		},
		{
			name:           "empty_prefix",
			prefix:         "",
			metricName:     "metric",
			labels:         []string{"label1"},
			labelValues:    []string{"label1", "value1"},
			expectedName:   ".metric",
			expectedLabels: "|#label1:value1",
			counterValue:   1.0,
			gaugeValue:     2.0,
			histoValue:     0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &labelProvider{prefix: tt.prefix}

			// Test Counter
			t.Run("counter", func(t *testing.T) {
				counter := provider.NewCounter(tt.metricName, tt.labels...)
				
				if len(tt.labelValues) > 0 {
					counter = counter.With(tt.labelValues...)
				}

				// Capture stdout
				output := captureStdout(func() {
					counter.Add(tt.counterValue)
				})

				// Verify output format: name:value|c|#labels
				expectedPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(tt.expectedName) + `:\d+\|c` + regexp.QuoteMeta(tt.expectedLabels) + `\n$`)
				if !expectedPattern.MatchString(output) {
					t.Errorf("Counter output doesn't match expected pattern.\nGot: %q\nExpected pattern: %s", output, expectedPattern.String())
				}

				// Verify the output contains expected labels
				if tt.expectedLabels != "" && !strings.Contains(output, tt.expectedLabels) {
					t.Errorf("Counter output missing expected labels.\nGot: %q\nExpected to contain: %q", output, tt.expectedLabels)
				}
			})

			// Test Gauge
			t.Run("gauge", func(t *testing.T) {
				gauge := provider.NewGauge(tt.metricName, tt.labels...)

				if len(tt.labelValues) > 0 {
					gauge = gauge.With(tt.labelValues...)
				}

				// Test Set
				output := captureStdout(func() {
					gauge.Set(tt.gaugeValue)
				})

				expectedPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(tt.expectedName) + `:\d+\|g` + regexp.QuoteMeta(tt.expectedLabels) + `\n$`)
				if !expectedPattern.MatchString(output) {
					t.Errorf("Gauge Set output doesn't match expected pattern.\nGot: %q\nExpected pattern: %s", output, expectedPattern.String())
				}

				// Test Add
				output = captureStdout(func() {
					gauge.Add(5.0)
				})

				if !expectedPattern.MatchString(output) {
					t.Errorf("Gauge Add output doesn't match expected pattern.\nGot: %q\nExpected pattern: %s", output, expectedPattern.String())
				}

				// Verify the output contains expected labels
				if tt.expectedLabels != "" && !strings.Contains(output, tt.expectedLabels) {
					t.Errorf("Gauge output missing expected labels.\nGot: %q\nExpected to contain: %q", output, tt.expectedLabels)
				}
			})

			// Test Histogram
			t.Run("histogram", func(t *testing.T) {
				histogram := provider.NewHistogram(tt.metricName, tt.labels...)

				if len(tt.labelValues) > 0 {
					histogram = histogram.With(tt.labelValues...)
				}

				output := captureStdout(func() {
					histogram.Observe(tt.histoValue)
				})

				// Verify output format: name:value|ms|#labels
				expectedPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(tt.expectedName) + `:\d+\|ms` + regexp.QuoteMeta(tt.expectedLabels) + `\n$`)
				if !expectedPattern.MatchString(output) {
					t.Errorf("Histogram output doesn't match expected pattern.\nGot: %q\nExpected pattern: %s", output, expectedPattern.String())
				}

				// Verify the output contains expected labels
				if tt.expectedLabels != "" && !strings.Contains(output, tt.expectedLabels) {
					t.Errorf("Histogram output missing expected labels.\nGot: %q\nExpected to contain: %q", output, tt.expectedLabels)
				}
			})
		})
	}
}

func TestLabelProvider_EmptyLabels(t *testing.T) {
	provider := &labelProvider{prefix: "test"}

	// Test with empty label arrays
	counter := provider.NewCounter("metric")
	gauge := provider.NewGauge("metric")
	histogram := provider.NewHistogram("metric")

	// Should not panic
	output := captureStdout(func() {
		counter.Add(1.0)
		gauge.Set(2.0)
		histogram.Observe(0.003)
	})

	// Verify output doesn't have label section (or has empty label section)
	lines := strings.SplitSeq(strings.TrimSpace(output), "\n")
	for line := range lines {
		// Should end with |c, |g, or |ms (no labels)
		if !strings.HasSuffix(line, "|c") && !strings.HasSuffix(line, "|g") && !strings.HasSuffix(line, "|ms") {
			t.Errorf("Expected line to end with metric type suffix, got: %q", line)
		}
	}
}

func TestLabelProvider_PartialLabelValues(t *testing.T) {
	provider := &labelProvider{prefix: "test"}

	// Create counter with 3 labels but only provide 2 label-value pairs
	counter := provider.NewCounter("requests", "method", "status", "path")
	counter = counter.With("method", "GET", "status", "200") // Only 2 pairs for 3 labels

	output := captureStdout(func() {
		counter.Add(1.0)
	})

	// Should still work, just with partial labels
	if !strings.Contains(output, "method:GET") || !strings.Contains(output, "status:200") {
		t.Errorf("Counter output missing expected labels: %q", output)
	}
}
