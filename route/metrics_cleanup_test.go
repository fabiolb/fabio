package route

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fabiolb/fabio/metrics"
	gkm "github.com/go-kit/kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// TestMetricsCleanup verifies that stale metrics are cleaned up when routes are removed
func TestMetricsCleanup(t *testing.T) {
	// Create a custom prometheus registry for this test
	reg := prometheus.NewRegistry()

	// Create histogram and counter vecs
	hv := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "fabio",
		Name:      "route",
		Help:      "test",
		Buckets:   prometheus.DefBuckets,
	}, []string{"service", "host", "path", "target"})
	reg.MustRegister(hv)

	rxCv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "fabio",
		Name:      "route_rx_total",
		Help:      "test",
	}, []string{"service", "host", "path", "target"})
	reg.MustRegister(rxCv)

	txCv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "fabio",
		Name:      "route_tx_total",
		Help:      "test",
	}, []string{"service", "host", "path", "target"})
	reg.MustRegister(txCv)

	// Create a test provider that wraps our vecs
	counters.histogram = &testDeletableHistogram{hv: hv}
	counters.rxCounter = &testDeletableCounter{cv: rxCv}
	counters.txCounter = &testDeletableCounter{cv: txCv}

	// Create initial table with two services
	t1, err := NewTable(bytes.NewBufferString(`
route add svc-a /path-a http://target-a:8080/
route add svc-b /path-b http://target-b:8080/
`))
	if err != nil {
		t.Fatalf("Failed to create table 1: %v", err)
	}

	// Store it
	SetTable(t1)

	// Simulate traffic to create metric series
	for _, routes := range t1 {
		for _, r := range routes {
			for _, target := range r.Targets {
				target.Timer.Observe(0.1)
				target.RxCounter.Add(100)
				target.TxCounter.Add(200)
			}
		}
	}

	// Check that metrics exist for both services
	mfs, _ := reg.Gather()
	t.Log("Metrics after initial traffic:")
	svcAFound, svcBFound := false, false
	for _, mf := range mfs {
		for _, m := range mf.GetMetric() {
			var labels []string
			for _, l := range m.GetLabel() {
				labels = append(labels, l.GetName()+"="+l.GetValue())
				if l.GetValue() == "svc-a" {
					svcAFound = true
				}
				if l.GetValue() == "svc-b" {
					svcBFound = true
				}
			}
			t.Logf("  %s{%s}", mf.GetName(), strings.Join(labels, ", "))
		}
	}

	if !svcAFound {
		t.Error("svc-a metrics not found after initial traffic")
	}
	if !svcBFound {
		t.Error("svc-b metrics not found after initial traffic")
	}

	// Now create a new table WITHOUT svc-a
	t2, err := NewTable(bytes.NewBufferString(`
route add svc-b /path-b http://target-b:8080/
`))
	if err != nil {
		t.Fatalf("Failed to create table 2: %v", err)
	}

	t.Logf("Old table keys: %v", collectTableMetricKeys(t1))
	t.Logf("New table keys: %v", collectTableMetricKeys(t2))

	// Set the new table - this should trigger cleanup
	SetTable(t2)

	// Check that svc-a metrics are gone
	mfs, _ = reg.Gather()
	t.Log("Metrics after removing svc-a:")
	svcAFound, svcBFound = false, false
	for _, mf := range mfs {
		for _, m := range mf.GetMetric() {
			var labels []string
			for _, l := range m.GetLabel() {
				labels = append(labels, l.GetName()+"="+l.GetValue())
				if l.GetValue() == "svc-a" {
					svcAFound = true
				}
				if l.GetValue() == "svc-b" {
					svcBFound = true
				}
			}
			t.Logf("  %s{%s}", mf.GetName(), strings.Join(labels, ", "))
		}
	}

	if svcAFound {
		t.Error("svc-a metrics should have been cleaned up but were still found")
	}
	if !svcBFound {
		t.Error("svc-b metrics should still exist but were not found")
	}
}

// testDeletableHistogram wraps HistogramVec for testing
type testDeletableHistogram struct {
	hv  *prometheus.HistogramVec
	lvs []string
}

func (h *testDeletableHistogram) Observe(v float64) {
	h.hv.WithLabelValues(extractValues(h.lvs)...).Observe(v)
}

func (h *testDeletableHistogram) With(labelValues ...string) gkm.Histogram {
	return &testDeletableHistogram{
		hv:  h.hv,
		lvs: append(append([]string{}, h.lvs...), labelValues...),
	}
}

func (h *testDeletableHistogram) DeleteLabelValues(labelValues ...string) bool {
	return h.hv.DeleteLabelValues(labelValues...)
}

// testDeletableCounter wraps CounterVec for testing
type testDeletableCounter struct {
	cv  *prometheus.CounterVec
	lvs []string
}

func (c *testDeletableCounter) Add(v float64) {
	c.cv.WithLabelValues(extractValues(c.lvs)...).Add(v)
}

func (c *testDeletableCounter) With(labelValues ...string) gkm.Counter {
	return &testDeletableCounter{
		cv:  c.cv,
		lvs: append(append([]string{}, c.lvs...), labelValues...),
	}
}

func (c *testDeletableCounter) DeleteLabelValues(labelValues ...string) bool {
	return c.cv.DeleteLabelValues(labelValues...)
}

