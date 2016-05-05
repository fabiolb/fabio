package consul

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/eBay/fabio/config"
	"github.com/hashicorp/consul/api"
)

func serviceRegistration(addr, name string, tags []string, interval, timeout time.Duration) (*api.AgentServiceRegistration, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	ipstr, portstr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portstr)
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(ipstr)
	if ip == nil {
		ip, err = config.LocalIP()
		if err != nil {
			return nil, err
		}
		if ip == nil {
			return nil, errors.New("no local ip")
		}
	}

	serviceID := fmt.Sprintf("%s-%s-%d", name, hostname, port)

	checkURL := fmt.Sprintf("http://%s:%d/health", ip, port)
	if ip.To16() != nil {
		checkURL = fmt.Sprintf("http://[%s]:%d/health", ip, port)
	}

	service := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    name,
		Address: ip.String(),
		Port:    port,
		Tags:    tags,
		Check: &api.AgentServiceCheck{
			HTTP:     checkURL,
			Interval: interval.String(),
			Timeout:  timeout.String(),
		},
	}

	return service, nil
}
