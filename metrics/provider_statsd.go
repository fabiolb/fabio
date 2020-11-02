package metrics

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-kit/kit/log"
	gkm "github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/statsd"
)

type StatsdProvider struct {
	S *statsd.Statsd
}

func NewStatsdProvider(prefix, addr string, interval time.Duration) (*StatsdProvider, error) {
	p := &StatsdProvider{
		S: statsd.New(prefix, log.NewNopLogger()),
	}
	_, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("error resolving statsd address %s: %w", addr, err)
	}
	t := time.NewTicker(interval)
	go func() {
		p.S.SendLoop(context.Background(), t.C, "udp", addr)
	}()

	return p, nil
}

// NewCounter - This assumes if there are labels, there will be a With() call
func (p *StatsdProvider) NewCounter(name string, labels ...string) gkm.Counter {
	if len(labels) == 0 {
		return p.S.NewCounter(name, 1)
	}
	return &statsdCounter{
		name:        name,
		p:           p,
		routeMetric: isRouteMetric(name),
	}
}

// NewGauge - this assumes if there are labels, there will be a With() call.
func (p *StatsdProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	if len(labels) == 0 {
		return p.S.NewGauge(name)
	}
	return &statsdGauge{
		name:        name,
		p:           p,
		routeMetric: isRouteMetric(name),
	}
}

// NewHistogram - this assumes if there are labels, there will be a With() call.
func (p *StatsdProvider) NewHistogram(name string, labels ...string) gkm.Histogram {
	var histogram gkm.Histogram
	if len(labels) == 0 {
		histogram = p.S.NewTiming(name, 1)
	}
	return &statsdHistogram{
		Histogram:   histogram,
		name:        name,
		p:           p,
		routeMetric: isRouteMetric(name),
	}
}

type statsdCounter struct {
	gkm.Counter
	name        string
	p           *StatsdProvider
	routeMetric bool
}

func (c *statsdCounter) With(labelValues ...string) gkm.Counter {
	var name string
	switch c.routeMetric {
	case true:
		var err error
		name, err = TargetNameWith(c.name, labelValues)
		if err != nil {
			panic(err)
		}
	case false:
		name = Flatten(c.name, labelValues, DotSeparator)
	}
	return &statsdCounter{
		Counter:     c.p.S.NewCounter(name, 1),
		name:        name,
		p:           c.p,
		routeMetric: c.routeMetric,
	}
}

type statsdGauge struct {
	gkm.Gauge
	name        string
	p           *StatsdProvider
	routeMetric bool
}

func (g *statsdGauge) With(labelValues ...string) gkm.Gauge {
	var name string
	switch g.routeMetric {
	case true:
		var err error
		name, err = TargetNameWith(g.name, labelValues)
		if err != nil {
			panic(err)
		}
	case false:
		name = Flatten(g.name, labelValues, DotSeparator)
	}
	return &statsdGauge{
		Gauge:       g.p.S.NewGauge(name),
		name:        name,
		p:           g.p,
		routeMetric: g.routeMetric,
	}
}

type statsdHistogram struct {
	gkm.Histogram
	name        string
	p           *StatsdProvider
	routeMetric bool
}

func (h *statsdHistogram) With(labelValues ...string) gkm.Histogram {
	var name string
	switch h.routeMetric {
	case true:
		var err error
		name, err = TargetNameWith(h.name, labelValues)
		if err != nil {
			panic(err)
		}
	case false:
		name = Flatten(h.name, labelValues, DotSeparator)
	}
	return &statsdHistogram{
		Histogram:   h.p.S.NewTiming(name, 1),
		name:        name,
		p:           h.p,
		routeMetric: h.routeMetric,
	}
}

func (h *statsdHistogram) Observe(value float64) {
	h.Histogram.Observe(value * 1000.0)
}
