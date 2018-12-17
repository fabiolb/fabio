package graphite

import (
	"github.com/fabiolb/fabio/config"
	"testing"
	"time"
)

const addr = ":9876"

// It shouldn't panic after creating several metrics with the same name
func TestIdenticalNamesForCounters(t *testing.T) {
	metricName := "metric"
	provider, err := NewProvider(config.Graphite{Interval: 1 * time.Second})

	if err != nil {
		t.Error(err)
	}

	counter := provider.NewCounter(metricName)
	counter.Add(1)
	counter = provider.NewCounter(metricName)
	counter.Add(1)
}
