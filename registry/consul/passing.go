package consul

import (
	"log"
	"strings"

	"github.com/eBay/fabio/_third_party/github.com/hashicorp/consul/api"
)

// passingServices filters out health checks for services which have
// passing health checks and where the neither the service instance itself
// nor the node is in maintenance mode.
func passingServices(checks []*api.HealthCheck) []*api.HealthCheck {
	var p []*api.HealthCheck
	for _, svc := range checks {
		// look at service checks only
		if !strings.HasPrefix(svc.CheckID, "service:") {
			continue
		}

		if svc.Status != "passing" {
			continue
		}

		// node or service in maintenance mode?
		for _, c := range checks {
			if c.CheckID == "serfHealth" && c.Node == svc.Node && c.Status == "critical" {
				log.Printf("[INFO] consul: Skipping service %q since agent on node %q is down: %s", svc.ServiceID, svc.Node, c.Output)
				goto skip
			}
			if c.CheckID == "_node_maintenance" && c.Node == svc.Node {
				log.Printf("[INFO] consul: Skipping service %q since node %q is in maintenance mode: %s", svc.ServiceID, svc.Node, c.Output)
				goto skip
			}
			if c.CheckID == "_service_maintenance:"+svc.ServiceID && c.Status == "critical" {
				log.Printf("[INFO] consul: Skipping service %q since it is in maintenance mode: %s", svc.ServiceID, c.Output)
				goto skip
			}
		}

		p = append(p, svc)

	skip:
	}

	return p
}
