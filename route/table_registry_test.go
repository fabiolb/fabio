package route

import (
	"reflect"
	"sort"
	"testing"

	"github.com/eBay/fabio/metrics"
)

func TestSyncRegistry(t *testing.T) {
	oldRegistry := ServiceRegistry
	ServiceRegistry = newStubRegistry()
	defer func() { ServiceRegistry = oldRegistry }()

	tbl := make(Table)
	tbl.AddRoute("svc-a", "/aaa", "http://localhost:1234", 1, nil)
	tbl.AddRoute("svc-b", "/bbb", "http://localhost:5678", 1, nil)
	if got, want := ServiceRegistry.Names(), []string{"svc-a._./aaa.localhost_1234", "svc-b._./bbb.localhost_5678"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	tbl.DelRoute("svc-b", "/bbb", "http://localhost:5678")
	syncRegistry(tbl)
	if got, want := ServiceRegistry.Names(), []string{"svc-a._./aaa.localhost_1234"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func newStubRegistry() metrics.Registry {
	return &stubRegistry{names: make(map[string]bool)}
}

type stubRegistry struct {
	names map[string]bool
}

func (p *stubRegistry) Names() []string {
	n := []string{}
	for k := range p.names {
		n = append(n, k)
	}
	sort.Strings(n)
	return n
}

func (p *stubRegistry) Unregister(name string) {
	delete(p.names, name)
}

func (p *stubRegistry) UnregisterAll() {
	p.names = map[string]bool{}
}

func (p *stubRegistry) GetCounter(name string) metrics.Counter {
	p.names[name] = true
	return metrics.NoopCounter{}
}

func (p *stubRegistry) GetTimer(name string) metrics.Timer {
	p.names[name] = true
	return metrics.NoopTimer{}
}
