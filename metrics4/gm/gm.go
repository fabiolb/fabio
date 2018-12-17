package gm

import (
	"github.com/fabiolb/fabio/metrics4"
	"github.com/fabiolb/fabio/metrics4/untagged"
	gm "github.com/rcrowley/go-metrics"
	"time"
)

type provider struct {
	r gm.Registry
}

func (p *provider) NewCounter(name string, labelsNames ...string) metrics4.Counter {
	if len(labelsNames) == 0 {
		return  &counter{gm.GetOrRegisterCounter(name, p.r)}
	}
	return untagged.NewCounter(p, name, labelsNames)
}

func (p *provider) NewGauge(name string, labelsNames ...string) metrics4.Gauge {
	if len(labelsNames) == 0 {
		return &gauge{gm.GetOrRegisterGaugeFloat64(name, p.r)}
	}
	return untagged.NewGauge(p, name, labelsNames)
}

func (p *provider) NewTimer(name string, labelsNames ...string) metrics4.Timer {
	if len(labelsNames) == 0 {
		return &timer{gm.GetOrRegisterTimer(name, p.r)}
	}
	return untagged.NewTimer(p, name, labelsNames)
}

func (p *provider) Close() error {
	return nil
}

func NewProvider(r gm.Registry) metrics4.Provider {
	return &provider{r}
}

type counter struct {
	c gm.Counter
}

func (c *counter) Add(value float64) {
	c.c.Inc(int64(value))
}

func (c *counter) With(labels ... string) metrics4.Counter {
	return c
}

type gauge struct {
	g gm.GaugeFloat64
}

func (g *gauge) Add(value float64) {
	g.g.Update(g.g.Value() + value)
}

func (g *gauge) Set(value float64) {
	g.g.Update(value)
}

func (g *gauge) With(labels ... string) metrics4.Gauge {
	return g
}

func NewGauge(g gm.GaugeFloat64) metrics4.Gauge {
	return &gauge{g}
}

type timer struct {
	t gm.Timer
}

func (t *timer) Observe(value float64) {
	t.t.Update(time.Duration(value))
}

func (t *timer) With(labels ... string) metrics4.Timer {
	return t
}

func NewTimer(t gm.Timer) metrics4.Timer {
	return &timer{t}
}
