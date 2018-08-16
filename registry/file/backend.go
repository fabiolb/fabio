// Package file implements a simple file based registry
// backend which reads the routes from a file.
//file content like registry.static.routes = route add web-svc /test http://127.0.0.1:8082
//registry.file.path = /home/zjj/fabio.txt
//registry.file.noroutehtmlpath = /home/zjj/404.html
package file

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/registry"
)

type be struct {
	cfg          *config.File
	routeMtime   time.Time
	norouteMtime time.Time
	Routes       string
	NoRouteHTML  string
	Interval     time.Duration
}

var (
	zero time.Time
)

func NewBackend(cfg *config.File) (registry.Backend, error) {
	b := &be{cfg: cfg, Interval: 2 * time.Second}
	if err := b.readRoute(); err != nil {
		return nil, err
	}
	if err := b.readNoRouteHtml(); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *be) readRoute() error {
	finfo, err := os.Stat(b.cfg.RoutesPath)
	if err != nil {
		log.Println("[ERROR] Cannot read routes stat from ", b.cfg.RoutesPath)
		return err
	}

	if b.routeMtime == zero || b.routeMtime != finfo.ModTime() {
		b.routeMtime = finfo.ModTime()
		routes, err := ioutil.ReadFile(b.cfg.RoutesPath)
		if err != nil {
			log.Println("[ERROR] Cannot read routes from ", b.cfg.RoutesPath)
			return err
		}
		b.Routes = string(routes)
	}
	return nil
}

func (b *be) readNoRouteHtml() error {
	if b.cfg.NoRouteHTMLPath != "" {
		finfo, err := os.Stat(b.cfg.NoRouteHTMLPath)
		if err != nil {
			log.Println("[ERROR] Cannot read no route HTML stat from ", b.cfg.NoRouteHTMLPath)
			return err
		}
		if b.norouteMtime == zero || b.norouteMtime != finfo.ModTime() {
			b.norouteMtime = finfo.ModTime()
			noroutehtml, err := ioutil.ReadFile(b.cfg.NoRouteHTMLPath)
			if err != nil {
				log.Println("[ERROR] Cannot read no route HTML from ", b.cfg.NoRouteHTMLPath)
				return err
			}
			b.NoRouteHTML = string(noroutehtml)
		}
	}
	return nil
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
	ch <- b.Routes
	go func() {
		for {
			b.readRoute()
			ch <- b.Routes
			time.Sleep(b.Interval)
		}
	}()
	return ch
}

func (b *be) WatchManual() chan string {
	return make(chan string)
}

func (b *be) WatchNoRouteHTML() chan string {
	ch := make(chan string, 1)
	ch <- b.NoRouteHTML
	go func() {
		for {
			b.readNoRouteHtml()
			ch <- b.NoRouteHTML
			time.Sleep(b.Interval)
		}
	}()
	return ch
}
