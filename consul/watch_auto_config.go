package consul

import (
	"fmt"
	"log"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/eBay/fabio/_third_party/github.com/hashicorp/consul/api"
)

// watchAutoConfig monitors the consul health checks and creates a new configuration
// on every change.
func watchAutoConfig(client *api.Client, tagPrefix string, config chan []string) {
	var lastIndex uint64

	for {
		q := &api.QueryOptions{RequireConsistent: true, WaitIndex: lastIndex}
		checks, meta, err := client.Health().State("any", q)
		if err != nil {
			log.Printf("[WARN] Error fetching health state. %v", err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("[INFO] Health changed to #%d", meta.LastIndex)
		config <- servicesConfig(client, passingServices(checks), tagPrefix)
		lastIndex = meta.LastIndex
	}
}

// servicesConfig determines which service instances have passing health checks
// and then finds the ones which have tags with the right prefix to build the config from.
func servicesConfig(client *api.Client, checks []*api.HealthCheck, tagPrefix string) []string {
	// map service name to list of service passing for which the health check is ok
	m := map[string]map[string]bool{}
	for _, check := range checks {
		name, id := check.ServiceName, check.ServiceID

		if _, ok := m[name]; !ok {
			m[name] = map[string]bool{}
		}
		m[name][id] = true
	}

	var config []string
	for name, passing := range m {
		cfg := serviceConfig(client, name, passing, tagPrefix)
		config = append(config, cfg...)
	}

	// sort config in reverse order to sort most specific config to the top
	sort.Sort(sort.Reverse(sort.StringSlice(config)))

	return config
}

// serviceConfig constructs the config for all good instances of a single service.
func serviceConfig(client *api.Client, name string, passing map[string]bool, tagPrefix string) (config []string) {
	if name == "" || len(passing) == 0 {
		return nil
	}

	q := &api.QueryOptions{RequireConsistent: true}
	svcs, _, err := client.Catalog().Service(name, "", q)
	if err != nil {
		log.Printf("[WARN] Error getting catalog service %s. %v", name, err)
		return nil
	}

	for _, svc := range svcs {
		// check if the instance is in the list of instances
		// which passed the health check
		if _, ok := passing[svc.ServiceID]; !ok {
			continue
		}

		for _, tag := range svc.ServiceTags {
			if host, path, ok := parseURLPrefixTag(tag, tagPrefix); ok {
				name, addr, port := svc.ServiceName, svc.ServiceAddress, svc.ServicePort
				if runtime.GOOS == "darwin" && !strings.Contains(addr, ".") {
					addr += ".local"
				}
				config = append(config, fmt.Sprintf("route add %s %s%s http://%s:%d/ tags %q", name, host, path, addr, port, strings.Join(svc.ServiceTags, ",")))
			}
		}
	}
	return config
}
