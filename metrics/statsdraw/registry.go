package statsdraw

import (
	"fmt"
	"time"

	alstatsd "github.com/alexcesaro/statsd"
)

func NewRegistry(prefix, addr string, interval time.Duration) (*registry, error) {
	opts := []alstatsd.Option{
		alstatsd.Address(addr),
		alstatsd.FlushPeriod(interval),
	}
	if prefix != "" {
		opts = append(opts, alstatsd.Prefix(prefix))
	}
	c, err := alstatsd.New(opts...)
	if err != nil {
		return nil, fmt.Errorf(" cannot init statsd client: %s", err)
	}
	return &registry{c}, nil
}

type registry struct {
	c *alstatsd.Client
}

func (r *registry) Names(string) []string     { return nil }
func (r *registry) Unregister(string, string) {}
func (r *registry) UnregisterAll(string)      {}

func (r *registry) Gauge(_, name string, n float64) {
	r.c.Count(name, n)
}

func (r *registry) Inc(_, name string, n int64) {
	r.c.Count(name, n)
}

func (r *registry) Time(_, name string, d time.Duration) {
	r.c.Timing(name, int(d/time.Millisecond))
}
