package consul

import (
	"errors"

	"github.com/eBay/fabio/_third_party/github.com/hashicorp/consul/api"
)

// Addr contains the host:port of the consul server
var Addr string

// Scheme contains the protocol used to connect to the consul server
var Scheme = "http"

// URL contains the base URL of the consul server
var URL string

// Datacenter returns the datacenter of the local agent
func Datacenter() (string, error) {
	client, err := newClient()
	if err != nil {
		return "", nil
	}

	self, err := client.Agent().Self()
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

func newClient() (*api.Client, error) {
	return api.NewClient(&api.Config{Address: Addr, Scheme: Scheme})
}
