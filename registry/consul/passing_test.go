package consul

import (
	"reflect"
	"testing"

	"github.com/eBay/fabio/_third_party/github.com/hashicorp/consul/api"
)

func TestPassingServices(t *testing.T) {
	var (
		serfPass     = &api.HealthCheck{Node: "node", CheckID: "serfHealth", Status: "passing"}
		serfFail     = &api.HealthCheck{Node: "node", CheckID: "serfHealth", Status: "critical"}
		svc1Pass     = &api.HealthCheck{Node: "node", CheckID: "service:abc", Status: "passing", ServiceName: "abc", ServiceID: "abc-1"}
		svc2Pass     = &api.HealthCheck{Node: "node", CheckID: "my-check-id", Status: "passing", ServiceName: "def", ServiceID: "def-1"}
		svc1Maint    = &api.HealthCheck{Node: "node", CheckID: "_service_maintenance:abc-1", Status: "critical", ServiceName: "abc", ServiceID: "abc-1"}
		svc1ID2Maint = &api.HealthCheck{Node: "node", CheckID: "_service_maintenance:abc-2", Status: "critical", ServiceName: "abc", ServiceID: "abc-2"}
		nodeMaint    = &api.HealthCheck{Node: "node", CheckID: "_node_maintenance", Status: "critical"}
	)

	tests := []struct {
		in, out []*api.HealthCheck
	}{
		{nil, nil},
		{[]*api.HealthCheck{}, nil},
		{[]*api.HealthCheck{svc1Pass}, []*api.HealthCheck{svc1Pass}},
		{[]*api.HealthCheck{svc1Pass, svc2Pass}, []*api.HealthCheck{svc1Pass, svc2Pass}},
		{[]*api.HealthCheck{serfPass, svc1Pass}, []*api.HealthCheck{svc1Pass}},
		{[]*api.HealthCheck{serfFail, svc1Pass}, nil},
		{[]*api.HealthCheck{nodeMaint, svc1Pass}, nil},
		{[]*api.HealthCheck{svc1Maint, svc1Pass}, nil},
		{[]*api.HealthCheck{nodeMaint, svc1Maint, svc1Pass}, nil},
		{[]*api.HealthCheck{serfFail, nodeMaint, svc1Maint, svc1Pass}, nil},
		{[]*api.HealthCheck{svc1ID2Maint, svc1Pass}, []*api.HealthCheck{svc1Pass}},
		{[]*api.HealthCheck{svc1Maint, svc1Pass, svc2Pass}, []*api.HealthCheck{svc2Pass}},
	}

	for i, tt := range tests {
		if got, want := passingServices(tt.in), tt.out; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}
