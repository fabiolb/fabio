package metrics

import "time"

// NoopRegistry is a stub implementation of the Registry interface.
type NoopRegistry struct{}

func (p NoopRegistry) Names() []string { return nil }

func (p NoopRegistry) Unregister(name string) {}

func (p NoopRegistry) UnregisterAll() {}

func (p NoopRegistry) GetCounter(name string) Counter { return noopCounter }

func (p NoopRegistry) GetTimer(name string) Timer { return noopTimer }

var noopCounter = NoopCounter{}

// NoopCounter is a stub implementation of the Counter interface.
type NoopCounter struct{}

func (c NoopCounter) Inc(n int64) {}

var noopTimer = NoopTimer{}

// NoopTimer is a stub implementation of the Timer interface.
type NoopTimer struct{}

func (t NoopTimer) Update(time.Duration) {}

func (t NoopTimer) UpdateSince(time.Time) {}

func (t NoopTimer) Rate1() float64 { return 0 }

func (t NoopTimer) Percentile(nth float64) float64 { return 0 }
