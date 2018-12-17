package statsd

import (
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/metrics4/untagged"
	"github.com/go-kit/kit/log"
	"time"

	"github.com/fabiolb/fabio/metrics4"
	"github.com/go-kit/kit/metrics/statsd"
)

type Provider struct {
	client     *statsd.Statsd
	ticker     *time.Ticker
	sampleRate float64
}

func NewProvider(cfg config.StatsD) (metrics4.Provider, error) {
	client := statsd.New(metrics4.FabioNamespace+"_", log.NewNopLogger())

	ticker := time.NewTicker(cfg.Interval)

	go client.SendLoop(ticker.C, cfg.Protocol, cfg.Addr)

	return &Provider{client, ticker, cfg.SampleRate}, nil
}

func (p *Provider) NewCounter(name string, labelsNames ...string) metrics4.Counter {
	if len(labelsNames) == 0 {
		return p.client.NewCounter(name, p.sampleRate)
	}
	return untagged.NewCounter(p, name, labelsNames)
}

func (p *Provider) NewGauge(name string, labelsNames ...string) metrics4.Gauge {
	if len(labelsNames) == 0 {
		return p.client.NewGauge(name)
	}
	return untagged.NewGauge(p, name, labelsNames)
}

func (p *Provider) NewTimer(name string, labelsNames ...string) metrics4.Timer {
	if len(labelsNames) == 0 {
		return p.client.NewTiming(name, p.sampleRate)
	}
	return untagged.NewTimer(p, name, labelsNames)
}

func (p *Provider) Close() error {
	p.ticker.Stop()
	return nil
}