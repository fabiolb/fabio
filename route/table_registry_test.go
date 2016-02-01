package route

import (
	"reflect"
	"sort"
	"testing"

	"github.com/eBay/fabio/metrics"
)

func TestSyncRegistry(t *testing.T) {
	names := func() []string {
		var n []string
		metrics.ServiceRegistry.Each(func(name string, x interface{}) {
			n = append(n, name)
		})
		sort.Strings(n)
		return n
	}

	metrics.ServiceRegistry.UnregisterAll()

	tbl := make(Table)
	tbl.AddRoute("svc-a", "/aaa", "http://localhost:1234", 1, nil)
	tbl.AddRoute("svc-b", "/bbb", "http://localhost:5678", 1, nil)
	if got, want := names(), []string{"svc-a._./aaa.localhost_1234", "svc-b._./bbb.localhost_5678"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	tbl.DelRoute("svc-b", "/bbb", "http://localhost:5678")
	syncRegistry(tbl)
	if got, want := names(), []string{"svc-a._./aaa.localhost_1234"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
