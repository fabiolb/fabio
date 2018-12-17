package circonus

import (
	"errors"
	"fmt"
	cgm "github.com/circonus-labs/circonus-gometrics"
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/metrics4"
	"github.com/fabiolb/fabio/metrics4/gm"
	rcgm "github.com/rcrowley/go-metrics"
	"log"
	"os"
	"sync"
	"time"
)

var (
	circonus *cgmRegistry
	once     sync.Once
)

func NewProvider(cfg config.Circonus, interval time.Duration) (metrics4.Provider, error) {
	r, err := circonusRegistry("", cfg, interval)
	if err != nil {
		return nil, err
	}
	return gm.NewProvider(r), nil
}

func circonusRegistry(prefix string, circonusCfg config.Circonus, interval time.Duration) (rcgm.Registry, error) {
	var initError error

	once.Do(func() {
		if circonusCfg.APIKey == "" {
			initError = errors.New("metrics: Circonus API token key")
			return
		}

		if circonusCfg.APIApp == "" {
			circonusCfg.APIApp = metrics4.FabioNamespace
		}

		host, err := os.Hostname()
		if err != nil {
			initError = fmt.Errorf("metrics: unable to initialize Circonus %s", err)
			return
		}

		cfg := &cgm.Config{}

		cfg.CheckManager.API.TokenKey = circonusCfg.APIKey
		cfg.CheckManager.API.TokenApp = circonusCfg.APIApp
		cfg.CheckManager.API.URL = circonusCfg.APIURL
		cfg.CheckManager.Check.ID = circonusCfg.CheckID
		cfg.CheckManager.Broker.ID = circonusCfg.BrokerID
		cfg.Interval = fmt.Sprintf("%.0fs", interval.Seconds())
		cfg.CheckManager.Check.InstanceID = host
		cfg.CheckManager.Check.DisplayName = fmt.Sprintf("%s /%s", host, metrics4.FabioNamespace)
		cfg.CheckManager.Check.SearchTag = fmt.Sprintf("service:%s", metrics4.FabioNamespace)

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

func (m *cgmRegistry) Names() []string { return nil }

func (m *cgmRegistry) Get(string) interface{} { return nil }

func (m *cgmRegistry) Register(string, interface{}) error { return nil }

func (m *cgmRegistry) RunHealthchecks() {}

func (m *cgmRegistry) GetOrRegister(string, interface{}) interface{} { return nil }

func (m *cgmRegistry) Each(func(string, interface{})) {}

func (m *cgmRegistry) Unregister(name string) {}

func (m *cgmRegistry) UnregisterAll() {}

func (m *cgmRegistry) GetCounter(name string) rcgm.Counter {
	metricName := fmt.Sprintf("%s`%s", m.prefix, name)
	return &cgmCounter{m.metrics, metricName}
}

func (m *cgmRegistry) GetTimer(name string) rcgm.Timer {
	metricName := fmt.Sprintf("%s`%s", m.prefix, name)
	return &cgmTimer{m.metrics, metricName}
}

type cgmCounter struct {
	metrics *cgm.CirconusMetrics
	name    string
}

func (c *cgmCounter) Inc(n int64) {
	c.metrics.IncrementByValue(c.name, uint64(n))
}

func (c *cgmCounter) Dec(n int64) {}

func (c *cgmCounter) Clear() {}

func (c *cgmCounter) Count() int64 {
	return 0
}

func (c *cgmCounter) Snapshot() rcgm.Counter {
	return c
}

type cgmTimer struct {
	metrics *cgm.CirconusMetrics
	name    string
}

func (t *cgmTimer) Percentile(nth float64) float64 { return 0 }

func (t *cgmTimer) Rate1() float64 { return 0 }

func (t *cgmTimer) Update(d time.Duration) {
	t.metrics.Timing(t.name, float64(d))
}

func (t *cgmTimer) UpdateSince(start time.Time) {}

func (t *cgmTimer) Count() int64 {
	return 0
}

func (t *cgmTimer) Min() int64 {
	return 0
}

func (t *cgmTimer) Max() int64 {
	return 0
}

func (t *cgmTimer) Mean() float64 {
	return 0
}

func (t *cgmTimer) Percentiles([]float64) []float64 {
	return nil
}

func (t *cgmTimer) Rate5() float64 {
	return 0
}

func (t *cgmTimer) Rate15() float64 {
	return 0
}

func (t *cgmTimer) Snapshot() rcgm.Timer {
	return t
}

func (t *cgmTimer) RateMean() float64 {
	return 0
}

func (t *cgmTimer) StdDev() float64 {
	return 0
}

func (t *cgmTimer) Sum() int64 {
	return 0
}

func (t *cgmTimer) Time(func()) {}

func (t *cgmTimer) Variance() float64 {
	return 0
}
