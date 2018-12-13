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
	return &metrics4.NoopCounter{}
}

func (p *Provider) NewGauge(name string, labels ...string) metrics4.Gauge {
	return p.client.NewGauge(name)
}

func (p *Provider) NewTimer(name string, labels ...string) metrics4.Timer {
	return &metrics4.NoopTimer{}
}

//func (p *Provider) NewHistogram(name string, labels ...string) metrics4.Histogram {
//	return &metrics4.NoopHistogram{}
//}

// TODO:(max): Add kinda destructor

//func (p *Provider) Unregister(interface{}) {}
//
//type Counter struct {
//	c      *statsd.Client
//	name   string
//	labels []string
//}
//
//func (v *Counter) Count(n int) {
//	v.c.Count(names.Flatten(v.name, v.labels, names.DotSeparator), n)
//}
//
//type Gauge struct {
//	c      *statsd.Client
//	name   string
//	labels []string
//}
//
//func (v *Gauge) Update(n int) {
//	v.c.Gauge(names.Flatten(v.name, v.labels, names.DotSeparator), n)
//}
//
//type Timer struct {
//	c      *statsd.Client
//	name   string
//	labels []string
//}
//
//func (v *Timer) Update(d time.Duration) {
//	v.c.Timing(names.Flatten(v.name, v.labels, names.DotSeparator), d)
//}
