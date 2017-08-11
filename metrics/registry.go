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
	Names(group string) []string

	// Unregister removes the registered metric and stops
	// reporting it to an external backend.
	Unregister(group, name string)

	// UnregisterAll removes all registered metrics and stops
	// reporting  them to an external backend.
	UnregisterAll(group string)

	// Gauge increments or decrements a gauge metric with the given name.
	Gauge(group, name string, n float64)

	// Inc increments a counter metric with the given name.
	Inc(group, name string, n int64)

	// Time updates a timer metric with the given name.
	Time(group, name string, d time.Duration)
}

const (
	// when adding more groups make sure to update the
	// gometrics registry
	DefaultGroup = "default"
	ServiceGroup = "services"
)

func GaugeDefault(name string, n float64)      { M.Gauge(DefaultGroup, name, n) }
func IncDefault(name string, n int64)          { M.Inc(DefaultGroup, name, n) }
func TimeDefault(name string, d time.Duration) { M.Time(DefaultGroup, name, d) }

func GaugeService(name string, n float64)      { M.Gauge(ServiceGroup, name, n) }
func IncService(name string, n int64)          { M.Inc(ServiceGroup, name, n) }
func TimeService(name string, d time.Duration) { M.Time(ServiceGroup, name, d) }
