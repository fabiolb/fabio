package consul

import (
	"log"
	"slices"
	"strings"

	"github.com/hashicorp/consul/api"
)

// passingServices takes a list of Consul Health Checks and only returns ones where the overall health of
// the Service Instance is passing. This includes the health of the Node that the Services Instance runs on.
func passingServices(checks []*api.HealthCheck, status []string, strict bool) []*api.HealthCheck {
	var p []*api.HealthCheck

CHECKS:
	for _, svc := range checks {
		if !isServiceCheck(svc) {
			continue
		}
		var total, passing int

		for _, c := range checks {
			if svc.Node == c.Node {
				if svc.ServiceID == c.ServiceID {
					total++
					if hasStatus(c, status) {
						passing++
					}
				}
				if c.CheckID == "serfHealth" && c.Status == "critical" {
					log.Printf("[DEBUG] consul: Skipping service %q since agent on node %q is down: %s", c.ServiceID, c.Node, c.Output)
					continue CHECKS
				}
				if c.CheckID == "_node_maintenance" {
					log.Printf("[DEBUG] consul: Skipping service %q since node %q is in maintenance mode: %s", c.ServiceID, c.Node, c.Output)
					continue CHECKS
				}
				if c.CheckID == "_service_maintenance:"+svc.ServiceID && c.Status == "critical" {
					log.Printf("[DEBUG] consul: Skipping service %q since it is in maintenance mode: %s", svc.ServiceID, c.Output)
					continue CHECKS
				}
			}
		}

		if passing == 0 {
			continue
		}
		if strict && total != passing {
			continue
		}

		p = append(p, svc)
	}

	return p
}

// isServiceCheck returns true if the health check is a valid service check.
func isServiceCheck(c *api.HealthCheck) bool {
	return c.ServiceID != "" &&
		c.CheckID != "serfHealth" &&
		c.CheckID != "_node_maintenance" &&
		!strings.HasPrefix(c.CheckID, "_service_maintenance:")
}

// hasStatus returns true if the health check status is one of the given
// values.
func hasStatus(c *api.HealthCheck, status []string) bool {
	return slices.Contains(status, c.Status)
}
