package route

import (
	"reflect"
	"strings"
	"testing"
)

func TestTableRoute(t *testing.T) {
	mustAdd := func(tbl Table, d *RouteDef) {
		if err := tbl.AddRoute(d); err != nil {
			t.Fatalf("got %v want nil for %#v", err, d)
		}
	}

	mustDel := func(tbl Table, d *RouteDef) {
		if err := tbl.DelRoute(d); err != nil {
			t.Fatalf("got %v want nil for %#v", err, d)
		}
	}

	tests := []struct {
		desc  string
		setup func(tbl Table) error
		cfg   []string
		err   string
	}{
		{
			desc: "invalid prefix",
			setup: func(tbl Table) error {
				return tbl.AddRoute(&RouteDef{Service: "svc", Src: "", Dst: "http://bbb.com/"})
			},
			err: errInvalidPrefix.Error(),
		},

		{
			desc: "invalid target",
			setup: func(tbl Table) error {
				return tbl.AddRoute(&RouteDef{Service: "svc", Src: "www.foo.com/", Dst: ""})
			},
			err: errInvalidTarget.Error(),
		},

		{
			desc: "invalid target url",
			setup: func(tbl Table) error {
				return tbl.AddRoute(&RouteDef{Service: "svc", Src: "www.foo.com/", Dst: "://aaa.com/"})
			},
			err: "route: invalid target",
		},

		{
			desc: "new prefix",
			setup: func(tbl Table) error {
				mustAdd(tbl, &RouteDef{Service: "svc-a", Src: "www.foo.com/", Dst: "http://aaa.com/"})
				return nil
			},
			cfg: []string{
				"route add svc-a www.foo.com/ http://aaa.com/",
			},
		},

		{
			desc: "add host to prefix",
			setup: func(tbl Table) error {
				mustAdd(tbl, &RouteDef{Service: "svc-a", Src: "www.foo.com/", Dst: "http://aaa.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-b", Src: "www.foo.com/", Dst: "http://bbb.com/"})
				return nil
			},
			cfg: []string{
				"route add svc-a www.foo.com/ http://aaa.com/",
				"route add svc-b www.foo.com/ http://bbb.com/",
			},
		},

		{
			desc: "add more specific prefix",
			setup: func(tbl Table) error {
				mustAdd(tbl, &RouteDef{Service: "svc-a", Src: "www.foo.com/", Dst: "http://aaa.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-b", Src: "www.foo.com/", Dst: "http://bbb.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-c", Src: "www.foo.com/ccc", Dst: "http://ccc.com/"})
				return nil
			},
			cfg: []string{
				"route add svc-c www.foo.com/ccc http://ccc.com/",
				"route add svc-a www.foo.com/ http://aaa.com/",
				"route add svc-b www.foo.com/ http://bbb.com/",
			},
		},

		{
			desc: "add more specific prefix to existing host",
			setup: func(tbl Table) error {
				mustAdd(tbl, &RouteDef{Service: "svc-a", Src: "www.foo.com/", Dst: "http://aaa.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-b", Src: "www.foo.com/", Dst: "http://bbb.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-b", Src: "www.foo.com/dddddd", Dst: "http://bbb.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-c", Src: "www.foo.com/ccc", Dst: "http://ccc.com/"})
				return nil
			},
			cfg: []string{
				"route add svc-b www.foo.com/dddddd http://bbb.com/",
				"route add svc-c www.foo.com/ccc http://ccc.com/",
				"route add svc-a www.foo.com/ http://aaa.com/",
				"route add svc-b www.foo.com/ http://bbb.com/",
			},
		},

		{
			desc: "add route without host",
			setup: func(tbl Table) error {
				mustAdd(tbl, &RouteDef{Service: "svc-a", Src: "www.foo.com/", Dst: "http://aaa.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-b", Src: "www.foo.com/", Dst: "http://bbb.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-d", Src: "/", Dst: "http://ddd.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-b", Src: "www.foo.com/dddddd", Dst: "http://bbb.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-c", Src: "/ccc", Dst: "http://ccc.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-c", Src: "www.foo.com/ccc", Dst: "http://ccc.com/"})
				return nil
			},
			cfg: []string{
				"route add svc-b www.foo.com/dddddd http://bbb.com/",
				"route add svc-c www.foo.com/ccc http://ccc.com/",
				"route add svc-a www.foo.com/ http://aaa.com/",
				"route add svc-b www.foo.com/ http://bbb.com/",
				"route add svc-c /ccc http://ccc.com/",
				"route add svc-d / http://ddd.com/",
			},
		},

		{
			desc: "delete route",
			setup: func(tbl Table) error {
				mustAdd(tbl, &RouteDef{Service: "svc-a", Src: "www.foo.com/", Dst: "http://aaa.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-b", Src: "www.foo.com/", Dst: "http://bbb.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-b", Src: "www.foo.com/dddddd", Dst: "http://bbb.com/"})
				mustAdd(tbl, &RouteDef{Service: "svc-c", Src: "www.foo.com/ccc", Dst: "http://ccc.com/"})
				mustDel(tbl, &RouteDef{Service: "svc-a", Src: "www.foo.com/", Dst: "http://aaa.com/"})
				return nil
			},
			cfg: []string{
				"route add svc-b www.foo.com/dddddd http://bbb.com/",
				"route add svc-c www.foo.com/ccc http://ccc.com/",
				"route add svc-b www.foo.com/ http://bbb.com/",
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop var
		t.Run(tt.desc, func(t *testing.T) {
			tbl := make(Table)
			err := tt.setup(tbl)
			if got, want := err, tt.err; got == nil && tt.err != "" {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := err, tt.err; got != nil && tt.err == "" {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := err, tt.err; got != nil && !strings.HasPrefix(got.Error(), tt.err) {
				t.Errorf("got %v want %v", got, want)
			}
			if err != nil {
				return
			}
			if got, want := tbl.Config(false), tt.cfg; !reflect.DeepEqual(got, want) {
				t.Errorf("got\n%s\nwant\n%s\n", strings.Join(got, "\n"), strings.Join(want, "\n"))
			}
		})
	}

}
