package route

import (
	"reflect"
	"strings"
	"testing"
)

func TestTableRoute(t *testing.T) {
	mustAdd := func(tbl Table, service, prefix, target string) {
		if err := tbl.AddRoute(service, prefix, target, 0, nil); err != nil {
			t.Fatalf("got %v want nil for %s, %s, %s", err, service, prefix, target)
		}
	}

	mustDel := func(tbl Table, service, prefix, target string) {
		if err := tbl.DelRoute(service, prefix, target); err != nil {
			t.Fatalf("got %v want nil for %s, %s, %s", err, service, prefix, target)
		}
	}

	tests := []struct {
		setup func(tbl Table) error
		cfg   []string
		err   string
	}{
		{ // invalid prefix
			setup: func(tbl Table) error { return tbl.AddRoute("svc", "", "http://bbb.com/", 0, nil) },
			err:   errInvalidPrefix.Error(),
		},

		{ // invalid target
			setup: func(tbl Table) error { return tbl.AddRoute("svc", "www.foo.com/", "", 0, nil) },
			err:   errInvalidTarget.Error(),
		},

		{ // invalid target url
			setup: func(tbl Table) error { return tbl.AddRoute("svc", "www.foo.com/", "://aaa.com/", 0, nil) },
			err:   "route: invalid target",
		},

		{ // new prefix
			setup: func(tbl Table) error {
				mustAdd(tbl, "svc-a", "www.foo.com/", "http://aaa.com/")
				return nil
			},
			cfg: []string{
				"route add svc-a www.foo.com/ http://aaa.com/",
			},
		},

		{ // add host to prefix
			setup: func(tbl Table) error {
				mustAdd(tbl, "svc-a", "www.foo.com/", "http://aaa.com/")
				mustAdd(tbl, "svc-b", "www.foo.com/", "http://bbb.com/")
				return nil
			},
			cfg: []string{
				"route add svc-a www.foo.com/ http://aaa.com/",
				"route add svc-b www.foo.com/ http://bbb.com/",
			},
		},

		{ // add more specific prefix
			setup: func(tbl Table) error {
				mustAdd(tbl, "svc-a", "www.foo.com/", "http://aaa.com/")
				mustAdd(tbl, "svc-b", "www.foo.com/", "http://bbb.com/")
				mustAdd(tbl, "svc-c", "www.foo.com/ccc", "http://ccc.com/")
				return nil
			},
			cfg: []string{
				"route add svc-c www.foo.com/ccc http://ccc.com/",
				"route add svc-a www.foo.com/ http://aaa.com/",
				"route add svc-b www.foo.com/ http://bbb.com/",
			},
		},

		{ // add more specific prefix to existing host
			setup: func(tbl Table) error {
				mustAdd(tbl, "svc-a", "www.foo.com/", "http://aaa.com/")
				mustAdd(tbl, "svc-b", "www.foo.com/", "http://bbb.com/")
				mustAdd(tbl, "svc-b", "www.foo.com/dddddd", "http://bbb.com/")
				mustAdd(tbl, "svc-c", "www.foo.com/ccc", "http://ccc.com/")
				return nil
			},
			cfg: []string{
				"route add svc-b www.foo.com/dddddd http://bbb.com/",
				"route add svc-c www.foo.com/ccc http://ccc.com/",
				"route add svc-a www.foo.com/ http://aaa.com/",
				"route add svc-b www.foo.com/ http://bbb.com/",
			},
		},

		{ // add route without host
			setup: func(tbl Table) error {
				mustAdd(tbl, "svc-a", "www.foo.com/", "http://aaa.com/")
				mustAdd(tbl, "svc-b", "www.foo.com/", "http://bbb.com/")
				mustAdd(tbl, "svc-d", "/", "http://ddd.com/")
				mustAdd(tbl, "svc-b", "www.foo.com/dddddd", "http://bbb.com/")
				mustAdd(tbl, "svc-c", "/ccc", "http://ccc.com/")
				mustAdd(tbl, "svc-c", "www.foo.com/ccc", "http://ccc.com/")
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

		{ // delete route
			setup: func(tbl Table) error {
				mustAdd(tbl, "svc-a", "www.foo.com/", "http://aaa.com/")
				mustAdd(tbl, "svc-b", "www.foo.com/", "http://bbb.com/")
				mustAdd(tbl, "svc-b", "www.foo.com/dddddd", "http://bbb.com/")
				mustAdd(tbl, "svc-c", "www.foo.com/ccc", "http://ccc.com/")
				mustDel(tbl, "svc-a", "www.foo.com/", "http://aaa.com/")
				return nil
			},
			cfg: []string{
				"route add svc-b www.foo.com/dddddd http://bbb.com/",
				"route add svc-c www.foo.com/ccc http://ccc.com/",
				"route add svc-b www.foo.com/ http://bbb.com/",
			},
		},
	}

	for i, tt := range tests {
		tbl := make(Table)
		err := tt.setup(tbl)
		if got, want := err, tt.err; got == nil && tt.err != "" {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
		if got, want := err, tt.err; got != nil && tt.err == "" {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
		if got, want := err, tt.err; got != nil && !strings.HasPrefix(got.Error(), tt.err) {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
		if err != nil {
			continue
		}
		if got, want := tbl.Config(false), tt.cfg; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got\n%s\nwant\n%s\n", i, strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
	}

}
