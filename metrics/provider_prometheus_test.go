package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPromProviderDeleteLabelValues(t *testing.T) {
	// Create a new registry to avoid conflicts with other tests
	reg := prometheus.NewRegistry()

	// Create histogram directly (not via provider to control registration)
	hv := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "test",
		Name:      "route",
		Help:      "test histogram",
		Buckets:   prometheus.DefBuckets,
	}, []string{"service", "host", "path", "target"})
	reg.MustRegister(hv)

	ph := &promHistogram{hv: hv}

	// Create metrics for two targets
	h1 := ph.With("service", "svc1", "host", "host1", "path", "/path1", "target", "http://target1/")
	h2 := ph.With("service", "svc2", "host", "host2", "path", "/path2", "target", "http://target2/")

	// Observe some values
	h1.Observe(0.1)
	h1.Observe(0.2)
	h2.Observe(0.3)

	// Verify both metrics exist
	count := testutil.CollectAndCount(hv)
	t.Logf("Metric count before delete: %d", count)
	if count == 0 {
		t.Fatal("Expected metrics to be registered")
	}

	// Gather metrics to check labels
	metrics, _ := reg.Gather()
	t.Logf("Metrics before delete:")
	for _, m := range metrics {
		for _, metric := range m.GetMetric() {
			var labels []string
			for _, l := range metric.GetLabel() {
				labels = append(labels, l.GetName()+"="+l.GetValue())
			}
			t.Logf("  %s{%s}", m.GetName(), strings.Join(labels, ", "))
		}
	}

	// Delete the first target's metrics (note: values only, not key-value pairs)
	deleted := ph.DeleteLabelValues("svc1", "host1", "/path1", "http://target1/")
	t.Logf("DeleteLabelValues returned: %v", deleted)

	// Gather metrics after delete
	metrics, _ = reg.Gather()
	t.Logf("Metrics after delete:")
	foundSvc1 := false
	foundSvc2 := false
	for _, m := range metrics {
		for _, metric := range m.GetMetric() {
			var labels []string
			for _, l := range metric.GetLabel() {
				labels = append(labels, l.GetName()+"="+l.GetValue())
				if l.GetName() == "service" && l.GetValue() == "svc1" {
					foundSvc1 = true
				}
				if l.GetName() == "service" && l.GetValue() == "svc2" {
					foundSvc2 = true
				}
			}
			t.Logf("  %s{%s}", m.GetName(), strings.Join(labels, ", "))
		}
	}

	if foundSvc1 {
		t.Error("svc1 metrics should have been deleted but were still found")
	}
	if !foundSvc2 {
		t.Error("svc2 metrics should still exist but were not found")
	}
}

func TestPromHistogramWithLabelValues(t *testing.T) {
	// Test that With() correctly accumulates label values
	hv := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "test2",
		Name:      "route2",
		Help:      "test histogram",
		Buckets:   prometheus.DefBuckets,
	}, []string{"service", "host", "path", "target"})

	ph := &promHistogram{hv: hv}

	// Add labels in stages like the real code does
	h := ph.With("service", "mysvc", "host", "myhost", "path", "/mypath", "target", "http://mytarget/")

	// Check internal state
	ph2 := h.(*promHistogram)
	t.Logf("lvs after With: %v", ph2.lvs)
	t.Logf("lvs length: %d", len(ph2.lvs))

	expectedLvs := []string{"service", "mysvc", "host", "myhost", "path", "/mypath", "target", "http://mytarget/"}
	if len(ph2.lvs) != len(expectedLvs) {
		t.Errorf("Expected lvs length %d, got %d", len(expectedLvs), len(ph2.lvs))
	}

	// Try to observe - this should not panic
	h.Observe(0.5)
	t.Log("Observe succeeded")
}
