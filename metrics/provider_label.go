package metrics

import (
	"fmt"
	gkm "github.com/go-kit/kit/metrics"
	"math"
	"strings"
	"sync/atomic"
)

type labelProvider struct {
	prefix string
}

func (p *labelProvider) NewCounter(name string, labels ...string) gkm.Counter {
	return &labelCounter{Name: strings.Join([]string{p.prefix, name}, DotSeparator), Labels: labels}
}

func (p *labelProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	return &labelGauge{Name: strings.Join([]string{p.prefix, name}, DotSeparator), Labels: labels}
}

func (p *labelProvider) NewHistogram(name string, labels ...string) gkm.Histogram {
	return &labelHistogram{Name: strings.Join([]string{p.prefix, name}, DotSeparator), Labels: labels}
}

type labelCounter struct {
	Name   string
	Labels []string
	Values []string
	v      int64
}

func (c *labelCounter) With(labelValues ...string) gkm.Counter {
	cc := &labelCounter{
		Name:   c.Name,
		Labels: c.Labels,
		Values: make([]string, len(labelValues)),
		v:      c.v,
	}
	copy(cc.Values, labelValues)
	return cc
}

func (c *labelCounter) Inc() {
	v := atomic.AddInt64(&c.v, 1)
	fmt.Printf("%s:%d|c%s\n", c.Name, v, Labels(c.Labels, c.Values, "|#", ":", ","))
}

func (c *labelCounter) Add(delta float64) {
	v := atomic.AddInt64(&c.v, int64(delta))
	fmt.Printf("%s:%d|c%s\n", c.Name, v, Labels(c.Labels, c.Values, "|#", ":", ","))
}

type labelGauge struct {
	valBits uint64
	Name    string
	Labels  []string
	Values  []string
}

func (g *labelGauge) With(labelValues ...string) gkm.Gauge {
	gc := &labelGauge{
		Name:   g.Name,
		Labels: g.Labels,
		Values: make([]string, len(labelValues)),
	}
	copy(gc.Values, labelValues)
	return gc
}

func (g *labelGauge) Set(n float64) {
	atomic.StoreUint64(&g.valBits, math.Float64bits(n))
	fmt.Printf("%s:%d|g%s\n", g.Name, int(n), Labels(g.Labels, g.Values, "|#", ":", ","))
}

func (g *labelGauge) Add(delta float64) {
	var oldBits uint64
	var newBits uint64
	for {
		oldBits = atomic.LoadUint64(&g.valBits)
		newBits = math.Float64bits(math.Float64frombits(oldBits) + delta)
		if atomic.CompareAndSwapUint64(&g.valBits, oldBits, newBits) {
			break
		}
	}
	fmt.Printf("%s:%d|g%s\n", g.Name, int(delta), Labels(g.Labels, g.Values, "|#", ":", ","))
}

type labelHistogram struct {
	Name   string
	Labels []string
	Values []string
}

func (h *labelHistogram) With(labels ...string) gkm.Histogram {
	h2 := &labelHistogram{}
	*h2 = *h
	h2.Values = make([]string, len(labels))
	copy(h2.Values, labels)
	return h2
}

func (h *labelHistogram) Observe(t float64) {
	fmt.Printf("%s:%d|ms%s\n", h.Name, int64(math.Round(t*1000.0)), Labels(h.Labels, h.Values, "|#", ":", ","))
}
