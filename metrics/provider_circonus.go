package metrics

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	cgm "github.com/circonus-labs/circonus-gometrics/v3"
	"github.com/fabiolb/fabio/config"
	gkm "github.com/go-kit/kit/metrics"
)

var (
	circonus *CirconusProvider
	circOnce sync.Once
)

const serviceName = "fabio"

func NewCirconusProvider(prefix string, circ config.Circonus, interval time.Duration) (*CirconusProvider, error) {
	var initError error

	circOnce.Do(func() {
		if circ.APIKey == "" && circ.SubmissionURL == "" {
			initError = errors.New("metrics: Circonus API token key or SubmissionURL")
			return
		}

		if circ.APIApp == "" {
			circ.APIApp = serviceName
		}

		host, err := os.Hostname()
		if err != nil {
			initError = fmt.Errorf("metrics: unable to initialize Circonus %s", err)
			return
		}

		cfg := &cgm.Config{}

		cfg.CheckManager.Check.SubmissionURL = circ.SubmissionURL
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
			initError = fmt.Errorf("metrics: unable to initialize Circonus %s", err)
			return
		}

		circonus = &CirconusProvider{metrics, prefix}

		metrics.Start()

		log.Print("[INFO] Sending metrics to Circonus")
	})

	return circonus, initError
}

type CirconusProvider struct {
	metrics *cgm.CirconusMetrics
	prefix  string
}

func (cp *CirconusProvider) metricName(name string) string {
	return fmt.Sprintf("%s`%s", cp.prefix, name)
}

func (cp *CirconusProvider) NewCounter(name string, labels ...string) gkm.Counter {
	return &cgmCounter{
		p:           cp,
		name:        cp.metricName(name),
		routeMetric: isRouteMetric(name),
	}
}

func (cp *CirconusProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	return &cgmGauge{
		p:           cp,
		name:        cp.metricName(name),
		routeMetric: isRouteMetric(name),
	}
}

func (cp *CirconusProvider) NewHistogram(name string, labels ...string) gkm.Histogram {
	return &cgmTimer{
		p:           cp,
		name:        cp.metricName(name),
		routeMetric: isRouteMetric(name),
	}
}

type cgmCounter struct {
	p           *CirconusProvider
	name        string
	routeMetric bool
}

func (c *cgmCounter) With(labelValues ...string) gkm.Counter {
	var name string
	switch c.routeMetric {
	case true:
		var err error
		name, err = TargetNameWith(c.name, labelValues)
		if err != nil {
			panic(err)
		}
		name = c.p.metricName(name)
	case false:
		name = Flatten(c.name, labelValues, DotSeparator)
	}
	return &cgmCounter{
		p:           c.p,
		name:        name,
		routeMetric: c.routeMetric,
	}
}

func (c *cgmCounter) Add(delta float64) {
	c.p.metrics.IncrementByValue(c.name, uint64(delta))
}

type cgmGauge struct {
	p           *CirconusProvider
	name        string
	routeMetric bool
}

func (g *cgmGauge) With(labelValues ...string) gkm.Gauge {
	var name string
	switch g.routeMetric {
	case true:
		var err error
		name, err = TargetNameWith(g.name, labelValues)
		if err != nil {
			panic(err)
		}
		name = g.p.metricName(name)
	case false:
		name = Flatten(g.name, labelValues, DotSeparator)
	}
	return &cgmGauge{
		p:           g.p,
		name:        name,
		routeMetric: g.routeMetric,
	}
}

func (g *cgmGauge) Set(value float64) {
	g.p.metrics.Gauge(g.name, value)
}

func (g *cgmGauge) Add(delta float64) {
	g.p.metrics.AddGauge(g.name, delta)
}

type cgmTimer struct {
	p           *CirconusProvider
	name        string
	routeMetric bool
}

func (t *cgmTimer) With(labelValues ...string) gkm.Histogram {
	var name string
	switch t.routeMetric {
	case true:
		var err error
		name, err = TargetNameWith(t.name, labelValues)
		if err != nil {
			panic(err)
		}
		name = t.p.metricName(name)
	case false:
		name = Flatten(t.name, labelValues, DotSeparator)
	}
	return &cgmTimer{
		p:           t.p,
		name:        name,
		routeMetric: t.routeMetric,
	}
}

func (t *cgmTimer) Observe(value float64) {
	t.p.metrics.Timing(t.name, value*float64(time.Second))
}
