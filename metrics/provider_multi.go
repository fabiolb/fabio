package metrics

import gkm "github.com/go-kit/kit/metrics"

// MultiProvider wraps zero or more providers.
type MultiProvider struct {
	p []Provider
}

func NewMultiProvider(p []Provider) *MultiProvider {
	return &MultiProvider{p}
}

// NewCounter creates a MultiCounter with counter objects for all registered
// providers.
func (mp *MultiProvider) NewCounter(name string, labels ...string) gkm.Counter {
	var c []gkm.Counter
	for _, p := range mp.p {
		c = append(c, p.NewCounter(name, labels...))
	}
	return &MultiCounter{c}
}

// NewGauge creates a MultiGauge with gauge objects for all registered
// providers.
func (mp *MultiProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	var v []gkm.Gauge
	for _, p := range mp.p {
		v = append(v, p.NewGauge(name, labels...))
	}
	return &MultiGauge{v}
}

func (mp *MultiProvider) NewHistogram(name string, labels ...string) gkm.Histogram {
	var h []gkm.Histogram
	for _, p := range mp.p {
		h = append(h, p.NewHistogram(name, labels...))
	}
	return &MultiHistogram{h: h}
}

// MultiCounter wraps zero or more counters.
type MultiCounter struct {
	c []gkm.Counter
}

func (mc *MultiCounter) Add(v float64) {
	for _, c := range mc.c {
		c.Add(v)
	}
}

func (mc *MultiCounter) With(labels ...string) gkm.Counter {
	cc := make([]gkm.Counter, len(mc.c))
	for i := range mc.c {
		cc[i] = mc.c[i].With(labels...)
	}
	return &MultiCounter{c: cc}
}

// DeleteLabelValues deletes the metric with the given label values from all underlying counters.
func (mc *MultiCounter) DeleteLabelValues(labelValues ...string) bool {
	deleted := false
	for _, c := range mc.c {
		if dc, ok := c.(DeletableCounter); ok {
			if dc.DeleteLabelValues(labelValues...) {
				deleted = true
			}
		}
	}
	return deleted
}

// MultiGauge wraps zero or more gauges.
type MultiGauge struct {
	v []gkm.Gauge
}

func (m *MultiGauge) Set(n float64) {
	for _, v := range m.v {
		v.Set(n)
	}
}

func (m *MultiGauge) With(labels ...string) gkm.Gauge {
	vc := make([]gkm.Gauge, len(m.v))
	for i := range m.v {
		vc[i] = m.v[i].With(labels...)
	}
	return &MultiGauge{v: vc}
}

func (m *MultiGauge) Add(val float64) {
	for _, v := range m.v {
		v.Add(val)
	}
}

// DeleteLabelValues deletes the metric with the given label values from all underlying gauges.
func (m *MultiGauge) DeleteLabelValues(labelValues ...string) bool {
	deleted := false
	for _, g := range m.v {
		if dg, ok := g.(DeletableGauge); ok {
			if dg.DeleteLabelValues(labelValues...) {
				deleted = true
			}
		}
	}
	return deleted
}

type MultiHistogram struct {
	h []gkm.Histogram
}

func (m *MultiHistogram) With(labelValues ...string) gkm.Histogram {
	hc := make([]gkm.Histogram, len(m.h))
	for i := range m.h {
		hc[i] = m.h[i].With(labelValues...)
	}
	return &MultiHistogram{h: hc}
}

func (m *MultiHistogram) Observe(value float64) {
	for _, v := range m.h {
		v.Observe(value)
	}
}

// DeleteLabelValues deletes the metric with the given label values from all underlying histograms.
func (m *MultiHistogram) DeleteLabelValues(labelValues ...string) bool {
	deleted := false
	for _, h := range m.h {
		if dh, ok := h.(DeletableHistogram); ok {
			if dh.DeleteLabelValues(labelValues...) {
				deleted = true
			}
		}
	}
	return deleted
}
