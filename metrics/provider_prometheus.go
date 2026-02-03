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

// promCounter wraps a Prometheus CounterVec and supports deletion of label values.
type promCounter struct {
	cv  *promclient.CounterVec
	lvs []string
}

func (c *promCounter) Add(delta float64) {
	c.cv.WithLabelValues(c.lvs...).Add(delta)
}

func (c *promCounter) With(labelValues ...string) gkm.Counter {
	return &promCounter{
		cv:  c.cv,
		lvs: append(c.lvs, labelValues...),
	}
}

func (c *promCounter) DeleteLabelValues(labelValues ...string) bool {
	return c.cv.DeleteLabelValues(labelValues...)
}

// promGauge wraps a Prometheus GaugeVec and supports deletion of label values.
type promGauge struct {
	gv  *promclient.GaugeVec
	lvs []string
}

func (g *promGauge) Set(value float64) {
	g.gv.WithLabelValues(g.lvs...).Set(value)
}

func (g *promGauge) Add(delta float64) {
	g.gv.WithLabelValues(g.lvs...).Add(delta)
}

func (g *promGauge) With(labelValues ...string) gkm.Gauge {
	return &promGauge{
		gv:  g.gv,
		lvs: append(g.lvs, labelValues...),
	}
}

func (g *promGauge) DeleteLabelValues(labelValues ...string) bool {
	return g.gv.DeleteLabelValues(labelValues...)
}

// promHistogram wraps a Prometheus HistogramVec and supports deletion of label values.
type promHistogram struct {
	hv  *promclient.HistogramVec
	lvs []string
}

func (h *promHistogram) Observe(value float64) {
	h.hv.WithLabelValues(h.lvs...).Observe(value)
}

func (h *promHistogram) With(labelValues ...string) gkm.Histogram {
	return &promHistogram{
		hv:  h.hv,
		lvs: append(h.lvs, labelValues...),
	}
}

func (h *promHistogram) DeleteLabelValues(labelValues ...string) bool {
	return h.hv.DeleteLabelValues(labelValues...)
}
