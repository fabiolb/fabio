package metrics4

import "github.com/go-kit/kit/metrics"

const FabioNamespace = "fabio"

type Counter metrics.Counter

type Gauge metrics.Gauge

type Timer metrics.Timer

// Provider is an abstraction of a metrics backend.
type Provider interface {
	// NewCounter creates a new counter object.
	NewCounter(name string) Counter

	// NewGauge creates a new gauge object.
	NewGauge(name string) Gauge

	// NewTimer creates a new timer object.
	//NewTimer(name string, labels ...string) Timer

	// Unregister removes a previously registered
	// name or metric. Required for go-metrics and
	// service pruning. This signature is probably not
	// correct.
	//Unregister(v interface{})
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
func (mp *MultiProvider) NewCounter(name string) Counter {
	var c []Counter
	for _, p := range mp.p {
		c = append(c, p.NewCounter(name))
	}
	return &MultiCounter{c}
}

// NewGauge creates a MultiGauge with gauge objects for all registered
// providers.
func (mp *MultiProvider) NewGauge(name string) Gauge {
	var g []Gauge
	for _, p := range mp.p {
		g = append(g, p.NewGauge(name))
	}
	return &MultiGauge{g}
}

// NewTimer creates a MultiTimer with timer objects for all registered
// providers.
//func (mp *MultiProvider) NewTimer(name string, labels ...string) Timer {
//	var t []Timer
//	for _, p := range mp.p {
//		t = append(t, p.NewTimer(name, labels...))
//	}
//	return &MultiTimer{t}
//}

// Unregister removes the metric object from all registered providers.
//func (mp *MultiProvider) Unregister(v interface{}) {
//	for _, p := range mp.p {
//		p.Unregister(v)
//	}
//}

// Count measures a number.
//type Counter interface {
//	Count(int)
//}

// MultiCounter wraps zero or more counters.
type MultiCounter struct {
	counters []Counter
}

func (mc *MultiCounter) Add(delta float64) {
	for _, c := range mc.counters {
		c.Add(delta)
	}
}

func (mc *MultiCounter) With(labelValues... string) metrics.Counter {
	for _, c := range mc.counters {
		c.With(labelValues...)
	}
	return mc
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

func (mg *MultiGauge) With(labelValues... string) metrics.Gauge {
	for _, g := range mg.gauges {
		g.With(labelValues...)
	}
	return mg
}

// Timer measures the time of an event.
//type Timer interface {
//	Update(time.Duration)
//}

// MultTimer wraps zero or more timers.
//type MultiTimer struct {
//	t []Timer
//}

//func (mt *MultiTimer) Update(d time.Duration) {
//	for _, t := range mt.t {
//		t.Update(d)
//	}
//}
