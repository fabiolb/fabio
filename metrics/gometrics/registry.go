package gometrics

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"time"

	graphite "github.com/cyberdelia/go-metrics-graphite"
	statsd "github.com/magiconair/go-metrics-statsd"
	gm "github.com/rcrowley/go-metrics"
)

// NewStdoutRegistry returns a go-metrics registry that reports to stdout.
func NewStdoutRegistry(interval time.Duration) (*registry, error) {
	logger := log.New(os.Stderr, "localhost: ", log.Lmicroseconds)
	r := gm.NewRegistry()
	go gm.Log(r, interval, logger)
	return &registry{r}, nil
}

// NewGraphiteRegistry returns a go-metrics registry that reports to a Graphite server.
func NewGraphiteRegistry(prefix, addr string, interval time.Duration) (*registry, error) {
	if addr == "" {
		return nil, errors.New(" graphite addr missing")
	}

	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf(" cannot connect to Graphite: %s", err)
	}

	r := gm.NewRegistry()
	go graphite.Graphite(r, interval, prefix, a)
	return &registry{r}, nil
}

// NewStatsDRegistry returns a go-metrics registry that reports to a StatsD server.
func NewStatsDRegistry(prefix, addr string, interval time.Duration) (*registry, error) {
	if addr == "" {
		return nil, errors.New(" statsd addr missing")
	}

	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf(" cannot connect to StatsD: %s", err)
	}

	r := gm.NewRegistry()
	go statsd.StatsD(r, interval, prefix, a)
	return &registry{r}, nil
}

// registry implements the Registry interface
// using the github.com/rcrowley/go-metrics library.
//
// go-metrics supports the concept of different registries
// which allows to have namespaces for registries. For fabio
// we only need two "default" and "services" so we don't implement
// a generic approach with a map, locks and lookup but just have
// a registry for services and one for the rest.
type registry struct {
	service gm.Registry
}

func (p *registry) reg(group string) gm.Registry {
	switch group {
	case "services":
		return p.service
	default:
		return gm.DefaultRegistry
	}
}

func (p *registry) Names(group string) (names []string) {
	p.reg(group).Each(func(name string, _ interface{}) {
		names = append(names, name)
	})
	sort.Strings(names)
	return names
}

func (p *registry) Unregister(group, name string) {
	p.reg(group).Unregister(name)
}

func (p *registry) UnregisterAll(group string) {
	p.reg(group).UnregisterAll()
}

func (p *registry) Gauge(group, name string, n float64) {
	gm.GetOrRegisterGauge(name, p.reg(group)).Update(int64(n))
}

func (p *registry) Inc(group, name string, n int64) {
	gm.GetOrRegisterCounter(name, p.reg(group)).Inc(n)
}

func (p *registry) Time(group, name string, d time.Duration) {
	gm.GetOrRegisterTimer(name, p.reg(group)).Update(d)
}
