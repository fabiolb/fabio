package consul

import (
	"log"
	"strings"

	"github.com/hashicorp/consul/api"
)

func passingServices(checks []*api.HealthCheck, status []string, strict bool) []*api.HealthCheck {
	var p []*api.HealthCheck
	for _, svc := range checks {
		if !isServiceCheck(svc) {
			continue
		}
		total, passing := countChecks(svc, checks, status)
		if passing == 0 {
			continue
		}
		if strict && total != passing {
			continue
		}
		if isAgentCritical(svc, checks) {
			continue
		}
		if isNodeInMaintenance(svc, checks) {
			continue
		}
		if isServiceInMaintenance(svc, checks) {
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

// isAgentCritical returns true if the agent on the node on which the service
// runs is critical.
func isAgentCritical(svc *api.HealthCheck, checks []*api.HealthCheck) bool {
	for _, c := range checks {
		if svc.Node == c.Node && c.CheckID == "serfHealth" && c.Status == "critical" {
			log.Printf("[DEBUG] consul: Skipping service %q since agent on node %q is down: %s", c.ServiceID, c.Node, c.Output)
			return true
		}
	}
	return false
}

// isNodeInMaintenance returns true if the node on which the service runs is in
// maintenance mode.
func isNodeInMaintenance(svc *api.HealthCheck, checks []*api.HealthCheck) bool {
	for _, c := range checks {
		if svc.Node == c.Node && c.CheckID == "_node_maintenance" {
			log.Printf("[DEBUG] consul: Skipping service %q since node %q is in maintenance mode: %s", c.ServiceID, c.Node, c.Output)
			return true
		}
	}
	return false
}

// isServiceInMaintenance returns true if the service instance is in
// maintenance mode.
func isServiceInMaintenance(svc *api.HealthCheck, checks []*api.HealthCheck) bool {
	for _, c := range checks {
		if svc.Node == c.Node && c.CheckID == "_service_maintenance:"+svc.ServiceID && c.Status == "critical" {
			log.Printf("[DEBUG] consul: Skipping service %q since it is in maintenance mode: %s", svc.ServiceID, c.Output)
			return true
		}
	}
	return false
}

// countChecks counts the number of service checks exist for a given service
// and how many of them are passing.
func countChecks(svc *api.HealthCheck, checks []*api.HealthCheck, status []string) (total int, passing int) {
	for _, c := range checks {
		if svc.Node == c.Node && svc.ServiceID == c.ServiceID {
			total++
			if hasStatus(c, status) {
				passing++
			}
		}
	}
	return
}

// hasStatus returns true if the health check status is one of the given
// values.
func hasStatus(c *api.HealthCheck, status []string) bool {
	for _, s := range status {
		if c.Status == s {
			return true
		}
	}
	return false
}
