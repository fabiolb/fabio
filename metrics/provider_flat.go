package metrics

import (
	"fmt"
	"math"
	"strings"
	"sync/atomic"

	gkm "github.com/go-kit/kit/metrics"
)

type flatProvider struct {
	prefix string
}

func (p *flatProvider) NewCounter(name string, labels ...string) gkm.Counter {
	return &flatCounter{Name: Flatten(strings.Join([]string{p.prefix, name}, DotSeparator), labels, DotSeparator)}
}

func (p *flatProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	return &flatGauge{Name: Flatten(strings.Join([]string{p.prefix, name}, DotSeparator), labels, DotSeparator)}
}

func (p *flatProvider) NewHistogram(name string, labels ...string) gkm.Histogram {
	return &flatHistogram{Name: Flatten(strings.Join([]string{p.prefix, name}, DotSeparator), labels, DotSeparator)}
}

type flatCounter struct {
	Name string
	v    uint64
}

func (c *flatCounter) With(labelValues ...string) gkm.Counter {
	return c
}

func (c *flatCounter) Add(v float64) {
	uv := atomic.AddUint64(&c.v, uint64(v))
	fmt.Printf("%s:%d|c\n", c.Name, uv)
}

type flatGauge struct {
	// Stolen from prometheus client gauge
	valBits uint64

	Name string
}

func (g *flatGauge) Set(n float64) {
	atomic.StoreUint64(&g.valBits, math.Float64bits(n))
	fmt.Printf("%s:%d|g\n", g.Name, int(n))
}

func (g *flatGauge) Add(delta float64) {
	var oldBits uint64
	var newBits uint64
	for {
		oldBits = atomic.LoadUint64(&g.valBits)
		newBits = math.Float64bits(math.Float64frombits(oldBits) + delta)
		if atomic.CompareAndSwapUint64(&g.valBits, oldBits, newBits) {
			break
		}
	}
	fmt.Printf("%s:%d|g\n", g.Name, int(math.Float64frombits(newBits)))
}

func (g *flatGauge) With(labelValues ...string) gkm.Gauge {
	return g
}

type flatHistogram struct {
	Name string
}

func (h *flatHistogram) Observe(t float64) {
	fmt.Printf(":%s:%d|ms\n", h.Name, int64(math.Round(t*1000.0)))
}
func (h *flatHistogram) With(labels ...string) gkm.Histogram {
	return h
}
