package statsd

import (
	"github.com/go-kit/kit/log"
	"strings"
	"time"

	"github.com/fabiolb/fabio/metrics4"
	"github.com/go-kit/kit/metrics/statsd"
)

type Provider struct {
	client *statsd.Statsd
	ticker *time.Ticker
}

func NewProvider(addr string, interval time.Duration) (metrics4.Provider, error) {
	client := statsd.New(metrics4.FabioNamespace + "_", log.NewNopLogger())

	ticker := time.NewTicker(interval)

	go client.SendLoop(ticker.C, "udp", addr)

	return &Provider{client, ticker}, nil
}

func (p *Provider) NewCounter(name string, labelsNames ...string) metrics4.Counter {
	var counter metrics4.Counter
	if len(labelsNames) == 0 {
		// TODO: Move 'sampleRate' of StatsD to Config
		counter = p.client.NewCounter(name, 1)
	}
	return &Counter{
		c: counter,
		m: &metric{
			p: p,
			name: name,
			labelsNames: labelsNames,
		},
	}
}

func (p *Provider) NewGauge(name string, labelsNames ...string) metrics4.Gauge {
	var gauge metrics4.Gauge
	if len(labelsNames) == 0 {
		// TODO: Move 'sampleRate' of StatsD to Config
		gauge = p.client.NewGauge(name)
	}
	return &Gauge{
		g: gauge,
		m: &metric{
			p: p,
			name: name,
			labelsNames: labelsNames,
		},
	}
}

func (p *Provider) NewTimer(name string, labelsNames ...string) metrics4.Timer {
	var timer metrics4.Timer
	if len(labelsNames) == 0 {
		// TODO: Move 'sampleRate' of StatsD to Config
		timer = p.client.NewTiming(name, 1)
	}
	return &Timer{
		t: timer,
		m: &metric{
			p: p,
			name: name,
			labelsNames: labelsNames,
		},
	}
}

func (p *Provider) Close() error {
	p.ticker.Stop()
	return nil
}

//func parseLabelsValues(labelsNames []string, labels []string) ([]string, error) {
//	labelsCount := len(labelsNames)
//	labelsValues := make([]string, labelsCount)
//
//	for i := 1; i <= labelsCount; i++ {
//		if labelsNames[i - 1] != labelsNames[(i * 2) - 1] {
//			return nil, errors.New("incorrect label name")
//		}
//
//		labelsValues[i] = labels[(i * 2) - 1]
//	}
//
//	return labelsValues, nil
//}

func makeNameFromLabels(labels []string) string {
	return strings.Join(labels, "_")
}

type metric struct {
	name string
	labelsNames []string
	p *Provider
}

type Counter struct {
	c metrics4.Counter
	m *metric
}

func (c *Counter) Add(delta float64) {
	if c.c != nil {
		c.c.Add(delta)
	}
}

func (c *Counter) With(labels ... string) metrics4.Counter {
	// TODO(max): Check labels
	return c.m.p.NewCounter(c.m.name + "_" + makeNameFromLabels(labels))
}

type Timer struct {
	t metrics4.Timer
	m *metric
}

func (t *Timer) Observe(value float64) {
	if t.t != nil {
		t.t.Observe(value)
	}
}

func (t *Timer) With(labels ... string) metrics4.Timer {
	// TODO(max): Check labels
	return t.m.p.NewTimer(t.m.name + "_" + makeNameFromLabels(labels))
}

type Gauge struct {
	g metrics4.Gauge
	m *metric
}

func (g *Gauge) Add(value float64) {
	if g.g != nil {
		g.g.Add(value)
	}
}

func (g *Gauge) Set(value float64) {
	if g.g != nil {
		g.g.Set(value)
	}
}

func (g *Gauge) With(labels ... string) metrics4.Gauge {
	// TODO(max): Check labels
	return g.m.p.NewGauge(g.m.name + "_" + makeNameFromLabels(labels))
}
