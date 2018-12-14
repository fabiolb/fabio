package statsd

import (
	"github.com/go-kit/kit/log"
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
	if len(labelsNames) == 0 {
		// TODO: Move 'sampleRate' of StatsD to Config
		return p.client.NewCounter(name, 1)
	}
	return metrics4.NewUntaggedCounter(p, name, labelsNames)
}

func (p *Provider) NewGauge(name string, labelsNames ...string) metrics4.Gauge {
	if len(labelsNames) == 0 {
		return p.client.NewGauge(name)
	}
	return metrics4.NewUntaggedGauge(p, name, labelsNames)
}

func (p *Provider) NewTimer(name string, labelsNames ...string) metrics4.Timer {
	if len(labelsNames) == 0 {
		// TODO: Move 'sampleRate' of StatsD to Config
		return p.client.NewTiming(name, 1)
	}
	return metrics4.NewUntaggedTimer(p, name, labelsNames)
}

func (p *Provider) Close() error {
	p.ticker.Stop()
	return nil
}
