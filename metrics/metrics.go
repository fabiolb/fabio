package metrics

import (
	"fmt"
	"github.com/fabiolb/fabio/config"
	gkm "github.com/go-kit/kit/metrics"
	"log"
	"strings"
)

// Provider is an abstraction of a metrics backend.
type Provider interface {
	// NewCounter creates a new counter object.
	NewCounter(name string, labels ...string) gkm.Counter

	// NewGauge creates a new gauge object.
	NewGauge(name string, labels ...string) gkm.Gauge

	// NewHistogram creates a new histogram object
	NewHistogram(name string, labels ...string) gkm.Histogram
}

func Initialize(cfg *config.Metrics) (Provider, error) {
	var p []Provider
	var prefix string
	var err error
	if prefix, err = parsePrefix(cfg.Prefix); err != nil {
		return nil, fmt.Errorf("metrics: invalid Prefix template: %w", err)
	}
	for _, x := range strings.Split(cfg.Target, ",") {
		x = strings.TrimSpace(x)
		switch x {
		case "flat","stdout":
			p = append(p, &flatProvider{prefix})
		case "label":
			p = append(p, &labelProvider{prefix})
		case "statsd_raw":
			pp, err := NewStatsdProvider(prefix, cfg.StatsDAddr, cfg.Interval)
			if err != nil {
				return nil, err
			}
			p = append(p, pp)
		case "statsd":
			return nil, fmt.Errorf("statsd support has been removed in favor of statsd_raw")
		case "prometheus":
			p = append(p, NewPromProvider(prefix, cfg.Prometheus.Subsystem, cfg.Prometheus.Buckets))
		case "circonus":
			pp, err := NewCirconusProvider(prefix, cfg.Circonus, cfg.Interval)
			if err != nil {
				return nil, err
			}
			p = append(p, pp)
		case "graphite":
			pp, err := NewGraphiteProvider(prefix, cfg.GraphiteAddr, 50, cfg.Interval)
			if err != nil {
				return nil, err
			}
			p = append(p, pp)
		default:
			return nil, fmt.Errorf("invalid metrics backend %s", x)
		}
		log.Printf("[INFO] Registering metrics provider %q", x)

		if len(p) == 0 {
			log.Printf("[INFO] Metrics disabled")
		}
	}
	if len(p) == 0 {
		return &DiscardProvider{}, nil
	}
	return NewMultiProvider(p), nil
}
