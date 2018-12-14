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
	client := statsd.New(metrics4.FabioNamespace, log.NewNopLogger())

	ticker := time.NewTicker(interval)

	go client.SendLoop(ticker.C, "udp", addr)

	return &Provider{client, ticker}, nil
}

func (p *Provider) NewCounter(name string, labels ...string) metrics4.Counter {
	// TODO: Move 'sampleRate' of StatsD to Config
	return p.client.NewCounter(name, 1)
}

func (p *Provider) NewGauge(name string, labels ...string) metrics4.Gauge {
	return p.client.NewGauge(name)
}

func (p *Provider) NewTimer(name string, labels ...string) metrics4.Timer {
	// TODO: Move 'sampleRate' of StatsD to Config
	return p.client.NewTiming(name, 1)
}

func (p *Provider) Close() error {
	p.ticker.Stop()
	return nil
}
