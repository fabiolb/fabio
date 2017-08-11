package circonus

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	cgm "github.com/circonus-labs/circonus-gometrics"
	"github.com/fabiolb/fabio/config"
)

const serviceName = "fabio"

// NewRegistry returns a provider that reports to Circonus.
func NewRegistry(prefix string, circ config.Circonus, interval time.Duration) (*registry, error) {

	if circ.APIKey == "" {
		return nil, errors.New("metrics: Circonus API token key")
	}

	if circ.APIApp == "" {
		circ.APIApp = serviceName
	}

	host, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("metrics: unable to initialize Circonus %s", err)
	}

	cfg := &cgm.Config{}
	cfg.CheckManager.API.TokenKey = circ.APIKey
	cfg.CheckManager.API.TokenApp = circ.APIApp
	cfg.CheckManager.API.URL = circ.APIURL
	cfg.CheckManager.Check.ID = circ.CheckID
	cfg.CheckManager.Broker.ID = circ.BrokerID
	cfg.Interval = fmt.Sprintf("%.0fs", interval.Seconds())
	cfg.CheckManager.Check.InstanceID = host
	cfg.CheckManager.Check.DisplayName = fmt.Sprintf("%s /%s", host, serviceName)
	cfg.CheckManager.Check.SearchTag = fmt.Sprintf("service:%s", serviceName)

	metrics, err := cgm.NewCirconusMetrics(cfg)
	if err != nil {
		return nil, fmt.Errorf("metrics: unable to initialize Circonus %s", err)
	}
	metrics.Start()

	log.Print("[INFO] Sending metrics to Circonus")
	return &registry{metrics, prefix}, nil
}

type registry struct {
	metrics *cgm.CirconusMetrics
	prefix  string
}

// Names is not supported by Circonus.
func (m *registry) Names(string) []string { return nil }

// Unregister is implicitly supported by Circonus,
// stop submitting the metric and it stops being sent to Circonus.
func (m *registry) Unregister(string, string) {}

// UnregisterAll is implicitly supported by Circonus,
// stop submitting metrics and they will no longer be sent to Circonus.
func (m *registry) UnregisterAll(string) {}

func (m *registry) Gauge(_, name string, n float64) {
	m.metrics.Gauge(m.prefix+"`"+name, n)
}

func (m *registry) Inc(_, name string, n int64) {
	m.metrics.Add(m.prefix+"`"+name, uint64(n))
}

func (m *registry) Time(_, name string, d time.Duration) {
	m.metrics.Timing(m.prefix+"`"+name, float64(d))
}
