package metrics

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"testing"
)

func TestFlatProvider(t *testing.T) {
	tests := []struct {
		name           string
		prefix         string
		metricName     string
		labels         []string
		expectedName   string
		counterValue   float64
		gaugeValue     float64
		histoValue     float64
	}{
		{
			name:           "simple_metrics_no_labels",
			prefix:         "test",
			metricName:     "requests",
			labels:         []string{},
			expectedName:   "test.requests",
			counterValue:   5.0,
			gaugeValue:     10.0,
			histoValue:     0.123,
		},
		{
			name:           "metrics_with_labels",
			prefix:         "app",
			metricName:     "http",
			labels:         []string{"service", "host", "path"},
			expectedName:   "app.http.service.host.path",
			counterValue:   15.0,
			gaugeValue:     25.0,
			histoValue:     0.456,
		},
		{
			name:           "empty_prefix",
			prefix:         "",
			metricName:     "metric",
			labels:         []string{"label1"},
			expectedName:   ".metric.label1",
			counterValue:   1.0,
			gaugeValue:     2.0,
			histoValue:     0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &flatProvider{prefix: tt.prefix}

			// Test Counter
			t.Run("counter", func(t *testing.T) {
				counter := provider.NewCounter(tt.metricName, tt.labels...)
				
				// Capture stdout
				output := captureStdout(func() {
					counter.Add(tt.counterValue)
				})

				// Verify output format: name:value|c
				expectedPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(tt.expectedName) + `:\d+\|c\n$`)
				if !expectedPattern.MatchString(output) {
					t.Errorf("Counter output doesn't match expected pattern.\nGot: %q\nExpected pattern: %s", output, expectedPattern.String())
				}

				// Test With() returns same counter (flat provider ignores label values)
				counter2 := counter.With("value1", "value2")
				if counter2 != counter {
					t.Error("With() should return the same counter for flat provider")
				}
			})

			// Test Gauge
			t.Run("gauge", func(t *testing.T) {
				gauge := provider.NewGauge(tt.metricName, tt.labels...)

				// Test Set
				output := captureStdout(func() {
					gauge.Set(tt.gaugeValue)
				})

				expectedPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(tt.expectedName) + `:\d+\|g\n$`)
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

				// Test With() returns same gauge
				gauge2 := gauge.With("value1")
				if gauge2 != gauge {
					t.Error("With() should return the same gauge for flat provider")
				}
			})

			// Test Histogram
			t.Run("histogram", func(t *testing.T) {
				histogram := provider.NewHistogram(tt.metricName, tt.labels...)

				output := captureStdout(func() {
					histogram.Observe(tt.histoValue)
				})

				// Verify output format: :name:value|ms
				expectedPattern := regexp.MustCompile(`^:` + regexp.QuoteMeta(tt.expectedName) + `:\d+\|ms\n$`)
				if !expectedPattern.MatchString(output) {
					t.Errorf("Histogram output doesn't match expected pattern.\nGot: %q\nExpected pattern: %s", output, expectedPattern.String())
				}

				// Test With() returns same histogram
				histogram2 := histogram.With("value1")
				if histogram2 != histogram {
					t.Error("With() should return the same histogram for flat provider")
				}
			})
		})
	}
}

// captureStdout captures stdout during function execution
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
