package consul

import (
	"reflect"
	"testing"

	"github.com/hashicorp/consul/api"
)

func TestChecksWithTagPrefix(t *testing.T) {
	serf := &api.HealthCheck{CheckID: "serfHealth"}
	nodeMaintenance := &api.HealthCheck{CheckID: "_node_maintenance"}
	serviceMaintenance := &api.HealthCheck{CheckID: "_service_maintenance:svc"}
	legacy := &api.HealthCheck{CheckID: "legacy", ServiceTags: []string{"prod", "urlprefix-/users"}}
	metadata := &api.HealthCheck{CheckID: "metadata", ServiceTags: []string{"fabio", "prod"}}
	both := &api.HealthCheck{CheckID: "both", ServiceTags: []string{"fabio", "urlprefix-/items"}}
	custom := &api.HealthCheck{CheckID: "custom", ServiceTags: []string{"route-/custom"}}
	unrelated := &api.HealthCheck{CheckID: "unrelated", ServiceTags: []string{"prod", "fabio-extra"}}
	tagless := &api.HealthCheck{CheckID: "tagless"}
	checks := api.HealthChecks{
		serf, nodeMaintenance, serviceMaintenance, legacy, metadata, both, custom, unrelated, tagless,
	}

	for _, tt := range []struct {
		name   string
		prefix string
		want   api.HealthChecks
	}{
		{
			name:   "legacy and metadata discovery with maintenance",
			prefix: "urlprefix-",
			want:   api.HealthChecks{serf, nodeMaintenance, serviceMaintenance, legacy, metadata, both},
		},
		{
			name:   "custom route prefix",
			prefix: "route-",
			want:   api.HealthChecks{serf, nodeMaintenance, serviceMaintenance, metadata, both, custom},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := checksWithTagPrefix(tt.prefix, checks); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("checksWithTagPrefix(%q) = %#v, want %#v", tt.prefix, got, tt.want)
			}
		})
	}
}
