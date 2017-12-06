// Package file implements a simple file based registry
// backend which reads the routes from a file once.
package file

import (
	"io/ioutil"
	"log"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/registry"
	"github.com/fabiolb/fabio/registry/static"
)

func NewBackend(cfg *config.File) (registry.Backend, error) {
	data, err := ioutil.ReadFile(cfg.Path)
	if err != nil {
		log.Println("[ERROR] Cannot read routes from ", cfg.Path)
		return nil, err
	}
	staticCfg := config.Static{cfg.NoRouteHTMLPath, string(data)}
	return static.NewBackend(&staticCfg)
}
