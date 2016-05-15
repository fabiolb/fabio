package consul

import (
	"errors"
	"log"
	"strings"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/registry"

	"github.com/hashicorp/consul/api"
)

// be is an implementation of a registry backend for consul.
type be struct {
	c         *api.Client
	dc        string
	cfg       *config.Consul
	serviceID string
}

func NewBackend(cfg *config.Consul) (registry.Backend, error) {
	// create a reusable client
	c, err := api.NewClient(&api.Config{Address: cfg.Addr, Scheme: "http", Token: cfg.Token})
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
	return &be{c: c, dc: dc, cfg: cfg}, nil
}

func (b *be) Register() error {
	if !b.cfg.Register {
		log.Printf("[INFO] consul: Not registering fabio in consul")
		return nil
	}

	service, err := serviceRegistration(b.cfg.ServiceAddr, b.cfg.ServiceName, b.cfg.ServiceTags, b.cfg.CheckInterval, b.cfg.CheckTimeout)
	if err != nil {
		return err
	}
	if err := b.c.Agent().ServiceRegister(service); err != nil {
		return err
	}

	log.Printf("[INFO] consul: Registered fabio with id %q", service.ID)
	log.Printf("[INFO] consul: Registered fabio with address %q", b.cfg.ServiceAddr)
	log.Printf("[INFO] consul: Registered fabio with tags %q", strings.Join(b.cfg.ServiceTags, ","))
	log.Printf("[INFO] consul: Registered fabio with health check to %q", service.Check.HTTP)

	b.serviceID = service.ID
	return nil
}

func (b *be) Deregister() error {
	log.Printf("[INFO] consul: Deregistering fabio")
	return b.c.Agent().ServiceDeregister(b.serviceID)
}

func (b *be) ReadManual() (value string, version uint64, err error) {
	// we cannot rely on the value provided by WatchManual() since
	// someone has to call that method first to kick off the go routine.
	return getKV(b.c, b.cfg.KVPath, 0)
}

func (b *be) WriteManual(value string, version uint64) (ok bool, err error) {
	// try to create the key first by using version 0
	if ok, err = putKV(b.c, b.cfg.KVPath, value, 0); ok {
		return
	}

	// then try the CAS update
	return putKV(b.c, b.cfg.KVPath, value, version)
}

func (b *be) WatchServices() chan string {
	log.Printf("[INFO] consul: Using dynamic routes")
	log.Printf("[INFO] consul: Using tag prefix %q", b.cfg.TagPrefix)

	svc := make(chan string)
	go watchServices(b.c, b.cfg.TagPrefix, svc)
	return svc
}

func (b *be) WatchManual() chan string {
	log.Printf("[INFO] consul: Watching KV path %q", b.cfg.KVPath)

	kv := make(chan string)
	go watchKV(b.c, b.cfg.KVPath, kv)
	return kv
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
