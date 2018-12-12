package prometheus

import (
	"github.com/fabiolb/fabio/metrics4"

	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type Provider struct {
	counters map[string]metrics4.Counter
	gauges map[string]metrics4.Gauge
}

func NewProvider() *Provider {
	return &Provider{
		make(map[string]metrics4.Counter),
		make(map[string]metrics4.Gauge),
	}
}

func (p *Provider) NewCounter(name string) metrics4.Counter {
	// TODO(max): Add lock ?
	if p.counters[name] == nil {
		p.counters[name] = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: metrics4.FabioNamespace,
			Subsystem: "",
			Name:      name,
			Help:      "",
		}, []string{})
	}

	return p.counters[name]
}

func (p *Provider) NewGauge(name string) metrics4.Gauge {
	// TODO(max): Add lock ?
	if p.gauges[name] == nil {
		p.gauges[name] = prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: metrics4.FabioNamespace,
			Subsystem: "",
			Name:      name,
			Help:      "",
		}, []string{})
	}

	return p.gauges[name]
}
