// Package static implements a simple static registry
// backend which uses statically configured routes.
package static

import (
	"io/ioutil"
	"log"

	"github.com/fabiolb/fabio/registry"
	"github.com/fabiolb/fabio/route"
)

type be struct{}

var staticRoutes string

func NewBackend(routes string) (registry.Backend, error) {
	staticRoutes = routes
	return &be{}, nil
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
	ch <- staticRoutes
	return ch
}

func (b *be) WatchManual() chan string {
	return make(chan string)
}

// WatchNoRouteHTML implementation that reads the noroute html from a
// noroute.html file if it exists
func (b *be) WatchNoRouteHTML() chan string {
	data, err := ioutil.ReadFile("noroute.html")
	if err != nil {
		log.Println("[WARN] No noroute.html to read noroute html from")
	}
	route.SetHTML(string(data))
	return make(chan string)
}
