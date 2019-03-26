package consul

import (
	"strings"

	"github.com/hashicorp/consul/api"
)

func matchingServices(prefix string, checks []*api.HealthCheck) []*api.HealthCheck {
	var matching []*api.HealthCheck
	for _, hc := range checks {
		if len(hc.ServiceTags) == 0 {
			continue
		}
		for _, tag := range hc.ServiceTags {
			if strings.HasPrefix(tag, prefix) {
				matching = append(matching, hc)
				continue
			}
		}
	}

	return matching
}
