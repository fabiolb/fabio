package metrics4

import (
	"time"

	"github.com/go-kit/kit/metrics"
)

const FabioNamespace = "fabio"

type Counter metrics.Counter

type Gauge metrics.Gauge

type Histogram metrics.Histogram

type TimerStruct struct {
	histogram Histogram
	start     time.Time
}

func NewTimerStruct(h Histogram, start time.Time) Timer {
	return &TimerStruct{
		h,
		start,
	}
}

type Timer interface {
	Start()
	Stop()
	Observe(duration time.Duration)
	With(labelValues ... string) Timer
}

func (t *TimerStruct) Stop() {
	t.histogram.Observe(float64(time.Since(t.start).Nanoseconds()) / float64(time.Millisecond))
}

func (t *TimerStruct) Start() {
	t.start = time.Now()
}

func (t *TimerStruct) Observe(duration time.Duration) {
	t.histogram.Observe(float64(duration.Nanoseconds()) / float64(time.Millisecond))
}

func (t *TimerStruct) With(labelValues ... string) Timer {
	return &TimerStruct{
		t.histogram.With(labelValues...),
		t.start,
	}
}

// Provider is an abstraction of a metrics backend.
type Provider interface {
	// NewCounter creates a new counter object.
	NewCounter(name string, labels... string) Counter

	// NewGauge creates a new gauge object.
	NewGauge(name string, labels... string) Gauge

	// NewTimer creates a new timer object.
	NewTimer(name string, labels... string) Timer

	// NewHistogram creates a new histogram object.
	NewHistogram(name string, labels... string) Histogram

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
func (mp *MultiProvider) NewCounter(name string, labels... string) Counter {
	var c []Counter
	for _, p := range mp.p {
		c = append(c, p.NewCounter(name, labels...))
	}
	return &MultiCounter{c}
}

// NewGauge creates a MultiGauge with gauge objects for all registered
// providers.
func (mp *MultiProvider) NewGauge(name string, labels... string) Gauge {
	var g []Gauge
	for _, p := range mp.p {
		g = append(g, p.NewGauge(name, labels...))
	}
	return &MultiGauge{g}
}

// NewTimer creates a MultiTimer with timer objects for all registered
// providers.
func (mp *MultiProvider) NewTimer(name string, labels... string) Timer {
	var t []Timer
	for _, p := range mp.p {
		t = append(t, p.NewTimer(name, labels...))
	}
	return &MultiTimer{t}
}

// NewHistogram creates a MultiHistogram with histogram objects for all registered
// providers.
func (mp *MultiProvider) NewHistogram(name string, labels... string) Histogram {
	var h []Histogram
	for _, p := range mp.p {
		h = append(h, p.NewHistogram(name, labels...))
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
	labeledHistograms := make([]Histogram, len(mh.histograms))
	for i, h := range mh.histograms {
		labeledHistograms[i] = h.With(labelValues...)
	}
	return &MultiHistogram{labeledHistograms}
}

// MultiTimer wraps zero or more timers.
type MultiTimer struct {
	timers []Timer
}

func (mt *MultiTimer) Observe(duration time.Duration) {
	for _, t := range mt.timers {
		t.Observe(duration)
	}
}

func (mt *MultiTimer) Start() {
	for _, t := range mt.timers {
		t.Start()
	}
}

func (mt *MultiTimer) Stop() {
	for _, t := range mt.timers {
		t.Stop()
	}
}

func (mt *MultiTimer) With(labelValues ... string) Timer {
	labeledTimers := make([]Timer, len(mt.timers))
	for i, t := range mt.timers {
		labeledTimers[i] = t.With(labelValues...)
	}
	return &MultiTimer{labeledTimers}
}
