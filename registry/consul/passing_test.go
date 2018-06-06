package consul

import (
	"reflect"
	"testing"

	"github.com/hashicorp/consul/api"
)

func TestPassingServices(t *testing.T) {
	var (
		serfPass      = &api.HealthCheck{Node: "node", CheckID: "serfHealth", Status: "passing"}
		serfFail      = &api.HealthCheck{Node: "node", CheckID: "serfHealth", Status: "critical"}
		svc1Pass      = &api.HealthCheck{Node: "node", CheckID: "service:abc", Status: "passing", ServiceName: "abc", ServiceID: "abc-1"}
		svc1Chk2Warn  = &api.HealthCheck{Node: "node", CheckID: "service:abc", Status: "warning", ServiceName: "abc", ServiceID: "abc-1"}
		svc1Node2Pass = &api.HealthCheck{Node: "node2", CheckID: "service:abc", Status: "passing", ServiceName: "abc", ServiceID: "abc-1"}
		svc1Warn      = &api.HealthCheck{Node: "node", CheckID: "service:abc", Status: "warning", ServiceName: "abc", ServiceID: "abc-2"}
		svc1Crit      = &api.HealthCheck{Node: "node", CheckID: "service:abc", Status: "critical", ServiceName: "abc", ServiceID: "abc-3"}
		svc2Pass      = &api.HealthCheck{Node: "node", CheckID: "my-check-id", Status: "passing", ServiceName: "def", ServiceID: "def-1"}
		svc1Maint     = &api.HealthCheck{Node: "node", CheckID: "_service_maintenance:abc-1", Status: "critical", ServiceName: "abc", ServiceID: "abc-1"}
		svc1ID2Maint  = &api.HealthCheck{Node: "node", CheckID: "_service_maintenance:abc-2", Status: "critical", ServiceName: "abc", ServiceID: "abc-2"}
		nodeMaint     = &api.HealthCheck{Node: "node", CheckID: "_node_maintenance", Status: "critical"}
	)

	tests := []struct {
		name    string
		strict  bool
		status  []string
		in, out []*api.HealthCheck
	}{
		{
			"expect no passing checks if checks array is nil",
			false, []string{"passing"}, nil, nil,
		},
		{
			"expect no passing checks if checks array is empty",
			false, []string{"passing"}, []*api.HealthCheck{}, nil,
		},
		{
			"expect check to pass if it has a matching status",
			false, []string{"passing"}, []*api.HealthCheck{svc1Pass}, []*api.HealthCheck{svc1Pass},
		},
		{
			"expect all checks to pass if they have a matching status",
			false, []string{"passing"}, []*api.HealthCheck{svc1Pass, svc2Pass}, []*api.HealthCheck{svc1Pass, svc2Pass},
		},
		{
			"expect that internal consul checks are filtered out",
			false, []string{"passing"}, []*api.HealthCheck{serfPass, svc1Pass}, []*api.HealthCheck{svc1Pass},
		},
		{
			"expect no passing checks if consul agent is unhealthy",
			false, []string{"passing"}, []*api.HealthCheck{serfFail, svc1Pass}, nil,
		},
		{
			"expect no passing checks if node is in maintenance mode",
			false, []string{"passing"}, []*api.HealthCheck{nodeMaint, svc1Pass}, nil,
		},
		{
			"expect no passing check if corresponding service is in maintenance mode",
			false, []string{"passing"}, []*api.HealthCheck{svc1Maint, svc1Pass}, nil,
		},
		{
			"expect no passing check if node and service are in maintenance mode",
			false, []string{"passing"}, []*api.HealthCheck{nodeMaint, svc1Maint, svc1Pass}, nil,
		},
		{
			"expect no passing check if agent is unhealthy or node and service are in maintenance mode",
			false, []string{"passing"}, []*api.HealthCheck{serfFail, nodeMaint, svc1Maint, svc1Pass}, nil,
		},
		{
			"expect check of service which is not in maintenance mode to pass if another instance of same service is in maintenance mode",
			false, []string{"passing"}, []*api.HealthCheck{svc1ID2Maint, svc1Pass}, []*api.HealthCheck{svc1Pass},
		},
		{
			"expect that no checks of a service which is in maintenance mode are returned even if it has a passing check",
			false, []string{"passing"}, []*api.HealthCheck{svc1Maint, svc1Pass, svc2Pass}, []*api.HealthCheck{svc2Pass},
		},
		{
			"expect that a service's failing check does not affect a healthy instance of same service running on different node",
			false, []string{"passing"}, []*api.HealthCheck{svc1Crit, svc1Node2Pass}, []*api.HealthCheck{svc1Node2Pass},
		},
		{
			"service in maintenance mode does not affect healthy service running on different node",
			false, []string{"passing"}, []*api.HealthCheck{svc1Maint, svc1Node2Pass}, []*api.HealthCheck{svc1Node2Pass},
		},
		{
			"expect that internal consul check and failing check are not returned",
			false, []string{"passing", "warning"}, []*api.HealthCheck{serfPass, svc1Pass, svc1Crit}, []*api.HealthCheck{svc1Pass},
		},
		{
			"expect that internal consul check is filtered out and check with warning is passing",
			false, []string{"passing", "warning"}, []*api.HealthCheck{serfPass, svc1Warn, svc1Crit}, []*api.HealthCheck{svc1Warn},
		},
		{
			"expect that warning and passing non-internal checks are returned",
			false, []string{"passing", "warning"}, []*api.HealthCheck{serfPass, svc1Pass, svc1Warn}, []*api.HealthCheck{svc1Pass, svc1Warn},
		},
		{
			"expect that warning und passing non-internal checks are returned",
			false, []string{"passing", "warning"}, []*api.HealthCheck{serfPass, svc1Warn, svc1Crit, svc1Pass}, []*api.HealthCheck{svc1Warn, svc1Pass},
		},
		{
			"in non-strict mode, expect that checks which belong to same service are passing, if at least one of them is passing",
			false, []string{"passing"}, []*api.HealthCheck{svc1Pass, svc1Chk2Warn}, []*api.HealthCheck{svc1Pass, svc1Chk2Warn},
		},
		{
			"in strict mode, expect that no checks which belong to same service are passing, if not all of them are passing",
			true, []string{"passing"}, []*api.HealthCheck{svc1Pass, svc1Chk2Warn}, nil,
		},
		{
			"in strict mode, expect that a failing check of one service does not affect a different service's passing check",
			true, []string{"passing"}, []*api.HealthCheck{svc1Pass, svc1Chk2Warn, svc2Pass}, []*api.HealthCheck{svc2Pass},
		},
		{
			"in strict mode, expect a check to pass if all of the other checks that belong to the same service are passing",
			true, []string{"passing", "warning"}, []*api.HealthCheck{svc1Pass, svc1Chk2Warn}, []*api.HealthCheck{svc1Pass, svc1Chk2Warn},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, want := passingServices(tt.in, tt.status, tt.strict), tt.out; !reflect.DeepEqual(got, want) {
				t.Errorf("got %v want %v", got, want)
			}
		})
	}
}
