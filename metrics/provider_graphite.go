package metrics

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	gkm "github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/graphite"
	"net"
	"time"
)

type GraphiteProvider struct {
	G       *graphite.Graphite
	buckets int
}

func (g *GraphiteProvider) NewCounter(name string, labels ...string) gkm.Counter {
	if len(labels) == 0 {
		return g.G.NewCounter(name)
	}
	return &graphiteCounter{
		p:           g,
		name:        name,
		routeMetric: isRouteMetric(name),
	}
}

func (g *GraphiteProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	if len(labels) == 0 {
		return g.G.NewGauge(name)
	}
	return &graphiteGauge{
		p:           g,
		name:        name,
		routeMetric: isRouteMetric(name),
	}
}

func (g *GraphiteProvider) NewHistogram(name string, labels ...string) gkm.Histogram {
	var histogram gkm.Histogram
	if len(labels) == 0 {
		histogram = g.G.NewHistogram(name, g.buckets)
	}
	return &graphiteHistogram{
		Histogram:   histogram,
		p:           g,
		name:        name,
		routeMetric: isRouteMetric(name),
	}
}

func NewGraphiteProvider(prefix, addr string, buckets int, interval time.Duration) (*GraphiteProvider, error) {
	g := &GraphiteProvider{
		G:       graphite.New(prefix, log.NewNopLogger()),
		buckets: buckets,
	}
	_, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("error resolving graphite address %s: %w", addr, err)
	}
	t := time.NewTicker(interval)
	go func() {
		g.G.SendLoop(context.Background(), t.C, "tcp", addr)
	}()
	return g, nil
}

type graphiteCounter struct {
	gkm.Counter
	p           *GraphiteProvider
	name        string
	routeMetric bool
}

func (c *graphiteCounter) With(labelValues ...string) gkm.Counter {
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
	return &graphiteCounter{
		Counter:     c.p.G.NewCounter(name),
		name:        name,
		p:           c.p,
		routeMetric: c.routeMetric,
	}
}

type graphiteGauge struct {
	gkm.Gauge
	p           *GraphiteProvider
	name        string
	routeMetric bool
}

func (g *graphiteGauge) With(labelValues ...string) gkm.Gauge {
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
	return &graphiteGauge{
		Gauge:       g.p.G.NewGauge(name),
		name:        name,
		p:           g.p,
		routeMetric: g.routeMetric,
	}
}

type graphiteHistogram struct {
	gkm.Histogram
	p           *GraphiteProvider
	name        string
	routeMetric bool
}

func (h *graphiteHistogram) With(labelValues ...string) gkm.Histogram {
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
	return &graphiteHistogram{
		Histogram:   h.p.G.NewHistogram(name, h.p.buckets),
		name:        name,
		p:           h.p,
		routeMetric: h.routeMetric,
	}
}

func (h *graphiteHistogram) Observe(value float64) {
	h.Histogram.Observe(value * 1000.0)
}
