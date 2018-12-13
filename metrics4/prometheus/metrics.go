package prometheus

import (
	"github.com/fabiolb/fabio/metrics4"
	"time"

	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type Provider struct {
	counters   map[string]metrics4.Counter
	gauges     map[string]metrics4.Gauge
	timers     map[string]metrics4.Timer
	histograms map[string]metrics4.Histogram
}

func NewProvider() *Provider {
	return &Provider{
		make(map[string]metrics4.Counter),
		make(map[string]metrics4.Gauge),
		make(map[string]metrics4.Timer),
		make(map[string]metrics4.Histogram),
	}
}

func (p *Provider) NewCounter(name string, labels... string) metrics4.Counter {
	// TODO(max): Add lock ?
	if p.counters[name] == nil {
		p.counters[name] = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: metrics4.FabioNamespace,
			Subsystem: "",
			Name:      name,
			Help:      "",
		}, labels)
	}

	return p.counters[name]
}

func (p *Provider) NewGauge(name string, labels... string) metrics4.Gauge {
	// TODO(max): Add lock ?
	if p.gauges[name] == nil {
		p.gauges[name] = prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: metrics4.FabioNamespace,
			Subsystem: "",
			Name:      name,
			Help:      "",
		}, labels)
	}

	return p.gauges[name]
}

func (p *Provider) NewHistogram(name string, labels... string) metrics4.Histogram {
	// TODO(max): Add lock ?
	if p.histograms[name] == nil {
		p.histograms[name] = prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Namespace: metrics4.FabioNamespace,
			Subsystem: "",
			Name:      name,
			Help:      "",
			// TODO: Look on 'Buckets'
		}, labels)
	}

	return p.histograms[name]
}

func (p *Provider) NewTimer(name string, labels... string) metrics4.Timer {
	if p.timers[name] == nil {
		h := prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Namespace: metrics4.FabioNamespace,
			Name:      name,
		}, labels)

		p.timers[name] = metrics4.NewTimerStruct(h, time.Now())
	}

	return p.timers[name]
}
