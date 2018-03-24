package label

import (
	"fmt"
	"time"

	"github.com/fabiolb/fabio/metrics4"
	"github.com/fabiolb/fabio/metrics4/names"
)

type Provider struct{}

func (p *Provider) NewCounter(name string, labels ...string) metrics4.Counter {
	return &Counter{Name: name, Labels: labels}
}

func (p *Provider) NewGauge(name string, labels ...string) metrics4.Gauge {
	return &Gauge{Name: name, Labels: labels}
}

func (p *Provider) NewTimer(name string, labels ...string) metrics4.Timer {
	return &Timer{Name: name, Labels: labels}
}

func (p *Provider) Unregister(interface{}) {}

type Counter struct {
	Name   string
	Labels []string
}

func (c *Counter) Count(n int) {
	fmt.Printf("%s:%d|c%s\n", c.Name, n, names.Labels(c.Labels, "|#", ":", ","))
}

type Gauge struct {
	Name   string
	Labels []string
}

func (g *Gauge) Update(n int) {
	fmt.Printf("%s:%d|g%s\n", g.Name, n, names.Labels(g.Labels, "|#", ":", ","))
}

type Timer struct {
	Name   string
	Labels []string
}

func (t *Timer) Update(d time.Duration) {
	fmt.Printf("%s:%d|ns%s\n", t.Name, d.Nanoseconds(), names.Labels(t.Labels, "|#", ":", ","))
}
