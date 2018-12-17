package graphite

import (
	"errors"
	"fmt"
	"github.com/cyberdelia/go-metrics-graphite"
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/metrics4"
	"github.com/fabiolb/fabio/metrics4/gm"
	rcgm "github.com/rcrowley/go-metrics"
	"net"
)

func NewProvider(cfg config.Graphite) (metrics4.Provider, error) {
	if cfg.Addr == "" {
		return nil, errors.New(" graphite addr missing")
	}

	a, err := net.ResolveTCPAddr("tcp", cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf(" cannot connect to Graphite: %s", err)
	}

	registry := rcgm.NewRegistry()

	go graphite.Graphite(registry, cfg.Interval, metrics4.FabioNamespace, a)

	return gm.NewProvider(registry), nil
}