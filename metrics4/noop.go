package metrics4

import (
	"time"

	"github.com/go-kit/kit/metrics"
)

var noopCounter = NoopCounter{}

type NoopCounter struct{}

func (c *NoopCounter) Add(float64) {}

func (c *NoopCounter) With(labels ... string) metrics.Counter {
	return c
}

var noopTimer = NoopTimer{}

type NoopTimer struct{}

func (t *NoopTimer) Observe(time.Duration) {}

func (t *NoopTimer) Start() {}

func (t *NoopTimer) Stop() {}

func (t *NoopTimer) With(labels ... string) Timer {
	return t
}

var noopGauge = NoopGauge{}

type NoopGauge struct{}

func (g *NoopGauge) Add(float64) {}

func (g *NoopGauge) Set(float64) {}

func (g *NoopGauge) With(... string) metrics.Gauge {
	return g
}
