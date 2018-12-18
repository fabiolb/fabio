package metrics4

import (
	"github.com/go-kit/kit/metrics"
	"io"
)

const FabioNamespace = "fabio"

type Counter = metrics.Counter

type Gauge = metrics.Gauge

type Timer = metrics.Histogram

// Provider is an abstraction of a metrics backend.
type Provider interface {
	// NewCounter creates a new counter object.
	// labels - array of labels names
	NewCounter(name string, labelsNames ... string) Counter

	// NewGauge creates a new gauge object.
	NewGauge(name string, labelsNames ... string) Gauge

	// NewTimer creates a new timer object.
	NewTimer(name string, labelsNames ... string) Timer

	// It extends Provider with Close method which closes a disposable objects that are connected with a provider.
	io.Closer
}

// MultiProvider wraps zero or more providers.
type MultiProvider struct {
	p      []Provider
}

func NewMultiProvider(p []Provider) *MultiProvider {
	return &MultiProvider{p}
}

// NewCounter creates a MultiCounter with counter objects for all registered
// providers.
func (mp *MultiProvider) NewCounter(name string, labels ... string) Counter {
	c := make([]Counter, len(mp.p))
	for i, p := range mp.p {
		c[i] = p.NewCounter(name, labels...)
	}
	return &MultiCounter{c}
}

// NewGauge creates a MultiGauge with gauge objects for all registered
// providers.
func (mp *MultiProvider) NewGauge(name string, labels ... string) Gauge {
	g := make([]Gauge, len(mp.p))
	for i, p := range mp.p {
		g[i] = p.NewGauge(name, labels...)
	}
	return &MultiGauge{g}
}

// NewTimer creates a MultiTimer with timer objects for all registered
// providers.
func (mp *MultiProvider) NewTimer(name string, labels ... string) Timer {
	t := make([]Timer, len(mp.p))
	for i, p := range mp.p {
		t[i] = p.NewTimer(name, labels...)
	}
	return &MultiTimer{t}
}

func (mp *MultiProvider) Close() error {
	var errors []error
	for _, p := range mp.p {
		e := p.Close()
		if e != nil {
			errors = append(errors, e)
		}
	}
	if len(errors) > 0 {
		// TODO(max): Define MultiError
		return errors[0]
	}
	return nil
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
