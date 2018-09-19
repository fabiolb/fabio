package route

import (
	"net/url"
	"testing"
)

func TestTarget_BuildRedirectURL(t *testing.T) {
	type routeTest struct {
		req  string
		want string
	}
	tests := []struct {
		route string
		tests []routeTest
	}{
		{ // simple absolute redirect
			route: "route add svc / http://bar.com/",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/"},
				{req: "/abc", want: "http://bar.com/"},
				{req: "/a/b/c", want: "http://bar.com/"},
				{req: "/?aaa=1", want: "http://bar.com/"},
			},
		},
		{ // absolute redirect to deep path with query
			route: "route add svc / http://bar.com/a/b/c?foo=bar",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/a/b/c?foo=bar"},
				{req: "/abc", want: "http://bar.com/a/b/c?foo=bar"},
				{req: "/a/b/c", want: "http://bar.com/a/b/c?foo=bar"},
				{req: "/?aaa=1", want: "http://bar.com/a/b/c?foo=bar"},
			},
		},
		{ // simple http -> https redirect with static path
			route: "route add redirect *:80/ https://$host/",
			tests: []routeTest{
				{req: "/", want: "https://foo.com/"},
				{req: "/abc", want: "https://foo.com/"},
				{req: "/a/b/c", want: "https://foo.com/"},
				{req: "/?aaa=1", want: "https://foo.com/"},
				{req: "/abc/?aaa=1", want: "https://foo.com/"},
			},
		},
		{ // simple redirect to corresponding path
			route: "route add svc / http://bar.com/$path",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/"},
				{req: "/abc", want: "http://bar.com/abc"},
				{req: "/a/b/c", want: "http://bar.com/a/b/c"},
				{req: "/?aaa=1", want: "http://bar.com/?aaa=1"},
				{req: "/abc/?aaa=1", want: "http://bar.com/abc/?aaa=1"},
			},
		},
		{ // simple http -> https redirect to corresponding host & path
			route: "route add redirect *:80/ https://$host/$path",
			tests: []routeTest{
				{req: "/", want: "https://foo.com/"},
				{req: "/abc", want: "https://foo.com/abc"},
				{req: "/a/b/c", want: "https://foo.com/a/b/c"},
				{req: "/?aaa=1", want: "https://foo.com/?aaa=1"},
				{req: "/abc/?aaa=1", want: "https://foo.com/abc/?aaa=1"},
			},
		},
		{ // simple redirect to corresponding path without / before $path
			route: "route add svc / http://bar.com$path",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/"},
				{req: "/abc", want: "http://bar.com/abc"},
				{req: "/a/b/c", want: "http://bar.com/a/b/c"},
				{req: "/?aaa=1", want: "http://bar.com/?aaa=1"},
				{req: "/abc/?aaa=1", want: "http://bar.com/abc/?aaa=1"},
			},
		},
		{ // simple http -> https redirect to corresponding host & path without / before $path
			route: "route add redirect *:80/ https://$host$path",
			tests: []routeTest{
				{req: "/", want: "https://foo.com/"},
				{req: "/abc", want: "https://foo.com/abc"},
				{req: "/a/b/c", want: "https://foo.com/a/b/c"},
				{req: "/?aaa=1", want: "https://foo.com/?aaa=1"},
				{req: "/abc/?aaa=1", want: "https://foo.com/abc/?aaa=1"},
			},
		},
		{ // arbitrary subdir on target with $path at end
			route: "route add svc / http://bar.com/bbb/$path",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/bbb/"},
				{req: "/abc", want: "http://bar.com/bbb/abc"},
				{req: "/a/b/c", want: "http://bar.com/bbb/a/b/c"},
				{req: "/?aaa=1", want: "http://bar.com/bbb/?aaa=1"},
				{req: "/abc/?aaa=1", want: "http://bar.com/bbb/abc/?aaa=1"},
			},
		},
		{ // http -> https redir to corresonding host w/ arbitrary subdir on target with $path at end
			route: "route add redirect *:80/ https://$host/bbb/$path",
			tests: []routeTest{
				{req: "/", want: "https://foo.com/bbb/"},
				{req: "/abc", want: "https://foo.com/bbb/abc"},
				{req: "/a/b/c", want: "https://foo.com/bbb/a/b/c"},
				{req: "/?aaa=1", want: "https://foo.com/bbb/?aaa=1"},
				{req: "/abc/?aaa=1", want: "https://foo.com/bbb/abc/?aaa=1"},
			},
		},
		{ // arbitrary subdir on target with $path at end but without / before $path
			route: "route add svc / http://bar.com/bbb$path",
			tests: []routeTest{
				{req: "/", want: "http://bar.com/bbb/"},
				{req: "/abc", want: "http://bar.com/bbb/abc"},
				{req: "/a/b/c", want: "http://bar.com/bbb/a/b/c"},
				{req: "/?aaa=1", want: "http://bar.com/bbb/?aaa=1"},
				{req: "/abc/?aaa=1", want: "http://bar.com/bbb/abc/?aaa=1"},
			},
		},
		{ // http -> https redir to corresonding host w/ arbitrary subdir on target with $path at end but without / before $path
			route: "route add redirect *:80/ https://$host/bbb$path",
			tests: []routeTest{
				{req: "/", want: "https://foo.com/bbb/"},
				{req: "/abc", want: "https://foo.com/bbb/abc"},
				{req: "/a/b/c", want: "https://foo.com/bbb/a/b/c"},
				{req: "/?aaa=1", want: "https://foo.com/bbb/?aaa=1"},
				{req: "/abc/?aaa=1", want: "https://foo.com/bbb/abc/?aaa=1"},
			},
		},
		{ // strip prefix
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
		tbl, _ := NewTable(tt.route)
		route := firstRoute(tbl)
		target := route.Targets[0]
		for _, rt := range tt.tests {
			reqURL, _ := url.Parse("http://foo.com" + rt.req)
			target.BuildRedirectURL(reqURL)
			if got := target.RedirectURL.String(); got != rt.want {
				t.Errorf("Got %s, wanted %s", got, rt.want)
			}
		}
	}
}
