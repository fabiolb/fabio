package consul

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/hashicorp/consul/api"
)

// ServiceMonitor generates fabio configurations from consul state.
type ServiceMonitor struct {
	client *api.Client
	config *config.Consul
	dc     string
	strict bool
}

func NewServiceMonitor(client *api.Client, config *config.Consul, dc string) *ServiceMonitor {
	return &ServiceMonitor{
		client: client,
		config: config,
		dc:     dc,
		strict: config.ChecksRequired == "all",
	}
}

// Watch monitors the consul health checks and sends a new
// configuration to the updates channel on every change.
func (w *ServiceMonitor) Watch(updates chan string) {
	var lastIndex uint64
	for {
		q := &api.QueryOptions{RequireConsistent: true, WaitIndex: lastIndex}
		checks, meta, err := w.client.Health().State("any", q)
		if err != nil {
			log.Printf("[WARN] consul: Error fetching health state. %v", err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("[DEBUG] consul: Health changed to #%d", meta.LastIndex)

		// determine which services have passing health checks
		passing := passingServices(checks, w.config.ServiceStatus, w.strict)

		// build the config for the passing services
		updates <- w.makeConfig(passing)

		// remember the last state and wait for the next change
		lastIndex = meta.LastIndex
	}
}

// makeCconfig determines which service instances have passing health checks
// and then finds the ones which have tags with the right prefix to build the config from.
func (w *ServiceMonitor) makeConfig(checks []*api.HealthCheck) string {
	// map service name to list of service passing for which the health check is ok
	m := map[string]map[string]bool{}
	for _, check := range checks {
		// Make the node part of the id, because according to the Consul docs
		// the ServiceID is unique per agent but not cluster wide
		// https://www.consul.io/api/agent/service.html#id
		name, id := check.ServiceName, fmt.Sprintf("%s.%s", check.Node, check.ServiceID)

		if _, ok := m[name]; !ok {
			m[name] = map[string]bool{}
		}
		m[name][id] = true
	}

	n := w.config.ServiceMonitors
	if n <= 0 {
		n = 1
	}

	sem := make(chan int, n)
	cfgs := make(chan []string, len(m))
	for name, passing := range m {
		name, passing := name, passing
		go func() {
			sem <- 1
			cfgs <- w.serviceConfig(name, passing)
			<-sem
		}()
	}

	var config []string
	for i := 0; i < len(m); i++ {
		cfg := <-cfgs
		config = append(config, cfg...)
	}

	// sort config in reverse order to sort most specific config to the top
	sort.Sort(sort.Reverse(sort.StringSlice(config)))

	return strings.Join(config, "\n")
}

// serviceConfig constructs the config for all good instances of a single service.
func (w *ServiceMonitor) serviceConfig(name string, passing map[string]bool) (config []string) {
	if name == "" || len(passing) == 0 {
		return nil
	}

	q := &api.QueryOptions{RequireConsistent: true}
	svcs, _, err := w.client.Catalog().Service(name, "", q)
	if err != nil {
		log.Printf("[WARN] consul: Error getting catalog service %s. %v", name, err)
		return nil
	}

	env := map[string]string{
		"DC": w.dc,
	}

	for _, svc := range svcs {
		// check if this instance passed the health check
		if _, ok := passing[svc.Node+"."+svc.ServiceID]; !ok {
			continue
		}

		r := routecmd{
			svc:    svc,
			env:    env,
			prefix: w.config.TagPrefix,
		}
		cmds := r.build()

		config = append(config, cmds...)
	}
	return config
}
