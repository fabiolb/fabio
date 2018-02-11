package metrics4

import (
	"time"
)

// Provider is an abstraction of a metrics backend.
type Provider interface {
	// NewCounter creates a new counter object.
	NewCounter(name string, labels ...string) Counter

	// NewGauge creates a new gauge object.
	NewGauge(name string, labels ...string) Gauge

	// NewTimer creates a new timer object.
	NewTimer(name string, labels ...string) Timer

	// Unregister removes a previously registered
	// name or metric. Required for go-metrics and
	// service pruning. This signature is probably not
	// correct.
	Unregister(v interface{})
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
func (mp *MultiProvider) NewCounter(name string, labels ...string) Counter {
	var c []Counter
	for _, p := range mp.p {
		c = append(c, p.NewCounter(name, labels...))
	}
	return &MultiCounter{c}
}

// NewGauge creates a MultiGauge with gauge objects for all registered
// providers.
func (mp *MultiProvider) NewGauge(name string, labels ...string) Gauge {
	var v []Gauge
	for _, p := range mp.p {
		v = append(v, p.NewGauge(name, labels...))
	}
	return &MultiGauge{v}
}

// NewTimer creates a MultiTimer with timer objects for all registered
// providers.
func (mp *MultiProvider) NewTimer(name string, labels ...string) Timer {
	var t []Timer
	for _, p := range mp.p {
		t = append(t, p.NewTimer(name, labels...))
	}
	return &MultiTimer{t}
}

// Unregister removes the metric object from all registered providers.
func (mp *MultiProvider) Unregister(v interface{}) {
	for _, p := range mp.p {
		p.Unregister(v)
	}
}

// Count measures a number.
type Counter interface {
	Count(int)
}

// MultiCounter wraps zero or more counters.
type MultiCounter struct {
	c []Counter
}

func (mc *MultiCounter) Count(n int) {
	for _, c := range mc.c {
		c.Count(n)
	}
}

// Gauge measures a value.
type Gauge interface {
	Update(int)
}

// MultiGauge wraps zero or more gauges.
type MultiGauge struct {
	v []Gauge
}

func (m *MultiGauge) Update(n int) {
	for _, v := range m.v {
		v.Update(n)
	}
}

// Timer measures the time of an event.
type Timer interface {
	Update(time.Duration)
}

// MultTimer wraps zero or more timers.
type MultiTimer struct {
	t []Timer
}

func (mt *MultiTimer) Update(d time.Duration) {
	for _, t := range mt.t {
		t.Update(d)
	}
}
