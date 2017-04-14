package consul

import (
	"reflect"
	"testing"

	"github.com/hashicorp/consul/api"
)

func TestPassingServices(t *testing.T) {
	var (
		serfPass     = &api.HealthCheck{Node: "node", CheckID: "serfHealth", Status: "passing"}
		serfFail     = &api.HealthCheck{Node: "node", CheckID: "serfHealth", Status: "critical"}
		svc1Pass     = &api.HealthCheck{Node: "node", CheckID: "service:abc", Status: "passing", ServiceName: "abc", ServiceID: "abc-1"}
		svc1Warn     = &api.HealthCheck{Node: "node", CheckID: "service:abc", Status: "warning", ServiceName: "abc", ServiceID: "abc-2"}
		svc1Crit     = &api.HealthCheck{Node: "node", CheckID: "service:abc", Status: "critical", ServiceName: "abc", ServiceID: "abc-3"}
		svc2Pass     = &api.HealthCheck{Node: "node", CheckID: "my-check-id", Status: "passing", ServiceName: "def", ServiceID: "def-1"}
		svc1Maint    = &api.HealthCheck{Node: "node", CheckID: "_service_maintenance:abc-1", Status: "critical", ServiceName: "abc", ServiceID: "abc-1"}
		svc1ID2Maint = &api.HealthCheck{Node: "node", CheckID: "_service_maintenance:abc-2", Status: "critical", ServiceName: "abc", ServiceID: "abc-2"}
		nodeMaint    = &api.HealthCheck{Node: "node", CheckID: "_node_maintenance", Status: "critical"}
	)

	tests := []struct {
		status  []string
		in, out []*api.HealthCheck
	}{
		{[]string{"passing"}, nil, nil},
		{[]string{"passing"}, []*api.HealthCheck{}, nil},
		{[]string{"passing"}, []*api.HealthCheck{svc1Pass}, []*api.HealthCheck{svc1Pass}},
		{[]string{"passing"}, []*api.HealthCheck{svc1Pass, svc2Pass}, []*api.HealthCheck{svc1Pass, svc2Pass}},
		{[]string{"passing"}, []*api.HealthCheck{serfPass, svc1Pass}, []*api.HealthCheck{svc1Pass}},
		{[]string{"passing"}, []*api.HealthCheck{serfFail, svc1Pass}, nil},
		{[]string{"passing"}, []*api.HealthCheck{nodeMaint, svc1Pass}, nil},
		{[]string{"passing"}, []*api.HealthCheck{svc1Maint, svc1Pass}, nil},
		{[]string{"passing"}, []*api.HealthCheck{nodeMaint, svc1Maint, svc1Pass}, nil},
		{[]string{"passing"}, []*api.HealthCheck{serfFail, nodeMaint, svc1Maint, svc1Pass}, nil},
		{[]string{"passing"}, []*api.HealthCheck{svc1ID2Maint, svc1Pass}, []*api.HealthCheck{svc1Pass}},
		{[]string{"passing"}, []*api.HealthCheck{svc1Maint, svc1Pass, svc2Pass}, []*api.HealthCheck{svc2Pass}},
		{[]string{"passing", "warning"}, []*api.HealthCheck{serfPass, svc1Pass, svc1Crit}, []*api.HealthCheck{svc1Pass}},
		{[]string{"passing", "warning"}, []*api.HealthCheck{serfPass, svc1Warn, svc1Crit}, []*api.HealthCheck{svc1Warn}},
		{[]string{"passing", "warning"}, []*api.HealthCheck{serfPass, svc1Pass, svc1Warn}, []*api.HealthCheck{svc1Pass, svc1Warn}},
		{[]string{"passing", "warning"}, []*api.HealthCheck{serfPass, svc1Warn, svc1Crit, svc1Pass}, []*api.HealthCheck{svc1Warn, svc1Pass}},
	}

	for i, tt := range tests {
		if got, want := passingServices(tt.in, tt.status), tt.out; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}
