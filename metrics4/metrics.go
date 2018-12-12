package metrics4

import (
	"github.com/go-kit/kit/metrics"
	"time"
)

const FabioNamespace = "fabio"

type Counter metrics.Counter

type Gauge metrics.Gauge

type Histogram metrics.Histogram

// TODO(max): Refactor Timer thingies
type Timer struct {
	histograms []Histogram
	start      time.Time
	unit       time.Duration
}

type ITimer interface {
	Unit(time.Duration)
	Reset()
	Stop()
	Duration(float64)
	With(labelValues... string) ITimer
}

func (t *Timer) Unit(u time.Duration) {
	t.unit = u
}

func (t *Timer) Stop() {
	duration := float64(time.Since(t.start).Nanoseconds()) / float64(t.unit)
	for _, h := range t.histograms {
		h.Observe(duration)
	}
}

func (t *Timer) Reset() {
	t.start = time.Now()
}

func (t *Timer) Duration(duration float64) {
	duration = duration / float64(t.unit)
	for _, h := range t.histograms {
		h.Observe(duration)
	}
}

func (t *Timer) With(labelValues... string) ITimer {
	for _, h := range t.histograms {
		h.With(labelValues...)
	}
	return t
}

// Provider is an abstraction of a metrics backend.
type Provider interface {
	// NewCounter creates a new counter object.
	NewCounter(name string) Counter

	// NewGauge creates a new gauge object.
	NewGauge(name string) Gauge

	// NewTimer creates a new timer object.
	NewTimer(name string) ITimer

	// NewHistogram creates a new histogram object.
	NewHistogram(name string) Histogram

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
func (mp *MultiProvider) NewTimer(name string) ITimer {
	var h []Histogram

	for _, p := range mp.p {
		h = append(h, p.NewHistogram(name))
	}

	return &Timer{
		histograms: h,
		start:      time.Now(),
		unit:       time.Millisecond,
	}
}

// NewHistogram creates a MultiTimer with timer objects for all registered
// providers.
func (mp *MultiProvider) NewHistogram(name string) Histogram {
	var h []Histogram
	for _, p := range mp.p {
		h = append(h, p.NewHistogram(name))
	}
	return &MultiHistogram{h}
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

func (mg *MultiGauge) With(labelValues ... string) metrics.Gauge {
	for _, g := range mg.gauges {
		g.With(labelValues...)
	}
	return mg
}

// MultiGauge wraps zero or more gauges.
type MultiHistogram struct {
	histograms []Histogram
}

func (mh *MultiHistogram) Observe(delta float64) {
	for _, h := range mh.histograms {
		h.Observe(delta)
	}
}

func (mh *MultiHistogram) With(labelValues ... string) metrics.Histogram {
	for _, h := range mh.histograms {
		h.With(labelValues...)
	}
	return mh
}

