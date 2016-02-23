package gcp

import (
	"testing"

	"google.golang.org/api/compute/v1"

	"github.com/eBay/fabio/route"
)

func TestBuildInstruction(t *testing.T) {
	for _, each := range []struct {
		spec, inst string
	}{
		{"src=http://in.com/&dst=http://0.0.0.0:8080/here",
			"route add test-node-001 in.com/ http://10.20.30.40:8080/here tags \"tic,tac\""},
		{"src=http://in.com&dst=http://0.0.0.0:8080",
			"route add test-node-001 in.com/ http://10.20.30.40:8080/ tags \"tic,tac\""},
		{"src=http://in.com:80&dst=http://0.0.0.0:80",
			"route add test-node-001 in.com:80/ http://10.20.30.40:80/ tags \"tic,tac\""},
	} {
		i := new(compute.Instance)
		i.Name = "test-node-001"
		i.NetworkInterfaces = []*compute.NetworkInterface{&compute.NetworkInterface{NetworkIP: "10.20.30.40"}}
		i.Tags = &compute.Tags{Items: []string{"tic", "tac"}}
		r := buildInstruction(i, each.spec)
		_, err := route.ParseString(r)
		if err != nil {
			t.Errorf("parse failed:%v", err)
			return
		}
		if got, want := r, each.inst; got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
}

func TestBuildInstructionNoTagsWithWeight(t *testing.T) {
	spec := "src=http://in.com/&dst=http://0.0.0.0:8080/here&weight=0.5"
	i := new(compute.Instance)
	i.Name = "test-node-001"
	i.NetworkInterfaces = []*compute.NetworkInterface{&compute.NetworkInterface{NetworkIP: "10.20.30.40"}}
	route := buildInstruction(i, spec)
	if got, want := route, "route add test-node-001 in.com/ http://10.20.30.40:8080/here weight=0.5"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestBuildInstructionMissingParameter(t *testing.T) {
	for _, each := range []struct {
		spec string
	}{
		{"dst=http://0.0.0.0:8080/here"},
		{"src=http://in.com/"},
		{""},
	} {
		i := new(compute.Instance)
		i.Name = "test-node-001"
		i.NetworkInterfaces = []*compute.NetworkInterface{&compute.NetworkInterface{NetworkIP: "10.20.30.40"}}
		route := buildInstruction(i, each.spec)
		if got, want := route, ""; got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
}
