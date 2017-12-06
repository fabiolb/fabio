// Package static implements a simple static registry
// backend which uses statically configured routes.
package static

import (
	"io/ioutil"
	"log"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/noroute"
	"github.com/fabiolb/fabio/registry"
)

type be struct {
	cfg *config.Static
}

func NewBackend(cfg *config.Static) (registry.Backend, error) {
	return &be{cfg}, nil
}

func (b *be) Register() error {
	return nil
}

func (b *be) Deregister() error {
	return nil
}

func (b *be) ReadManual() (value string, version uint64, err error) {
	return "", 0, nil
}

func (b *be) WriteManual(value string, version uint64) (ok bool, err error) {
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

// WatchNoRouteHTML implementation that reads the noroute html from a
// noroute.html file if it exists
func (b *be) WatchNoRouteHTML() chan string {
	data, err := ioutil.ReadFile(b.cfg.NoRouteHTMLPath)
	if err != nil {
		log.Printf("[WARN] Could not read NoRouteHTMLPath (%s)", b.cfg.NoRouteHTMLPath)
	}
	noroute.SetHTML(string(data))
	return make(chan string)
}
