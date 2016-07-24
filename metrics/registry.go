package metrics

import "time"

// Registry defines an interface for metrics values which
// can be implemented by different metrics libraries.
// An implementation must be safe to be used by multiple
// go routines.
type Registry interface {
	// Names returns the list of registered metrics acquired
	// through the GetXXX() functions. It should return them
	// sorted in alphabetical order.
	Names() []string

	// Unregister removes the registered metric and stops
	// reporting it to an external backend.
	Unregister(name string)

	// UnregisterAll removes all registered metrics and stops
	// reporting  them to an external backend.
	UnregisterAll()

	// GetCounter returns a counter metric for the given name.
	// If the metric does not exist yet it should be created
	// otherwise the existing metric should be returned.
	GetCounter(name string) Counter

	// GetTimer returns a timer metric for the given name.
	// If the metric does not exist yet it should be created
	// otherwise the existing metric should be returned.
	GetTimer(name string) Timer
}

// Counter defines a metric for counting events.
type Counter interface {
	// Inc increases the counter value by 'n'.
	Inc(n int64)
}

// Timer defines a metric for counting and timing durations for events.
type Timer interface {
	// Percentile returns the nth percentile of the duration.
	Percentile(nth float64) float64

	// Rate1 returns the 1min rate.
	Rate1() float64

	// Update counts an event and records the duration.
	Update(time.Duration)

	// UpdateSince counts an event and records the duration
	// as the delta between 'start' and the function is called.
	UpdateSince(start time.Time)
}
