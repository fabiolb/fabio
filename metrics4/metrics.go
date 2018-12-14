package metrics4

import (
	"github.com/go-kit/kit/metrics"
)

const FabioNamespace = "fabio"

type Counter metrics.Counter

type Gauge metrics.Gauge

type Timer = metrics.Histogram

//type Timer interface {
//	Observe(float64)
//	With(... string) Timer
//}

// Provider is an abstraction of a metrics backend.
type Provider interface {
	// NewCounter creates a new counter object.
	NewCounter(name string, labels ... string) Counter

	// NewGauge creates a new gauge object.
	NewGauge(name string, labels ... string) Gauge

	// NewTimer creates a new timer object.
	NewTimer(name string, labels ... string) Timer

	// Dispose()

	// Unregister removes a previously registered
	// name or metric. Required for go-metrics and
	// service pruning.
	// Unregister(name string)
}

// MultiProvider wraps zero or more providers.
type MultiProvider struct {
	p []Provider
}

func NewMultiProvider(p []Provider) *MultiProvider {
	return &MultiProvider{p}
}

// NewCounter creates a MultiCounter with counter objects for all registered
// providers.
func (mp *MultiProvider) NewCounter(name string, labels ... string) Counter {
	var c []Counter
	for _, p := range mp.p {
		c = append(c, p.NewCounter(name, labels...))
	}
	return &MultiCounter{c}
}

// NewGauge creates a MultiGauge with gauge objects for all registered
// providers.
func (mp *MultiProvider) NewGauge(name string, labels ... string) Gauge {
	var g []Gauge
	for _, p := range mp.p {
		g = append(g, p.NewGauge(name, labels...))
	}
	return &MultiGauge{g}
}

// NewTimer creates a MultiTimer with timer objects for all registered
// providers.
func (mp *MultiProvider) NewTimer(name string, labels ... string) Timer {
	var t []Timer
	for _, p := range mp.p {
		t = append(t, p.NewTimer(name, labels...))
	}
	return &MultiTimer{t}
}

// MultiCounter wraps zero or more counters.
type MultiCounter struct {
	counters []Counter
}

func (mc *MultiCounter) Add(delta float64) {
	for _, c := range mc.counters {
		c.Add(delta)
	}
}

func (mc *MultiCounter) With(labelValues ... string) metrics.Counter {
	labeledCounters := make([]Counter, len(mc.counters))
	for i, c := range mc.counters {
		labeledCounters[i] = c.With(labelValues...)
	}
	return &MultiCounter{labeledCounters}
}

// MultiGauge wraps zero or more gauges.
type MultiGauge struct {
	gauges []Gauge
}

func (mg *MultiGauge) Add(delta float64) {
	for _, g := range mg.gauges {
		g.Add(delta)
	}
}

func (mg *MultiGauge) Set(delta float64) {
	for _, g := range mg.gauges {
		g.Set(delta)
	}
}

func (mg *MultiGauge) With(labelValues ... string) metrics.Gauge {
	labeledGauges := make([]Gauge, len(mg.gauges))
	for i, g := range mg.gauges {
		labeledGauges[i] = g.With(labelValues...)
	}
	return &MultiGauge{labeledGauges}
}

// MultiTimer wraps zero or more timers.
type MultiTimer struct {
	timers []Timer
}

func (mt *MultiTimer) Observe(duration float64) {
	for _, t := range mt.timers {
		t.Observe(duration)
	}
}

func (mt *MultiTimer) With(labelValues ... string) Timer {
	labeledTimers := make([]Timer, len(mt.timers))
	for i, t := range mt.timers {
		labeledTimers[i] = t.With(labelValues...)
	}
	return &MultiTimer{labeledTimers}
}
