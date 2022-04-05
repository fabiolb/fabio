package metrics

import (
	gkm "github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
)

type DiscardProvider struct{}

func (dp DiscardProvider) NewCounter(name string, labels ...string) gkm.Counter {
	return discard.NewCounter()
}

func (dp DiscardProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	return discard.NewGauge()
}

func (dp DiscardProvider) NewHistogram(name string, labels ...string) gkm.Histogram {
	return discard.NewHistogram()
}
