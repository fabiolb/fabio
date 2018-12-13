package statsd

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"time"

	"github.com/fabiolb/fabio/metrics4"
	"github.com/go-kit/kit/metrics/statsd"
)

type Provider struct {
	client *statsd.Statsd
	ticker *time.Ticker
}

func NewProvider(addr string, interval time.Duration) (metrics4.Provider, error) {
	client := statsd.New(metrics4.FabioNamespace, log.NewNopLogger())

	ticker := time.NewTicker(interval)

	go client.SendLoop(ticker.C, "udp", addr)

	return &Provider{client, ticker}, nil
}

func (p *Provider) NewCounter(name string, labels ...string) metrics4.Counter {
	return p.client.NewCounter(name, 1)
}

func (p *Provider) NewGauge(name string, labels ...string) metrics4.Gauge {
	return p.client.NewGauge(name)
}

func (p *Provider) NewTimer(name string, labels ...string) metrics4.Timer {
	return &Timer{
		timing: p.client.NewTiming(name, 1),
	}
}

type Timer struct {
	timing    *statsd.Timing
	histogram metrics.Histogram
}

func (t *Timer) Observe(value float64) {
	if t.timing != nil {
		t.timing.Observe(value)
	} else if t.histogram != nil {
		t.histogram.Observe(value)
	}
}

func (t *Timer) With(labelValues ... string) metrics4.Timer {
	if t.timing != nil {
		return &Timer{
			histogram: t.timing.With(labelValues...),
		}
	} else {
		return &Timer{
			histogram: t.histogram.With(labelValues...),
		}
	}
}
