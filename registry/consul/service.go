package consul

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/hashicorp/consul/api"
	"sync"
)

// Channel used to pass data to serviceConfig when using Go Routines
type ServiceChannel struct {
	Client    *api.Client
	Name      string
	Passing   map[string]bool
	TagPrefix string
}

// watchServices monitors the consul health checks and creates a new configuration
// on every change.
func watchServices(client *api.Client, config *config.Consul, svcConfig chan string) {
	var lastIndex uint64
	var strict bool = strings.EqualFold("all", config.ChecksRequired)

	for {
		q := &api.QueryOptions{RequireConsistent: true, WaitIndex: lastIndex}
		checks, meta, err := client.Health().State("any", q)
		if err != nil {
			log.Printf("[WARN] consul: Error fetching health state. %v", err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("[DEBUG] consul: Health changed to #%d", meta.LastIndex)
		svcConfig <- servicesConfig(client, passingServices(checks, config.ServiceStatus, strict), config.TagPrefix, config.ConcurrentConsulRequests)
		lastIndex = meta.LastIndex
	}
}

// servicesConfig determines which service instances have passing health checks
// and then finds the ones which have tags with the right prefix to build the config from.
func servicesConfig(client *api.Client, checks []*api.HealthCheck, tagPrefix string, concurrentRequests int) string {
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

	//Create Buffered Channel
	serviceChan := make(chan ServiceChannel, concurrentRequests)

	//Create Wait Group
	var wg sync.WaitGroup
	//config is where the update strings are stored
	var config []string

	//Spin up Go Routines for getting service info from Consul
	for i := 1; i <= concurrentRequests; i++ {
		wg.Add(1)
		go serviceConfig(serviceChan, &wg, &config)
	}
	//Call serviceConfig Go Routines for every service
	for name, passing := range m {
		serviceChan <- ServiceChannel{Client: client, Name: name, Passing: passing, TagPrefix: tagPrefix}
	}

	close(serviceChan)
	wg.Wait()

	// sort config in reverse order to sort most specific config to the top
	sort.Sort(sort.Reverse(sort.StringSlice(config)))
	return strings.Join(config, "\n")
}

// serviceConfig constructs the config for all good instances of a single service.
//func serviceConfig(client *api.Client, name string, passing map[string]bool, tagPrefix string, ch chan []string)
func serviceConfig(ch chan ServiceChannel, wg *sync.WaitGroup, config *[]string) {

	defer wg.Done()
	for service := range ch {
		if service.Name == "" || len(service.Passing) == 0 {
			return
		}

		dc, err := datacenter(service.Client)
		if err != nil {
			log.Printf("[WARN] consul: Error getting datacenter. %s", err)
			return
		}

		q := &api.QueryOptions{RequireConsistent: true}
		svcs, _, err := service.Client.Catalog().Service(service.Name, "", q)
		if err != nil {
			log.Printf("[WARN] consul: Error getting catalog service %s. %v", service.Name, err)
			return
		}

		env := map[string]string{
			"DC": dc,
		}

		for _, svc := range svcs {
			// check if the instance is in the list of instances
			// which passed the health check
			if _, ok := service.Passing[fmt.Sprintf("%s.%s", svc.Node, svc.ServiceID)]; !ok {
				continue
			}

			// get all tags which do not have the tag prefix
			var svctags []string
			for _, tag := range svc.ServiceTags {
				if !strings.HasPrefix(tag, service.TagPrefix) {
					svctags = append(svctags, tag)
				}
			}

			// generate route commands
			for _, tag := range svc.ServiceTags {
				if route, opts, ok := parseURLPrefixTag(tag, service.TagPrefix, env); ok {
					name, addr, port := svc.ServiceName, svc.ServiceAddress, svc.ServicePort

					// use consul node address if service address is not set
					if addr == "" {
						addr = svc.Address
					}

					// add .local suffix on OSX for simple host names w/o domain
					if runtime.GOOS == "darwin" && !strings.Contains(addr, ".") && !strings.HasSuffix(addr, ".local") {
						addr += ".local"
					}

					// build route command
					weight := ""
					ropts := []string{}
					tags := strings.Join(svctags, ",")
					addr = net.JoinHostPort(addr, strconv.Itoa(port))
					dst := "http://" + addr + "/"
					for _, o := range strings.Fields(opts) {
						switch {
						case o == "proto=tcp":
							dst = "tcp://" + addr
						case o == "proto=https":
							dst = "https://" + addr
						case strings.HasPrefix(o, "weight="):
							weight = o[len("weight="):]
						case strings.HasPrefix(o, "redirect="):
							redir := strings.Split(o[len("redirect="):], ",")
							if len(redir) == 2 {
								dst = redir[1]
								ropts = append(ropts, fmt.Sprintf("redirect=%s", redir[0]))
							} else {
								log.Printf("[ERROR] Invalid syntax for redirect: %s. should be redirect=<code>,<url>", o)
								continue
							}
						default:
							ropts = append(ropts, o)
						}
					}

					cfg := "route add " + name + " " + route + " " + dst
					if weight != "" {
						cfg += " weight " + weight
					}
					if tags != "" {
						cfg += " tags " + strconv.Quote(tags)
					}
					if len(ropts) > 0 {
						cfg += " opts " + strconv.Quote(strings.Join(ropts, " "))
					}

					mutex := sync.Mutex{}
					mutex.Lock()
					*config = append(*config, cfg)
					mutex.Unlock()
				}
			}
		}
	}
}
