package consul

import (
	"log"
	"strings"

	"github.com/eBay/fabio/route"

	"github.com/eBay/fabio/_third_party/github.com/hashicorp/consul/api"
)

type Watcher struct {
	client     *api.Client
	tagPrefix  string
	configPath string
}

func NewWatcher(tagPrefix, configPath string) (*Watcher, error) {
	client, err := api.NewClient(&api.Config{Address: Addr, Scheme: "http"})
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		client:     client,
		tagPrefix:  tagPrefix,
		configPath: configPath,
	}
	return w, nil
}

func (w *Watcher) Watch() {
	var (
		auto   []string
		manual []string
		t      route.Table
		err    error

		autoConfig   = make(chan []string)
		manualConfig = make(chan []string)
	)

	go watchAutoConfig(w.client, w.tagPrefix, autoConfig)
	go watchManualConfig(w.client, w.configPath, manualConfig)

	for {
		select {
		case auto = <-autoConfig:
		case manual = <-manualConfig:
		}

		if len(auto) == 0 && len(manual) == 0 {
			continue
		}

		input := strings.Join(append(auto, manual...), "\n")
		t, err = route.ParseString(input)
		if err != nil {
			log.Printf("[WARN] %s", err)
			continue
		}
		route.SetTable(t)
	}
}
