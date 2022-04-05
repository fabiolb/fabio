package metrics

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-kit/kit/log"
	gkm "github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/dogstatsd"
)

type DogstatsdProvider struct {
	D *dogstatsd.Dogstatsd
}

func (dp *DogstatsdProvider) NewCounter(name string, labels ...string) gkm.Counter {
	return dp.D.NewCounter(name, 1)
}

func (dp *DogstatsdProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	return dp.D.NewGauge(name)
}

func (dp *DogstatsdProvider) NewHistogram(name string, labels ...string) gkm.Histogram {
	return &dogstatsdHistogram{dp.D.NewHistogram(name, 1)}
}

func NewDogstatsdProvider(prefix, addr string, interval time.Duration) (*DogstatsdProvider, error) {
	d := &DogstatsdProvider{
		D: dogstatsd.New(prefix, log.NewNopLogger()),
	}
	_, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("error resolving dogstatsd address %s: %w", addr, err)
	}
	t := time.NewTicker(interval)
	go func() {
		d.D.SendLoop(context.Background(), t.C, "udp", addr)
	}()
	return d, nil
}

type dogstatsdHistogram struct {
	gkm.Histogram
}

func (dh *dogstatsdHistogram) Observe(value float64) {
	dh.Histogram.Observe(value * 1000.0)
}
