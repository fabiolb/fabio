package custom

import (
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/registry"
	"log"
)

type BE struct {
	cfg *config.CustomBE
}

func NewBackend(cfg *config.CustomBE) (registry.Backend, error) {
	return &BE{cfg}, nil
}

func (b *BE) Register(services []string) error {
	return nil
}

func (b *BE) Deregister(serviceName string) error {
	return nil
}

func (b *BE) DeregisterAll() error {
	return nil
}

func (b *BE) ManualPaths() ([]string, error) {
	return nil, nil
}

func (b *BE) ReadManual(string) (value string, version uint64, err error) {
	return "", 0, nil
}

func (b *BE) WriteManual(path string, value string, version uint64) (ok bool, err error) {
	return false, nil
}

func (b *BE) WatchServices() chan string {

	log.Printf("[INFO] custom: Using custom routes from %s", b.cfg.Host)
	ch := make(chan string, 1)
	go customRoutes(b.cfg, ch)
	return ch
}

func (b *BE) WatchManual() chan string {
	return make(chan string)
}

func (b *BE) WatchNoRouteHTML() chan string {
	ch := make(chan string, 1)
	//TODO figure out what to send back
	ch <- b.cfg.NoRouteHTML
	return ch
}
