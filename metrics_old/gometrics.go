package metrics

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"time"

	graphite "github.com/cyberdelia/go-metrics-graphite"
	statsd "github.com/pubnub/go-metrics-statsd"
	gm "github.com/rcrowley/go-metrics"
)

// gmStdoutRegistry returns a go-metrics registry that reports to stdout.
func gmStdoutRegistry(interval time.Duration) (Registry, error) {
	logger := log.New(os.Stderr, "localhost: ", log.Lmicroseconds)
	r := gm.NewRegistry()
	go gm.Log(r, interval, logger)
	return &gmRegistry{r}, nil
}

// gmGraphiteRegistry returns a go-metrics registry that reports to a Graphite server.
func gmGraphiteRegistry(prefix, addr string, interval time.Duration) (Registry, error) {
	if addr == "" {
		return nil, errors.New(" graphite addr missing")
	}

	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf(" cannot connect to Graphite: %s", err)
	}

	r := gm.NewRegistry()
	go graphite.Graphite(r, interval, prefix, a)
	return &gmRegistry{r}, nil
}

// gmStatsDRegistry returns a go-metrics registry that reports to a StatsD server.
func gmStatsDRegistry(prefix, addr string, interval time.Duration) (Registry, error) {
	if addr == "" {
		return nil, errors.New(" statsd addr missing")
	}

	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf(" cannot connect to StatsD: %s", err)
	}

	r := gm.NewRegistry()
	go statsd.StatsD(r, interval, prefix, a)
	return &gmRegistry{r}, nil
}

// gmRegistry implements the Registry interface
// using the github.com/rcrowley/go-metrics library.
type gmRegistry struct {
	r gm.Registry
}

func (p *gmRegistry) Names() (names []string) {
	p.r.Each(func(name string, _ interface{}) {
		names = append(names, name)
	})
	sort.Strings(names)
	return names
}

func (p *gmRegistry) Unregister(name string) {
	p.r.Unregister(name)
}

func (p *gmRegistry) UnregisterAll() {
	p.r.UnregisterAll()
}

func (p *gmRegistry) GetCounter(name string) Counter {
	return gm.GetOrRegisterCounter(name, p.r)
}

func (p *gmRegistry) GetTimer(name string) Timer {
	return gm.GetOrRegisterTimer(name, p.r)
}
