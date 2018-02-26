package route

import (
	"net/url"
	"testing"
)

func TestTarget_GetRedirectURL(t *testing.T) {
	type routeTest struct {
		req  string
		want string
	}
	tests := []struct {
		desc  string
		route string
		tests []routeTest
	}{
		{
			desc:  "simple absolute redirect",
			route: "route add svc / http://bar.com/",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/"},
				{req: "/abc", want: "http://bar.com/"},
				{req: "/a/b/c", want: "http://bar.com/"},
				{req: "/?aaa=1", want: "http://bar.com/"},
			},
		},
		{
			desc:  "absolute redirect to deep path with query",
			route: "route add svc / http://bar.com/a/b/c?foo=bar",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/a/b/c?foo=bar"},
				{req: "/abc", want: "http://bar.com/a/b/c?foo=bar"},
				{req: "/a/b/c", want: "http://bar.com/a/b/c?foo=bar"},
				{req: "/?aaa=1", want: "http://bar.com/a/b/c?foo=bar"},
			},
		},
		{
			desc:  "simple redirect to corresponding path",
			route: "route add svc / http://bar.com/$path",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/"},
				{req: "/abc", want: "http://bar.com/abc"},
				{req: "/a/b/c", want: "http://bar.com/a/b/c"},
				{req: "/?aaa=1", want: "http://bar.com/?aaa=1"},
				{req: "/abc/?aaa=1", want: "http://bar.com/abc/?aaa=1"},
			},
		},
		{
			desc:  "same as above but without / before $path",
			route: "route add svc / http://bar.com$path",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/"},
				{req: "/abc", want: "http://bar.com/abc"},
				{req: "/a/b/c", want: "http://bar.com/a/b/c"},
				{req: "/?aaa=1", want: "http://bar.com/?aaa=1"},
				{req: "/abc/?aaa=1", want: "http://bar.com/abc/?aaa=1"},
			},
		},
		{
			desc:  "arbitrary subdir on target with $path at end",
			route: "route add svc / http://bar.com/bbb/$path",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/bbb/"},
				{req: "/abc", want: "http://bar.com/bbb/abc"},
				{req: "/a/b/c", want: "http://bar.com/bbb/a/b/c"},
				{req: "/?aaa=1", want: "http://bar.com/bbb/?aaa=1"},
				{req: "/abc/?aaa=1", want: "http://bar.com/bbb/abc/?aaa=1"},
			},
		},
		{
			desc:  "same as above but without / before $path",
			route: "route add svc / http://bar.com/bbb$path",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/bbb/"},
				{req: "/abc", want: "http://bar.com/bbb/abc"},
				{req: "/a/b/c", want: "http://bar.com/bbb/a/b/c"},
				{req: "/?aaa=1", want: "http://bar.com/bbb/?aaa=1"},
				{req: "/abc/?aaa=1", want: "http://bar.com/bbb/abc/?aaa=1"},
			},
		},
		{
			desc:  "strip prefix",
			route: "route add svc /stripme http://bar.com/$path opts \"strip=/stripme\"",
			tests: []routeTest{
				{req: "/stripme/", want: "http://bar.com/"},
				{req: "/stripme/abc", want: "http://bar.com/abc"},
				{req: "/stripme/a/b/c", want: "http://bar.com/a/b/c"},
				{req: "/stripme/?aaa=1", want: "http://bar.com/?aaa=1"},
				{req: "/stripme/abc/?aaa=1", want: "http://bar.com/abc/?aaa=1"},
			},
		},
	}
	firstRoute := func(tbl Table) *Route {
		for _, routes := range tbl {
			return routes[0]
		}
		return nil
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			tbl, _ := NewTable(tt.route)
			route := firstRoute(tbl)
			target := route.Targets[0]
			for _, rt := range tt.tests {
				t.Run("", func(t *testing.T) {
					reqURL, _ := url.Parse("http://foo.com" + rt.req)
					got := target.GetRedirectURL(reqURL)
					if got.String() != rt.want {
						t.Errorf("Got %s, wanted %s", got, rt.want)
					}
				})
			}
		})
	}
}
