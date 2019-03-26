package consul

import (
	"reflect"
	"testing"

	"github.com/hashicorp/consul/api"
)

func TestMatchingServices(t *testing.T) {
	var (
		noTags      = &api.HealthCheck{}
		noRoute     = &api.HealthCheck{ServiceTags: []string{}}
		singleRoute = &api.HealthCheck{ServiceTags: []string{"route-/something"}}
	)

	tests := []struct {
		name    string
		prefix  string
		in, out []*api.HealthCheck
	}{
		{
			"expect match when tag starts correctly",
			"route-", []*api.HealthCheck{singleRoute}, []*api.HealthCheck{singleRoute},
		},
		{
			"expect no match when tag does not starts correctly",
			"xxxxx-", []*api.HealthCheck{singleRoute}, nil,
		},
		{
			"expect no match when no routes",
			"route-", []*api.HealthCheck{noRoute}, nil,
		},
		{
			"expect no match when no tags",
			"route-", []*api.HealthCheck{noTags}, nil,
		},
		{
			"expect no match when no input",
			"route-", []*api.HealthCheck{}, nil,
		},
		{
			"expect no match when no prefix",
			"", []*api.HealthCheck{}, nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, want := matchingServices(tt.prefix, tt.in), tt.out; !reflect.DeepEqual(got, want) {
				t.Errorf("got %v want %v", got, want)
			}
		})
	}
}
