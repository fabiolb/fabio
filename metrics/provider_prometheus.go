package metrics

import (
	gkm "github.com/go-kit/kit/metrics"
	promclient "github.com/prometheus/client_golang/prometheus"
)

type PromProvider struct {
	Opts    promclient.Opts
	Buckets []float64
}

func NewPromProvider(namespace, subsystem string, buckets []float64) Provider {
	namespace = clean(namespace)
	if len(subsystem) > 0 {
		subsystem = clean(subsystem)
	}
	return &PromProvider{
		Opts: promclient.Opts{
			Namespace: namespace,
			Subsystem: subsystem,
		},
		Buckets: buckets,
	}
}

func (p *PromProvider) NewCounter(name string, labels ...string) gkm.Counter {
	copts := promclient.CounterOpts(p.Opts)
	copts.Name = clean(name)
	cv := promclient.NewCounterVec(copts, labels)
	promclient.MustRegister(cv)
	return &promCounter{cv: cv}
}

func (p *PromProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	gopts := promclient.GaugeOpts(p.Opts)
	gopts.Name = clean(name)
	gv := promclient.NewGaugeVec(gopts, labels)
	promclient.MustRegister(gv)
	return &promGauge{gv: gv}
}

func (p *PromProvider) NewHistogram(name string, labels ...string) gkm.Histogram {
	hopts := promclient.HistogramOpts{
		Namespace:   p.Opts.Namespace,
		Subsystem:   p.Opts.Subsystem,
		Name:        clean(name),
		Help:        p.Opts.Help,
		ConstLabels: p.Opts.ConstLabels,
		Buckets:     p.Buckets,
	}
	hv := promclient.NewHistogramVec(hopts, labels)
	promclient.MustRegister(hv)
	return &promHistogram{hv: hv}
}

// makeLabels converts a slice of alternating key-value pairs into prometheus.Labels.
// e.g., ["service", "foo", "host", "bar"] -> {"service": "foo", "host": "bar"}
func makeLabels(lvs []string) promclient.Labels {
	labels := promclient.Labels{}
	for i := 0; i < len(lvs); i += 2 {
		labels[lvs[i]] = lvs[i+1]
	}
	return labels
}

// extractLabelValues extracts only the values from alternating key-value pairs.
// e.g., ["service", "foo", "host", "bar"] -> ["foo", "bar"]
func extractLabelValues(lvs []string) []string {
	values := make([]string, 0, len(lvs)/2)
	for i := 1; i < len(lvs); i += 2 {
		values = append(values, lvs[i])
	}
	return values
}

// promCounter wraps a Prometheus CounterVec and supports deletion of label values.
type promCounter struct {
	cv  *promclient.CounterVec
	lvs []string // alternating key-value pairs
}

func (c *promCounter) Add(delta float64) {
	c.cv.With(makeLabels(c.lvs)).Add(delta)
}

func (c *promCounter) With(labelValues ...string) gkm.Counter {
	return &promCounter{
		cv:  c.cv,
		lvs: append(append([]string{}, c.lvs...), labelValues...),
	}
}

// DeleteLabelValues removes the metric with the given label values (values only, not key-value pairs).
func (c *promCounter) DeleteLabelValues(labelValues ...string) bool {
	return c.cv.DeleteLabelValues(labelValues...)
}

// promGauge wraps a Prometheus GaugeVec and supports deletion of label values.
type promGauge struct {
	gv  *promclient.GaugeVec
	lvs []string // alternating key-value pairs
}

func (g *promGauge) Set(value float64) {
	g.gv.With(makeLabels(g.lvs)).Set(value)
}

func (g *promGauge) Add(delta float64) {
	g.gv.With(makeLabels(g.lvs)).Add(delta)
}

func (g *promGauge) With(labelValues ...string) gkm.Gauge {
	return &promGauge{
		gv:  g.gv,
		lvs: append(append([]string{}, g.lvs...), labelValues...),
	}
}

// DeleteLabelValues removes the metric with the given label values (values only, not key-value pairs).
func (g *promGauge) DeleteLabelValues(labelValues ...string) bool {
	return g.gv.DeleteLabelValues(labelValues...)
}

// promHistogram wraps a Prometheus HistogramVec and supports deletion of label values.
type promHistogram struct {
	hv  *promclient.HistogramVec
	lvs []string // alternating key-value pairs
}

func (h *promHistogram) Observe(value float64) {
	h.hv.With(makeLabels(h.lvs)).Observe(value)
}

func (h *promHistogram) With(labelValues ...string) gkm.Histogram {
	return &promHistogram{
		hv:  h.hv,
		lvs: append(append([]string{}, h.lvs...), labelValues...),
	}
}

// DeleteLabelValues removes the metric with the given label values (values only, not key-value pairs).
func (h *promHistogram) DeleteLabelValues(labelValues ...string) bool {
	return h.hv.DeleteLabelValues(labelValues...)
}
