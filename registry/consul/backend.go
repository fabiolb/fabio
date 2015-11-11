package consul

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/registry"

	"github.com/eBay/fabio/_third_party/github.com/hashicorp/consul/api"
)

// be is an implementation of a registry backend for consul.
type be struct {
	c   *api.Client
	dc  string
	cfg *config.Consul
}

func NewBackend(cfg *config.Consul) (registry.Backend, error) {
	// create a reusable client
	c, err := api.NewClient(&api.Config{Address: cfg.Addr, Scheme: "http"})
	if err != nil {
		return nil, err
	}

	// ping the agent
	dc, err := datacenter(c)
	if err != nil {
		return nil, err
	}

	// we're good
	log.Printf("[INFO] consul: Connecting to %q in datacenter %q", cfg.Addr, dc)
	log.Printf("[INFO] consul: UI is on %q", cfg.URL)
	return &be{c, dc, cfg}, nil
}

func (b *be) ConfigURL() string {
	return fmt.Sprintf("%sui/#/%s/kv%s/edit", b.cfg.URL, b.dc, b.cfg.KVPath)
}

func (b *be) Watch() chan string {
	log.Printf("[INFO] consul: Using dynamic routes")
	log.Printf("[INFO] consul: Using tag prefix %q", b.cfg.TagPrefix)
	log.Printf("[INFO] consul: Watching KV path %q", b.cfg.KVPath)

	kv := make(chan []string)
	svc := make(chan []string)
	routes := make(chan string)
	go watchKV(b.c, b.cfg.KVPath, kv)
	go watchServices(b.c, b.cfg.TagPrefix, svc)
	go watchRoutes(kv, svc, routes)
	return routes
}

func watchRoutes(kv, svc chan []string, routes chan string) {
	var (
		last   string
		kvcfg  []string
		svccfg []string
	)

	for {
		select {
		case kvcfg = <-kv:
		case svccfg = <-svc:
		}

		if len(kvcfg)+len(svccfg) == 0 {
			continue
		}

		// build next config by appending kv config to service config
		// order matters
		next := strings.Join(append(svccfg, kvcfg...), "\n")
		if next == last {
			continue
		}

		routes <- next
		last = next
	}
}

// datacenter returns the datacenter of the local agent
func datacenter(c *api.Client) (string, error) {
	self, err := c.Agent().Self()
	if err != nil {
		return "", err
	}

	cfg, ok := self["Config"]
	if !ok {
		return "", errors.New("consul: self.Config not found")
	}
	dc, ok := cfg["Datacenter"].(string)
	if !ok {
		return "", errors.New("consul: self.Datacenter not found")
	}
	return dc, nil
}
