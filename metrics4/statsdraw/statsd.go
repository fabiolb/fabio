package statsdraw

//import (
//	"time"
//
//	"github.com/alexcesaro/statsd"
//	"github.com/fabiolb/fabio/metrics4"
//	"github.com/fabiolb/fabio/metrics4/names"
//)
//
//type Provider struct {
//	c *statsd.Client
//}
//
//func NewProvider(prefix, addr string, interval time.Duration) (*Provider, error) {
//	opts := []statsd.Option{
//		statsd.Address(addr),
//		statsd.FlushPeriod(interval),
//	}
//	if prefix != "" {
//		opts = append(opts, statsd.Prefix(prefix))
//	}
//
//	c, err := statsd.New(opts...)
//	if err != nil {
//		return nil, err
//	}
//	return &Provider{c}, nil
//}
//
//func (p *Provider) NewCounter(name string, labels ...string) metrics4.Counter {
//	return &Counter{c: p.c, name: name, labels: labels}
//}
//
//func (p *Provider) NewGauge(name string, labels ...string) metrics4.Gauge {
//	return &Gauge{c: p.c, name: name, labels: labels}
//}
//
//func (p *Provider) NewTimer(name string, labels ...string) metrics4.Timer {
//	return &Timer{c: p.c, name: name, labels: labels}
//}
//
//func (p *Provider) GetMetrics() []*metrics4.Metric {
//	return make([]*metrics4.Metric, 0)
//}
//
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
