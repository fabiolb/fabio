package metrics

import (
	"errors"
	"fmt"
	"time"

	alstatsd "github.com/alexcesaro/statsd"
)

func newRawStatsDRegistry(prefix, addr string, interval time.Duration) (Registry, error) {
	if addr == "" {
		return nil, errors.New(" statsd addr missing")
	}

	c, err := alstatsd.New(alstatsd.Address(addr), alstatsd.FlushPeriod(interval))
	if err != nil {
		return nil, fmt.Errorf(" cannot init statsd client: %s", err)
	}

	return &rawStatsDRegistry{c}, nil
}

type rawStatsDRegistry struct {
	c *alstatsd.Client
}

func (r *rawStatsDRegistry) Names() []string        { return nil }
func (r *rawStatsDRegistry) Unregister(name string) {}
func (r *rawStatsDRegistry) UnregisterAll()         {}

func (r *rawStatsDRegistry) GetCounter(name string) Counter {
	return &rawStatsDCounter{r.c, name}
}

func (r *rawStatsDRegistry) GetTimer(name string) Timer {
	return &rawStatsDTimer{r.c, name}
}

type rawStatsDCounter struct {
	c    *alstatsd.Client
	name string
}

func (c *rawStatsDCounter) Inc(n int64) {
	c.c.Increment(c.name)
}

type rawStatsDTimer struct {
	c    *alstatsd.Client
	name string
}

func (t *rawStatsDTimer) Update(d time.Duration) {
	t.c.Timing(t.name, int(d/time.Millisecond))
}

func (t *rawStatsDTimer) UpdateSince(start time.Time) {
	t.Update(time.Now().Sub(start))
}

func (t *rawStatsDTimer) Rate1() float64                 { return 0 }
func (t *rawStatsDTimer) Percentile(nth float64) float64 { return 0 }
