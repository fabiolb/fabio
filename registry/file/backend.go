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
	routes, err := ioutil.ReadFile(cfg.RoutesPath)
	if err != nil {
		log.Println("[ERROR] Cannot read routes from ", cfg.RoutesPath)
		return nil, err
	}
	noroutehtml, err := ioutil.ReadFile(cfg.NoRouteHTMLPath)
	if err != nil {
		log.Println("[ERROR] Cannot read no route HTML from ", cfg.NoRouteHTMLPath)
		return nil, err
	}
	staticCfg := &config.Static{
		NoRouteHTML: string(noroutehtml),
		Routes:      string(routes),
	}
	return static.NewBackend(staticCfg)
}
