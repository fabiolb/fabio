package metrics

import (
	"context"
	"fmt"
	"net"
	"time"

	gkm "github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/dogstatsd"
	"github.com/go-kit/log"
)

type DogstatsdProvider struct {
	D *dogstatsd.Dogstatsd
}

func (dp *DogstatsdProvider) NewCounter(name string, labels ...string) gkm.Counter {
	return &dogstatsdCounter{dp.D.NewCounter(name, 1)}
}

func (dp *DogstatsdProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	return &dogstatsdGauge{dp.D.NewGauge(name)}
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

type dogstatsdCounter struct {
	gkm.Counter
}
type dogstatsdGauge struct {
	gkm.Gauge
}
type dogstatsdHistogram struct {
	gkm.Histogram
}

func (dh *dogstatsdHistogram) Observe(value float64) {
	dh.Histogram.Observe(value * 1000.0)
}

func (dh *dogstatsdCounter) With(labelValues ...string) gkm.Counter {
	return dh.Counter.With(correctReservedTagKeys(labelValues)...)
}

func (dh *dogstatsdGauge) With(labelValues ...string) gkm.Gauge {
	return dh.Gauge.With(correctReservedTagKeys(labelValues)...)
}

func (dh *dogstatsdHistogram) With(labelValues ...string) gkm.Histogram {
	return dh.Histogram.With(correctReservedTagKeys(labelValues)...)
}

func correctReservedTagKeys(labelValues []string) []string {
	var rval []string
	for i, v := range labelValues {
		if i%2 == 0 {
			rval = append(rval, correctReservedTagKey(v))
		} else {
			rval = append(rval, v)
		}
	}
	return rval
}

func correctReservedTagKey(label string) string {
	switch label {
	case "host":
		return "fabio-host"
	case "device":
		return "fabio-device"
	case "source":
		return "fabio-source"
	case "service":
		return "fabio-service"
	case "env":
		return "fabio-env"
	case "version":
		return "fabio-version"
	default:
		return label
	}
}
