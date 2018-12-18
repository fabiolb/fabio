package circonus

import (
	"errors"
	"fmt"
	cgm "github.com/circonus-labs/circonus-gometrics"
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/metrics4"
	"github.com/fabiolb/fabio/metrics4/untagged"
	"log"
	"os"
	"sync"
	"time"
)

var (
	metrics *cgm.CirconusMetrics
	once    sync.Once
)

type Provider struct {
	c *cgm.CirconusMetrics
	prefix string
}

func NewProvider(circonusCfg config.Circonus, interval time.Duration, prefix string) (metrics4.Provider, error) {
	var initError error
	var metrics *cgm.CirconusMetrics

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

		metrics.Start()

		log.Print("[INFO] Sending metrics to Circonus")
	})

	return &Provider{metrics, prefix}, initError
}

func (p *Provider) NewCounter(name string, labelsNames ... string) metrics4.Counter {
	name = getPrefixName(p.prefix, name)
	if len(labelsNames) == 0 {
		return &Counter{p.c, name}
	}
	return untagged.NewCounter(p, name, labelsNames)
}

func (p *Provider) NewGauge(name string, labelsNames ... string) metrics4.Gauge {
	name = getPrefixName(p.prefix, name)
	if len(labelsNames) == 0 {
		return &Gauge{p.c, name}
	}
	return untagged.NewGauge(p, name, labelsNames)
}

func (p *Provider) NewTimer(name string, labelsNames ... string) metrics4.Timer {
	name = getPrefixName(p.prefix, name)
	if len(labelsNames) == 0 {
		return &Timer{p.c.NewHistogram(name)}
	}
	return untagged.NewTimer(p, name, labelsNames)
}

func (p *Provider) Close() error {
	return nil
}

func getPrefixName(prefix string, name string) string {
	if len(prefix) == 0 {
		return name
	}
	return fmt.Sprintf("%s`%s", prefix, name)
}

type Counter struct {
	metrics *cgm.CirconusMetrics
	name    string
}

func (c *Counter) Add(value float64) {
	c.metrics.Add(c.name, uint64(value))
}

func (c *Counter) With(labels ... string) metrics4.Counter {
	return c
}

type Gauge struct {
	metrics *cgm.CirconusMetrics
	name    string
}

func (g *Gauge) Add(value float64) {
	g.metrics.Add(g.name, uint64(value))
}

func (g *Gauge) Set(value float64) {
	g.metrics.Set(g.name, uint64(value))
}

func (g *Gauge) With(labels ... string) metrics4.Gauge {
	return g
}

type Timer struct {
	h *cgm.Histogram
}

func (t *Timer) Observe(duration float64) {
	t.h.RecordValue(duration)
}

func (g *Timer) With(labels ... string) metrics4.Timer {
	return g
}
