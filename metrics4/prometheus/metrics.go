package prometheus

import (
	"github.com/fabiolb/fabio/metrics4"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"strings"
	"sync"
)

type Provider struct {
	counters map[string]metrics4.Counter
	gauges   map[string]metrics4.Gauge
	timers   map[string]metrics4.Timer
	mutex    sync.Mutex
	prefix   string
}

func normalizeName(name string) string {
	name = strings.Replace(name, ".", "_", -1)
	name = strings.Replace(name, "-", "_", -1)
	return name
}

func NewProvider(prefix string) metrics4.Provider {
	return &Provider{
		counters: make(map[string]metrics4.Counter),
		gauges:   make(map[string]metrics4.Gauge),
		timers:   make(map[string]metrics4.Timer),
		prefix:   prefix,
	}
}

func (p *Provider) NewCounter(name string, labels ... string) metrics4.Counter {
	name = normalizeName(name)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.counters[name] == nil {
		p.counters[name] = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: normalizeName(p.prefix),
			Name:      name,
		}, labels)
	}

	return p.counters[name]
}

func (p *Provider) NewGauge(name string, labels ... string) metrics4.Gauge {
	name = normalizeName(name)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.gauges[name] == nil {
		p.gauges[name] = prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: normalizeName(p.prefix),
			Name:      name,
		}, labels)
	}

	return p.gauges[name]
}

func (p *Provider) NewTimer(name string, labels ... string) metrics4.Timer {
	name = normalizeName(name)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.timers[name] == nil {
		p.timers[name] = prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Namespace: normalizeName(p.prefix),
			Name:      name,
		}, labels)
	}

	return p.timers[name]
}

func (p *Provider) Close() error {
	return nil
}
