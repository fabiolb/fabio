// Package file implements a simple file based registry
// backend which reads the routes from a file.
//file content like registry.static.routes = route add web-svc /test http://127.0.0.1:8082
//registry.file.path = /home/zjj/fabio.txt
//registry.file.noroutehtmlpath = /home/zjj/404.html
package file

import (
	"os"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/registry"
)

type be struct {
	cfg             *config.File
	routesData      *filedata
	noRouteHTMLData *filedata
}

func NewBackend(cfg *config.File) (registry.Backend, error) {
	if _, err := os.Stat(cfg.RoutesPath); err != nil {
		return nil, err
	}

	b := &be{cfg: cfg, routesData: &filedata{path: cfg.RoutesPath}, noRouteHTMLData: &filedata{path: cfg.NoRouteHTMLPath}}
	return b, nil
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
	ch := make(chan string)
	go func() {
		for {
			readFile(b.routesData)
			ch <- b.routesData.content
			time.Sleep(b.cfg.Interval)
		}
	}()
	return ch
}

func (b *be) WatchManual() chan string {
	return make(chan string)
}

func (b *be) WatchNoRouteHTML() chan string {
	ch := make(chan string)
	if b.noRouteHTMLData.path != "" {
		go func() {
			for {
				readFile(b.noRouteHTMLData)
				ch <- b.noRouteHTMLData.content
				time.Sleep(b.cfg.Interval)
			}
		}()
	}
	return ch
}
