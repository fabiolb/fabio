package route

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/fabiolb/fabio/metrics"
)

func TestSyncRegistry(t *testing.T) {
	oldRegistry := metrics.M
	metrics.M = newStubRegistry()
	defer func() { metrics.M = oldRegistry }()

	tbl := make(Table)
	tbl.addRoute(&RouteDef{Service: "svc-a", Src: "/aaa", Dst: "http://localhost:1234", Weight: 1})
	tbl.addRoute(&RouteDef{Service: "svc-b", Src: "/bbb", Dst: "http://localhost:5678", Weight: 1})
	if got, want := metrics.M.Names(metrics.ServiceGroup), []string{"svc-a._./aaa.localhost_1234", "svc-b._./bbb.localhost_5678"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	tbl.delRoute(&RouteDef{Service: "svc-b", Src: "/bbb", Dst: "http://localhost:5678"})
	syncRegistry(tbl)
	if got, want := metrics.M.Names(metrics.ServiceGroup), []string{"svc-a._./aaa.localhost_1234"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func newStubRegistry() metrics.Registry {
	return &stubRegistry{names: make(map[string]map[string]bool)}
}

type stubRegistry struct {
	names map[string]map[string]bool
}

func (p *stubRegistry) Names(group string) []string {
	n := []string{}
	for k := range p.names[group] {
		n = append(n, k)
	}
	sort.Strings(n)
	return n
}

func (p *stubRegistry) Unregister(group, name string) {
	delete(p.names[group], name)
}

func (p *stubRegistry) UnregisterAll(group string) {
	p.names[group] = map[string]bool{}
}

func (p *stubRegistry) Gauge(group, name string, n float64) {
	if p.names[group] == nil {
		p.names[group] = map[string]bool{}
	}
	p.names[group][name] = true
}

func (p *stubRegistry) Inc(group, name string, n int64) {
	if p.names[group] == nil {
		p.names[group] = map[string]bool{}
	}
	p.names[group][name] = true
}

func (p *stubRegistry) Time(group, name string, d time.Duration) {
	if p.names[group] == nil {
		p.names[group] = map[string]bool{}
	}
	p.names[group][name] = true
}
