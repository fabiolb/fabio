package prometheus

import (
	"github.com/fabiolb/fabio/metrics4"

	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type Provider struct {
	counters map[string]metrics4.Counter
}

func NewProvider() *Provider {
	return &Provider{make(map[string]metrics4.Counter)}
}

func (p *Provider) NewCounter(name string) metrics4.Counter {
	// TODO(max): Add lock ?
	if p.counters[name] == nil {
		p.counters[name] = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: metrics4.FABIO_NAMESPACE,
			Subsystem: "",
			Name:      name,
			Help:      "",
		}, []string{})
	}

	return p.counters[name]
}
