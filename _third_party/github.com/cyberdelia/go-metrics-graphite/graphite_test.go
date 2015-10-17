package graphite

import (
	"net"
	"time"

	"github.com/eBay/fabio/_third_party/github.com/rcrowley/go-metrics"
)

func ExampleGraphite() {
	addr, _ := net.ResolveTCPAddr("net", ":2003")
	go Graphite(metrics.DefaultRegistry, 1*time.Second, "some.prefix", addr)
}

func ExampleGraphiteWithConfig() {
	addr, _ := net.ResolveTCPAddr("net", ":2003")
	go GraphiteWithConfig(GraphiteConfig{
		Addr:          addr,
		Registry:      metrics.DefaultRegistry,
		FlushInterval: 1 * time.Second,
		DurationUnit:  time.Millisecond,
		Percentiles:   []float64{0.5, 0.75, 0.99, 0.999},
	})
}
