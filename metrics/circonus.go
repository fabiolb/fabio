package metrics

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	cgm "github.com/circonus-labs/circonus-gometrics"
)

type cgmRegistry struct {
	metrics *cgm.CirconusMetrics
	prefix  string
}

type cgmTimer struct {
	metrics *cgm.CirconusMetrics
	name    string
}

var (
	circonus *cgmRegistry
	once     sync.Once
)

const serviceName = "fabio"

// circonusBackend returns a provider that reports to Circonus.
func circonusBackend(prefix string,
	circKey string,
	circApp string,
	circURL string,
	circBrokerID string,
	circCheckID string,
	interval time.Duration) (Registry, error) {

	var initError error

	once.Do(func() {

		if circKey == "" {
			initError = errors.New("metrics: Circonus API token key")
			return
		}

		if circApp == "" {
			circApp = serviceName
		}

		host, err := os.Hostname()
		if err != nil {
			initError = fmt.Errorf("metrics: unable to initialize Circonus %s", err)
			return
		}

		cfg := &cgm.Config{}

		cfg.CheckManager.API.TokenKey = circKey
		cfg.CheckManager.API.TokenApp = circApp
		cfg.CheckManager.API.URL = circURL
		cfg.CheckManager.Check.ID = circCheckID
		cfg.CheckManager.Broker.ID = circBrokerID
		cfg.Interval = fmt.Sprintf("%.0fs", interval.Seconds())
		cfg.CheckManager.Check.InstanceID = host
		cfg.CheckManager.Check.DisplayName = fmt.Sprintf("%s /%s", host, serviceName)
		cfg.CheckManager.Check.SearchTag = fmt.Sprintf("service:%s", serviceName)

		metrics, err := cgm.NewCirconusMetrics(cfg)
		if err != nil {
			initError = fmt.Errorf("metrics: unable to initialize Circonus %s", err)
		} else {
			circonus = &cgmRegistry{metrics, prefix}
		}

		metrics.Start()

		log.Print("[INFO] Sending metrics to Circonus")

	})

	return circonus, initError
}

// Names returns the list of registered metrics acquired
// through the GetXXX() functions. It should return them
// sorted in alphabetical order.
// Unsupported by Circonus.
func (m *cgmRegistry) Names() []string { return nil }

// Unregister removes the registered metric and stops
// reporting it to an external backend.
// Implicitly supported by Circonus, stop submitting
// the metric and it stops being sent to Circonus.
func (m *cgmRegistry) Unregister(name string) {}

// UnregisterAll removes all registered metrics and stops
// reporting  them to an external backend.
// Implicitly supported by Circonus, stop submitting
// metrics and they will no longer be sent to Circonus.
func (m *cgmRegistry) UnregisterAll() {}

// GetTimer returns a timer metric for the given name.
// If the metric does not exist yet it should be created
// otherwise the existing metric should be returned.
func (m *cgmRegistry) GetTimer(name string) Timer {
	metricName := fmt.Sprintf("%s`%s", m.prefix, name)
	return &cgmTimer{m.metrics, metricName}
}

// Percentile returns the nth percentile of the duration.
// Circonus does not support in-memory derivatives.
func (t cgmTimer) Percentile(nth float64) float64 {
	return 0
}

// Rate1 returns the 1min rate.
// Circonus does not support in-memory derivatives.
func (t cgmTimer) Rate1() float64 {
	return 0
}

// UpdateSince counts an event and records the duration
// as the delta between 'start' and when the function is called.
// Circonus - add another delta to the named histogram.
// Histograms are automatically created when a new metric
// name is encountered.
func (t cgmTimer) UpdateSince(start time.Time) {
	t.metrics.Timing(t.name, float64(time.Since(start)))
}
