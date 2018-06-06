// Package static implements a simple static registry
// backend which uses statically configured routes.
package static

import (
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/registry"
)

type be struct {
	cfg *config.Static
}

func NewBackend(cfg *config.Static) (registry.Backend, error) {
	return &be{cfg}, nil
}

func (b *be) Register(services []string) error {
	return nil
}

func (b *be) Deregister(serviceName string) error {
	return nil
}

func (b *be) DeregisterAll() error {
	return nil
}

func (b *be) ManualPaths() ([]string, error) {
	return nil, nil
}

func (b *be) ReadManual(string) (value string, version uint64, err error) {
	return "", 0, nil
}

func (b *be) WriteManual(path string, value string, version uint64) (ok bool, err error) {
	return false, nil
}

func (b *be) WatchServices() chan string {
	ch := make(chan string, 1)
	ch <- b.cfg.Routes
	return ch
}

func (b *be) WatchManual() chan string {
	return make(chan string)
}

func (b *be) WatchNoRouteHTML() chan string {
	ch := make(chan string, 1)
	ch <- b.cfg.NoRouteHTML
	return ch
}