// extractValues extracts only values from alternating key-value pairs
func extractValues(lvs []string) []string {
	vals := make([]string, 0, len(lvs)/2)
	for i := 1; i < len(lvs); i += 2 {
		vals = append(vals, lvs[i])
	}
	return vals
}

// Verify interfaces are satisfied
var _ metrics.DeletableHistogram = (*testDeletableHistogram)(nil)
var _ metrics.DeletableCounter = (*testDeletableCounter)(nil)

// TestMetricsCleanupViaMultiProvider tests the cleanup through the MultiProvider
// which is the actual code path used in production.
func TestMetricsCleanupViaMultiProvider(t *testing.T) {
	// Create a custom prometheus registry for this test
	reg := prometheus.NewRegistry()

	// Create a test provider that mimics the real PromProvider
	testProv := &testPromProvider{reg: reg}

	// Wrap it in a MultiProvider like production does
	multiProv := metrics.NewMultiProvider([]metrics.Provider{testProv})

	// Set up the route package to use this provider
	SetMetricsProvider(multiProv)

	// Verify that the histogram is a MultiHistogram
	_, ok := counters.histogram.(*metrics.MultiHistogram)
	if !ok {
		t.Fatalf("Expected MultiHistogram, got %T", counters.histogram)
	}

	// Also verify it implements DeletableHistogram
	_, ok = counters.histogram.(metrics.DeletableHistogram)
	if !ok {
		t.Fatal("MultiHistogram should implement DeletableHistogram")
	}

	// Create initial table with two services
	t1, err := NewTable(bytes.NewBufferString(`
route add svc-a /path-a http://target-a:8080/
route add svc-b /path-b http://target-b:8080/
`))
	if err != nil {
		t.Fatalf("Failed to create table 1: %v", err)
	}
	SetTable(t1)

	// Simulate traffic to create metric series
	for _, routes := range t1 {
		for _, r := range routes {
			for _, target := range r.Targets {
				target.Timer.Observe(0.1)
			}
		}
	}

	// Verify svc-a metrics exist
	mfs, _ := reg.Gather()
	svcAFound := false
	for _, mf := range mfs {
		for _, m := range mf.GetMetric() {
			for _, l := range m.GetLabel() {
				if l.GetValue() == "svc-a" {
					svcAFound = true
				}
			}
		}
	}
	if !svcAFound {
		t.Error("svc-a metrics should exist after traffic")
	}

	// Remove svc-a from the table
	t2, err := NewTable(bytes.NewBufferString(`
route add svc-b /path-b http://target-b:8080/
`))
	if err != nil {
		t.Fatalf("Failed to create table 2: %v", err)
	}
	SetTable(t2)

	// Verify svc-a metrics are gone
	mfs, _ = reg.Gather()
	svcAFound = false
	for _, mf := range mfs {
		for _, m := range mf.GetMetric() {
			for _, l := range m.GetLabel() {
				if l.GetValue() == "svc-a" {
					svcAFound = true
					t.Logf("Found unexpected metric: %s with label %s=%s", mf.GetName(), l.GetName(), l.GetValue())
				}
			}
		}
	}
	if svcAFound {
		t.Error("svc-a metrics should have been cleaned up via MultiProvider")
	}
}

// testPromProvider is a test provider that creates deletable metrics
type testPromProvider struct {
	reg *prometheus.Registry
}

func (p *testPromProvider) NewCounter(name string, labels ...string) gkm.Counter {
	cv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "test",
		Name:      sanitizeName(name),
	}, labels)
	p.reg.MustRegister(cv)
	return &testDeletableCounter{cv: cv}
}

func (p *testPromProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	gv := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "test",
		Name:      sanitizeName(name),
	}, labels)
	p.reg.MustRegister(gv)
	return &testDeletableGauge{gv: gv}
}

func (p *testPromProvider) NewHistogram(name string, labels ...string) gkm.Histogram {
	hv := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "test",
		Name:      sanitizeName(name),
		Buckets:   prometheus.DefBuckets,
	}, labels)
	p.reg.MustRegister(hv)
	return &testDeletableHistogram{hv: hv}
}

// sanitizeName replaces dots with underscores for prometheus metric names
func sanitizeName(name string) string {
	return strings.ReplaceAll(name, ".", "_")
}

// testDeletableGauge for completeness
type testDeletableGauge struct {
	gv  *prometheus.GaugeVec
	lvs []string
}

func (g *testDeletableGauge) Set(v float64) {
	g.gv.WithLabelValues(extractValues(g.lvs)...).Set(v)
}

func (g *testDeletableGauge) Add(v float64) {
	g.gv.WithLabelValues(extractValues(g.lvs)...).Add(v)
}

func (g *testDeletableGauge) With(labelValues ...string) gkm.Gauge {
	return &testDeletableGauge{
		gv:  g.gv,
		lvs: append(append([]string{}, g.lvs...), labelValues...),
	}
}

func (g *testDeletableGauge) DeleteLabelValues(labelValues ...string) bool {
	return g.gv.DeleteLabelValues(labelValues...)
}

var _ metrics.DeletableGauge = (*testDeletableGauge)(nil)
var _ metrics.Provider = (*testPromProvider)(nil)
