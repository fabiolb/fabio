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

var (
	circonus *cgmRegistry
	once     sync.Once
)

const serviceName = "fabio"

// circonusRegistry returns a provider that reports to Circonus.
func circonusRegistry(prefix string,
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
			return
		}

		circonus = &cgmRegistry{metrics, prefix}

		metrics.Start()

		log.Print("[INFO] Sending metrics to Circonus")
	})

	return circonus, initError
}

type cgmRegistry struct {
	metrics *cgm.CirconusMetrics
	prefix  string
}

// Names is not supported by Circonus.
func (m *cgmRegistry) Names() []string { return nil }

// Unregister is implicitly supported by Circonus,
// stop submitting the metric and it stops being sent to Circonus.
func (m *cgmRegistry) Unregister(name string) {}

// UnregisterAll is implicitly supported by Circonus,
// stop submitting metrics and they will no longer be sent to Circonus.
func (m *cgmRegistry) UnregisterAll() {}

// GetCounter returns a counter for the given metric name.
func (m *cgmRegistry) GetCounter(name string) Counter {
	metricName := fmt.Sprintf("%s`%s", m.prefix, name)
	return &cgmCounter{m.metrics, metricName}
}

// GetTimer returns a timer for the given metric name.
func (m *cgmRegistry) GetTimer(name string) Timer {
	metricName := fmt.Sprintf("%s`%s", m.prefix, name)
	return &cgmTimer{m.metrics, metricName}
}

type cgmCounter struct {
	metrics *cgm.CirconusMetrics
	name    string
}

// Inc increases the counter by n.
func (c *cgmCounter) Inc(n int64) {
	c.metrics.IncrementByValue(c.name, uint64(n))
}

type cgmTimer struct {
	metrics *cgm.CirconusMetrics
	name    string
}

// Percentile is not supported by Circonus.
func (t *cgmTimer) Percentile(nth float64) float64 { return 0 }

// Rate1 is not supported by Circonus.
func (t *cgmTimer) Rate1() float64 { return 0 }

// UpdateSince adds delta between start and current time as
// a sample to a histogram. The histogram is created if it
// does not already exist.
func (t *cgmTimer) UpdateSince(start time.Time) {
	t.metrics.Timing(t.name, float64(time.Since(start)))
}
