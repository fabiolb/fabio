package consul

import (
	"github.com/eBay/fabio/mdllog"
	"strings"

	"github.com/hashicorp/consul/api"
)

// passingServices filters out health checks for services which have
// passing health checks and where the neither the service instance itself
// nor the node is in maintenance mode.
func passingServices(checks []*api.HealthCheck, status []string) []*api.HealthCheck {
	var p []*api.HealthCheck
	for _, svc := range checks {
		// first filter out non-service checks
		if svc.ServiceID == "" || svc.CheckID == "serfHealth" || svc.CheckID == "_node_maintenance" || strings.HasPrefix("_service_maintenance:", svc.CheckID) {
			continue
		}

		// then make sure the service health check is passing
		if !contains(status, svc.Status) {
			continue
		}

		// then check whether the agent is still alive and both the
		// node and the service are not in maintenance mode.
		for _, c := range checks {
			if c.CheckID == "serfHealth" && c.Node == svc.Node && c.Status == "critical" {
				mdllog.Info.Printf("[INFO] consul: Skipping service %q since agent on node %q is down: %s", svc.ServiceID, svc.Node, c.Output)
				goto skip
			}
			if c.CheckID == "_node_maintenance" && c.Node == svc.Node {
				mdllog.Info.Printf("[INFO] consul: Skipping service %q since node %q is in maintenance mode: %s", svc.ServiceID, svc.Node, c.Output)
				goto skip
			}
			if c.CheckID == "_service_maintenance:"+svc.ServiceID && c.Status == "critical" {
				mdllog.Info.Printf("[INFO] consul: Skipping service %q since it is in maintenance mode: %s", svc.ServiceID, c.Output)
				goto skip
			}
		}

		p = append(p, svc)

	skip:
	}

	return p
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
