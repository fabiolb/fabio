package route

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const (
	// helper constants for the Lookup function
	globEnabled  = false
	globDisabled = true
)

// Global GlobCache for Testing
var globCache = NewGlobCache(1000)

func TestTableParse(t *testing.T) {
	genRoutes := func(n int, format string) (a []string) {
		for i := range n {
			a = append(a, fmt.Sprintf(format, i))
		}
		return a
	}

	tests := []struct {
		desc    string
		in, out []string
	}{

		{"1 service, 1 prefix",
			[]string{
				`route add svc-a / http://aaa.com/`,
			},
			[]string{
				`route add svc-a / http://aaa.com/ weight 1.0000`,
			},
		},

		{"1 service, 1 prefix, 3 instances",
			[]string{
				`route add svc-a / http://aaa.com:1111/`,
				`route add svc-a / http://aaa.com:2222/`,
				`route add svc-a / http://aaa.com:3333/`,
			},
			[]string{
				`route add svc-a / http://aaa.com:1111/ weight 0.3333`,
				`route add svc-a / http://aaa.com:2222/ weight 0.3333`,
				`route add svc-a / http://aaa.com:3333/ weight 0.3333`,
			},
		},

		{"1 service, 1 prefix with option",
			[]string{
				`route add svc-a / http://aaa.com/ opts "strip=/foo"`,
				`route add svc-b / http://bbb.com/ opts "strip=/bar"`,
			},
			[]string{
				`route add svc-a / http://aaa.com/ weight 0.5000 opts "strip=/foo"`,
				`route add svc-b / http://bbb.com/ weight 0.5000 opts "strip=/bar"`,
			},
		},

		{"1 service, 1 prefix, 2 instances with different options",
			[]string{
				`route add svc-a / http://aaa.com/ opts "strip=/foo"`,
				`route add svc-b / http://bbb.com/ opts "strip=/bar"`,
			},
			[]string{
				`route add svc-a / http://aaa.com/ weight 0.5000 opts "strip=/foo"`,
				`route add svc-b / http://bbb.com/ weight 0.5000 opts "strip=/bar"`,
			},
		},

		{"2 service, 1 prefix",
			[]string{
				`route add svc-a / http://aaa.com/`,
				`route add svc-b / http://bbb.com/`,
			},
			[]string{
				`route add svc-a / http://aaa.com/ weight 0.5000`,
				`route add svc-b / http://bbb.com/ weight 0.5000`,
			},
		},

		{"1 service, 2 prefix",
			[]string{
				`route add svc-a /one http://aaa.com/`,
				`route add svc-a /two http://aaa.com/`,
			},
			[]string{
				`route add svc-a /two http://aaa.com/ weight 1.0000`,
				`route add svc-a /one http://aaa.com/ weight 1.0000`,
			},
		},

		{"2 service, 2 prefix",
			[]string{
				`route add svc-a /a http://aaa.com/`,
				`route add svc-b /b http://bbb.com/`,
			},
			[]string{
				`route add svc-b /b http://bbb.com/ weight 1.0000`,
				`route add svc-a /a http://aaa.com/ weight 1.0000`,
			},
		},

		{"sort by more specific prefix",
			[]string{
				`route add svc-a / http://aaa.com/`,
				`route add svc-b /b http://bbb.com/`,
			},
			[]string{
				`route add svc-b /b http://bbb.com/ weight 1.0000`,
				`route add svc-a / http://aaa.com/ weight 1.0000`,
			},
		},

		{"sort prefix with host before prefix without host",
			[]string{
				`route add svc-a / http://aaa.com/`,
				`route add svc-b b.com/ http://bbb.com/`,
			},
			[]string{
				`route add svc-b b.com/ http://bbb.com/ weight 1.0000`,
				`route add svc-a / http://aaa.com/ weight 1.0000`,
			},
		},

		{"add more specific prefix to existing host",
			[]string{
				`route add svc-a a.com/ http://aaa.com/`,
				`route add svc-a a.com/a http://aaa.com/`,
			},
			[]string{
				`route add svc-a a.com/a http://aaa.com/ weight 1.0000`,
				`route add svc-a a.com/ http://aaa.com/ weight 1.0000`,
			},
		},

		{"delete route by service, path and target",
			[]string{
				`route add svc-a / http://aaa.com/`,
				`route add svc-b / http://bbb.com/`,
				`route del svc-b / http://bbb.com/`,
			},
			[]string{
				`route add svc-a / http://aaa.com/ weight 1.0000`,
			},
		},

		{"delete route by service and path",
			[]string{
				`route add svc-a / http://aaa.com/`,
				`route add svc-a / http://aaa.com:2222/`,
				`route add svc-b / http://bbb.com/`,
				`route del svc-a /`,
			},
			[]string{
				`route add svc-b / http://bbb.com/ weight 1.0000`,
			},
		},

		{"delete route by service",
			[]string{
				`route add svc-a /a http://aaa.com/`,
				`route add svc-a / http://aaa.com/`,
				`route add svc-b / http://bbb.com/`,
				`route del svc-a`,
			},
			[]string{
				`route add svc-b / http://bbb.com/ weight 1.0000`,
			},
		},

		{"delete route by service and tags",
			[]string{
				`route add svc-a /a http://aaa.com/ tags "a,b"`,
				`route add svc-a /  http://aaa.com/ tags "b,c"`,
				`route add svc-b /  http://bbb.com/ tags "c,d"`,
				`route del svc-a tags "a,b"`,
			},
			[]string{
				`route add svc-a / http://aaa.com/ weight 0.5000 tags "b,c"`,
				`route add svc-b / http://bbb.com/ weight 0.5000 tags "c,d"`,
			},
		},

		{"delete route by tags",
			[]string{
				`route add svc-a /a http://aaa.com/ tags "a,b"`,
				`route add svc-a /  http://aaa.com/ tags "b,c"`,
				`route add svc-b /  http://bbb.com/ tags "c,d"`,
				`route del tags "b"`,
			},
			[]string{
				`route add svc-b / http://bbb.com/ weight 1.0000 tags "c,d"`,
			},
		},

		{"weigh fixed weight -1 -> auto distribution",
			[]string{
				`route add svc / http://bar:111/ weight -1`,
			},
			[]string{
				`route add svc / http://bar:111/ weight 1.0000`,
			},
		},

		{"weigh only fixed weights and sum(fixedWeight) < 1 -> normalize to 100%",
			[]string{
				`route add svc / http://bar:111/ weight 0.2`,
				`route add svc / http://bar:222/ weight 0.3`,
			},
			[]string{
				`route add svc / http://bar:111/ weight 0.4000`,
				`route add svc / http://bar:222/ weight 0.6000`,
			},
		},

		{"weigh only fixed weights and sum(fixedWeight) > 1 -> normalize to 100%",
			[]string{
				`route add svc / http://bar:111/ weight 2`,
				`route add svc / http://bar:222/ weight 3`,
			},
			[]string{
				`route add svc / http://bar:111/ weight 0.4000`,
				`route add svc / http://bar:222/ weight 0.6000`,
			},
		},

		{"weigh multiple entries for same instance with no fixed weight -> de-duplication",
			[]string{
				`route add svc / http://bar:111/`,
				`route add svc / http://bar:111/`,
			},
			[]string{
				`route add svc / http://bar:111/ weight 1.0000`,
			},
		},

		{"weigh multiple entries with no fixed weight -> even distribution",
			[]string{
				`route add svc / http://bar:111/`,
				`route add svc / http://bar:222/`,
			},
			[]string{
				`route add svc / http://bar:111/ weight 0.5000`,
				`route add svc / http://bar:222/ weight 0.5000`,
			},
		},

		{"weigh multiple entries with de-dup and no fixed weight -> even distribution",
			[]string{
				`route add svc / http://bar:111/`,
				`route add svc / http://bar:111/`,
				`route add svc / http://bar:222/`,
			},
			[]string{
				`route add svc / http://bar:111/ weight 0.5000`,
				`route add svc / http://bar:222/ weight 0.5000`,
			},
		},

		{"weigh mixed fixed and auto weights -> even distribution of remaining weight across non-fixed weighted targets",
			[]string{
				`route add svc / http://bar:111/`,
				`route add svc / http://bar:222/`,
				`route add svc / http://bar:333/ weight 0.5`,
			},
			[]string{
				`route add svc / http://bar:111/ weight 0.2500`,
				`route add svc / http://bar:222/ weight 0.2500`,
				`route add svc / http://bar:333/ weight 0.5000`,
			},
		},

		{"weigh fixed weight == 100% -> route only to fixed weighted targets",
			[]string{
				`route add svc / http://bar:111/`,
				`route add svc / http://bar:222/ weight 0.2500`,
				`route add svc / http://bar:333/ weight 0.7500`,
			},
			[]string{
				`route add svc / http://bar:222/ weight 0.2500`,
				`route add svc / http://bar:333/ weight 0.7500`,
			},
		},

		{"weigh fixed weight > 100%  -> route only to fixed weighted targets and normalize weight",
			[]string{
				`route add svc / http://bar:111/`,
				`route add svc / http://bar:222/ weight 1`,
				`route add svc / http://bar:333/ weight 3`,
			},
			[]string{
				`route add svc / http://bar:222/ weight 0.2500`,
				`route add svc / http://bar:333/ weight 0.7500`,
			},
		},

		{"weigh dynamic weight matched on service name",
			[]string{
				`route add svca / http://bar:111/`,
				`route add svcb / http://bar:222/`,
				`route add svcb / http://bar:333/`,
				`route weight svcb / weight 0.1`,
			},
			[]string{
				`route add svca / http://bar:111/ weight 0.9000`,
				`route add svcb / http://bar:222/ weight 0.0500`,
				`route add svcb / http://bar:333/ weight 0.0500`,
			},
		},

		{"weigh dynamic weight matched on service name and tags",
			[]string{
				`route add svc / http://bar:111/ tags "a"`,
				`route add svc / http://bar:222/ tags "b"`,
				`route add svc / http://bar:333/ tags "b"`,
				`route weight svc / weight 0.1 tags "b"`,
			},
			[]string{
				`route add svc / http://bar:111/ weight 0.9000 tags "a"`,
				`route add svc / http://bar:222/ weight 0.0500 tags "b"`,
				`route add svc / http://bar:333/ weight 0.0500 tags "b"`,
			},
		},

		{"weigh dynamic weight matched on tags",
			[]string{
				`route add svca / http://bar:111/ tags "a"`,
				`route add svcb / http://bar:222/ tags "b"`,
				`route add svcb / http://bar:333/ tags "b"`,
				`route weight / weight 0.1 tags "b"`,
			},
			[]string{
				`route add svca / http://bar:111/ weight 0.9000 tags "a"`,
				`route add svcb / http://bar:222/ weight 0.0500 tags "b"`,
				`route add svcb / http://bar:333/ weight 0.0500 tags "b"`,
			},
		},

		{"weigh more than 1000 routes",
			genRoutes(1234, `route add svc / http://bar:%d/`),
			genRoutes(1234, `route add svc / http://bar:%d/ weight 0.0008`),
		},

		{"weigh more than 1000 routes with a fixed route target",
			func() (a []string) {
				a = genRoutes(1234, `route add svc / http://bar:%d/`)
				a = append(a, `route add svc / http://static:12345/ tags "a"`)
				a = append(a, `route weight svc / weight 0.2 tags "a"`)
				return a
			}(),
			func() (a []string) {
				a = genRoutes(1234, `route add svc / http://bar:%d/ weight 0.0006`)
				a = append(a, `route add svc / http://static:12345/ weight 0.2000 tags "a"`)
				return a
			}(),
		},

		{"weigh route: no target match",
			[]string{
				`route weight svcb / weight 0.1`,
				`route weight svc / weight 0.1 tags "b"`,
				`route weight / weight 0.1 tags "b"`,
			},
			[]string(nil),
		},
	}

	atof := func(s string) float64 {
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic(err)
		}
		return n
	}

	for _, tt := range tests {
		// perform a test which parses the tt.in routes into a table and
		// compares the weighted, generated routing table with tt.out. verify,
		// that the distribution of the target URLs for each prefix in the
		// generated routing table matches the weight This test assumes that
		// the table generates the correct routing table but does not test the
		// actual lookup which it probably should.
		t.Run(tt.desc, func(t *testing.T) {
			// parse the routes
			tbl, err := NewTable(bytes.NewBufferString(strings.Join(tt.in, "\n")))
			if err != nil {
				t.Fatalf("got %v want nil", err)
			}

			// compare the generated routes with the normalized weights
			if got, want := tbl.config(true), tt.out; !reflect.DeepEqual(got, want) {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
			}

			// check that the weights returned in the generated config match
			// the distribution in the wTargets array of the corresponding route.
			checked := map[string]bool{}

			for _, s := range tt.out {
				// route add <svc> <path> ...
				path := strings.Fields(s)[3]

				// if we have already checked this path then skip this.
				// Otherwise, this test becomes O(n^2) and will time out for
				// the large number of routes.
				if checked[path] {
					continue
				}
				checked[path] = true

				// fetch the route
				r := tbl.route(hostpath(path))
				if r == nil {
					t.Fatalf("got nil want route %s", path)
				}

				// check that there are at least some slots
				if len(r.wTargets) == 0 {
					t.Fatalf("got 0 targets want some")
				}

				// pre-generate the target urls for comparison as this
				// will otherwise slow the test down significantly
				targetURLs := make([]string, len(r.wTargets))
				for i, tg := range r.wTargets {
					targetURLs[i] = tg.URL.Scheme + "://" + tg.URL.Host + tg.URL.Path
				}

				// count how often the 'url' from 'route add svc <path> <url>'
				// appears in the list of wTargets for all the URLs
				// from the routes to determine whether the actual
				// distribution of each target within the wTarget slice
				// matches what we expect
				for _, s := range tt.out {
					// route add <svc> <path> <url> weight <weight> ...`,
					p := strings.Fields(s)

					// skip if the path doesn't match
					if path != p[3] {
						continue
					}

					// count how often the target url appears in the list of wTargets
					count := 0
					for _, u := range targetURLs {
						if u == p[4] {
							count++
						}
					}

					// calc the weight as nSlots/totalSlots
					gotWeight := float64(count) / float64(len(r.wTargets))

					// round the weight down to the number of decimal points
					// supported by maxSlots
					gotWeight = float64(int(gotWeight*float64(maxSlots))) / float64(maxSlots)

					// compare to the weight from the generated config
					wantWeight := atof(p[6])

					// check that the actual weight is within 2% of the computed weight
					if math.Abs(gotWeight-wantWeight) > 0.02 {
						t.Errorf("got weight %f want %f", gotWeight, wantWeight)
					}

					// TODO(fs): verify distriibution of targets across the ring
					// TODO(fs): verify lookup with 'rr' works as expected. Current test is by proxy of generated config.
				}
			}
		})
	}
}

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		req  *http.Request
		host string
	}{
		{&http.Request{Host: "foo.com"}, "foo.com"},
		{&http.Request{Host: "foo.com:80"}, "foo.com"},
		{&http.Request{Host: "foo.com:81"}, "foo.com:81"},
		{&http.Request{Host: "foo.com", TLS: &tls.ConnectionState{}}, "foo.com"},
		{&http.Request{Host: "foo.com:443", TLS: &tls.ConnectionState{}}, "foo.com"},
		{&http.Request{Host: "foo.com:444", TLS: &tls.ConnectionState{}}, "foo.com:444"},
	}

	for i, tt := range tests {
		if got, want := normalizeHost(tt.req.Host, tt.req.TLS != nil), tt.host; got != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}

// see https://github.com/fabiolb/fabio/issues/448
// for more information on the issue and purpose of this test
func TestTableLookupIssue448(t *testing.T) {
	s := `
	route add mock foo.com:80/ https://foo.com/ opts "redirect=301"
	route add mock aaa.com:80/ http://bbb.com/ opts "redirect=301"
	route add mock ccc.com:443/bar https://ccc.com/baz opts "redirect=301"
	route add mock / http://foo.com/
	`

	tbl, err := NewTable(bytes.NewBufferString(s))
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		req         *http.Request
		dst         string
		globEnabled bool
	}{
		{
			req: &http.Request{
				Host: "foo.com",
				URL:  mustParse("/"),
			},
			dst: "https://foo.com/",
			// empty upstream header should follow redirect - standard behavior
		},
		{
			req: &http.Request{
				Host:   "foo.com",
				URL:    mustParse("/"),
				Header: http.Header{"X-Forwarded-Proto": {"http"}},
			},
			dst: "https://foo.com/",
			// upstream http request to same host and path should follow redirect
		},
		{
			req: &http.Request{
				Host:   "foo.com",
				URL:    mustParse("/"),
				Header: http.Header{"X-Forwarded-Proto": {"https"}},
				TLS:    &tls.ConnectionState{},
			},
			dst: "http://foo.com/",
			// upstream https request to same host and path should NOT follow redirect"
		},
		{
			req: &http.Request{
				Host:   "aaa.com",
				URL:    mustParse("/"),
				Header: http.Header{"X-Forwarded-Proto": {"http"}},
			},
			dst: "http://bbb.com/",
			// upstream http request to different http host should follow redirect
		},
		{
			req: &http.Request{
				Host:   "ccc.com",
				URL:    mustParse("/bar"),
				Header: http.Header{"X-Forwarded-Proto": {"https"}},
				TLS:    &tls.ConnectionState{},
			},
			dst: "https://ccc.com/baz",
			// upstream https request to same https host with different path should follow redirect"
		},
	}

	for i, tt := range tests {
		if got, want := tbl.Lookup(tt.req, rndPicker, prefixMatcher, globCache, globEnabled).URL.String(), tt.dst; got != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}

func TestTableLookup(t *testing.T) {
	s := `
	route add svc / http://foo.com:800
	route add svc /foo http://foo.com:900
	route add svc abc.com/ http://foo.com:1000
	route add svc abc.com/foo http://foo.com:1500
	route add svc abc.com/foo/ http://foo.com:2000
	route add svc abc.com/foo/bar http://foo.com:2500
	route add svc abc.com/foo/bar/ http://foo.com:3000
	route add svc z.abc.com/foo/ http://foo.com:3100
	route add svc *.abc.com/ http://foo.com:4000
	route add svc *.abc.com/foo/ http://foo.com:5000
	route add svc *.aaa.abc.com/ http://foo.com:6000
	route add svc *.bbb.abc.com/ http://foo.com:6100
	route add svc xyz.com:80/ https://xyz.com
	`

	tbl, err := NewTable(bytes.NewBufferString(s))
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		req         *http.Request
		dst         string
		globEnabled bool
	}{
		// match on host and path with and without trailing slash
		{&http.Request{Host: "abc.com", URL: mustParse("/")}, "http://foo.com:1000", globEnabled},
		{&http.Request{Host: "abc.com", URL: mustParse("/bar")}, "http://foo.com:1000", globEnabled},
		{&http.Request{Host: "abc.com", URL: mustParse("/foo")}, "http://foo.com:1500", globEnabled},
		{&http.Request{Host: "abc.com", URL: mustParse("/foo/")}, "http://foo.com:2000", globEnabled},
		{&http.Request{Host: "abc.com", URL: mustParse("/foo/bar")}, "http://foo.com:2500", globEnabled},
		{&http.Request{Host: "abc.com", URL: mustParse("/foo/bar/")}, "http://foo.com:3000", globEnabled},

		// do not match on host but maybe on path
		{&http.Request{Host: "def.com", URL: mustParse("/")}, "http://foo.com:800", globEnabled},
		{&http.Request{Host: "def.com", URL: mustParse("/bar")}, "http://foo.com:800", globEnabled},
		{&http.Request{Host: "def.com", URL: mustParse("/foo")}, "http://foo.com:900", globEnabled},

		// strip default port
		{&http.Request{Host: "abc.com:80", URL: mustParse("/")}, "http://foo.com:1000", globEnabled},
		{&http.Request{Host: "abc.com:443", URL: mustParse("/"), TLS: &tls.ConnectionState{}}, "http://foo.com:1000", globEnabled},

		// not using default port
		{&http.Request{Host: "abc.com:443", URL: mustParse("/")}, "http://foo.com:800", globEnabled},
		{&http.Request{Host: "abc.com:80", URL: mustParse("/"), TLS: &tls.ConnectionState{}}, "http://foo.com:800", globEnabled},

		// glob match the host
		{&http.Request{Host: "x.abc.com", URL: mustParse("/")}, "http://foo.com:4000", globEnabled},
		{&http.Request{Host: "y.abc.com", URL: mustParse("/abc")}, "http://foo.com:4000", globEnabled},
		{&http.Request{Host: "x.abc.com", URL: mustParse("/foo/")}, "http://foo.com:5000", globEnabled},
		{&http.Request{Host: "y.abc.com", URL: mustParse("/foo/")}, "http://foo.com:5000", globEnabled},
		{&http.Request{Host: ".abc.com", URL: mustParse("/foo/")}, "http://foo.com:5000", globEnabled},
		{&http.Request{Host: "x.y.abc.com", URL: mustParse("/foo/")}, "http://foo.com:5000", globEnabled},
		{&http.Request{Host: "y.abc.com:80", URL: mustParse("/foo/")}, "http://foo.com:5000", globEnabled},
		{&http.Request{Host: "x.aaa.abc.com", URL: mustParse("/")}, "http://foo.com:6000", globEnabled},
		{&http.Request{Host: "x.aaa.abc.com", URL: mustParse("/foo")}, "http://foo.com:6000", globEnabled},
		{&http.Request{Host: "x.bbb.abc.com", URL: mustParse("/")}, "http://foo.com:6100", globEnabled},
		{&http.Request{Host: "x.bbb.abc.com", URL: mustParse("/foo")}, "http://foo.com:6100", globEnabled},
		{&http.Request{Host: "y.abc.com:443", URL: mustParse("/foo/"), TLS: &tls.ConnectionState{}}, "http://foo.com:5000", globEnabled},

		// exact match has precedence over glob match
		{&http.Request{Host: "z.abc.com", URL: mustParse("/foo/")}, "http://foo.com:3100", globEnabled},

		// explicit port on route
		{&http.Request{Host: "xyz.com", URL: mustParse("/")}, "https://xyz.com", globEnabled},
	}

	for i, tt := range tests {
		if got, want := tbl.Lookup(tt.req, rndPicker, prefixMatcher, globCache, globEnabled).URL.String(), tt.dst; got != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}

func TestTableLookup_656(t *testing.T) {
	// A typical HTTPS redirect
	s := `
	route add my-service example.com:80/ https://example.com$path opts "redirect=301"
	route add my-service example.com/ http://127.0.0.1:3000/
	`

	tbl, err := NewTable(bytes.NewBufferString(s))
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	target := tbl.Lookup(req, rrPicker, prefixMatcher, globCache, globDisabled)

	if target == nil {
		t.Fatal("No route match")
	}
	if got, want := target.RedirectCode, 301; got != want {
		t.Errorf("target.RedirectCode = %d, want %d", got, want)
	}
	if got, want := fmt.Sprint(target.RedirectURL), "https://example.com/foo"; got != want {
		t.Errorf("target.RedirectURL = %s, want %s", got, want)
	}
}

func TestTableLookupMethod(t *testing.T) {
	tbl, err := NewTable(bytes.NewBufferString(`
route add svc / http://global/
route add svc /api http://api-default/
route add svc /api http://post-put/ opts "method=POST,PUT"
route add svc /api http://post/ opts "method=POST"
route add svc /api http://get/ opts "method=GET"
route add svc /api/v2 http://post-v2/ opts "method=POST"
route add svc example.com/api http://host-get/ opts "method=GET"
`))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name, method, host, path, want string
		match                          matcher
		globDisabled                   bool
	}{
		{"GET", "GET", "other.com", "/api", "http://get/", prefixMatcher, false},
		{"POST narrower method set", "POST", "other.com", "/api", "http://post/", prefixMatcher, false},
		{"PUT multiple methods", "PUT", "other.com", "/api", "http://post-put/", prefixMatcher, false},
		{"DELETE falls back to same path", "DELETE", "other.com", "/api", "http://api-default/", prefixMatcher, false},
		{"method case is significant", "get", "other.com", "/api", "http://api-default/", prefixMatcher, false},
		{"specific path POST", "POST", "other.com", "/api/v2/thing", "http://post-v2/", prefixMatcher, false},
		{"specific path GET falls back to parent", "GET", "other.com", "/api/v2/thing", "http://get/", prefixMatcher, false},
		{"specific path PATCH falls back to parent", "PATCH", "other.com", "/api/v2/thing", "http://api-default/", prefixMatcher, false},
		{"different path", "PATCH", "other.com", "/other", "http://global/", prefixMatcher, false},
		{"host-specific GET", "GET", "example.com", "/api", "http://host-get/", prefixMatcher, false},
		{"host fallback to hostless POST", "POST", "example.com", "/api", "http://post/", prefixMatcher, false},
		{"empty method uses unconditional route", "", "other.com", "/api", "http://api-default/", prefixMatcher, false},
		{"iprefix", "GET", "other.com", "/API/v2", "http://get/", iPrefixMatcher, false},
		{"host without glob", "GET", "example.com", "/api", "http://host-get/", prefixMatcher, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{Method: tt.method, Host: tt.host, URL: mustParse(tt.path)}
			target := tbl.Lookup(req, rrPicker, tt.match, globCache, tt.globDisabled)
			if target == nil {
				t.Fatalf("no route for %s %s%s; want %s", tt.method, tt.host, tt.path, tt.want)
			}
			if got := target.URL.String(); got != tt.want {
				t.Errorf("got %s; want %s", got, tt.want)
			}
		})
	}

	methodsOnly, err := NewTable(bytes.NewBufferString(`
route add svc /api http://get-only/ opts "method=GET"
route add svc example.org/ http://post-only/ opts "method=POST"
`))
	if err != nil {
		t.Fatal(err)
	}
	for _, method := range []string{"POST", ""} {
		req := &http.Request{Method: method, Host: "other.com", URL: mustParse("/api")}
		if got := methodsOnly.Lookup(req, rrPicker, prefixMatcher, globCache, globEnabled); got != nil {
			t.Errorf("unmatched method %q selected %s; want nil", method, got.URL)
		}
	}
	if got := methodsOnly.LookupHost("example.org", rrPicker); got != nil {
		t.Errorf("TCP lookup selected method-limited target %s; want nil", got.URL)
	}

	other, err := NewTable(bytes.NewBufferString(`
route add svc /user/* http://glob-get/ opts "method=GET"
route add svc /* http://glob-fallback/
route add svc grpc.com/ http://grpc-default/
route add svc grpc.com/ http://grpc-get/ opts "method=GET"
`))
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ method, want string }{
		{"GET", "http://glob-get/"},
		{"POST", "http://glob-fallback/"},
	} {
		req := &http.Request{Method: tt.method, Host: "other.com", URL: mustParse("/user/123")}
		if target := other.Lookup(req, rrPicker, globMatcher, globCache, globEnabled); target == nil || target.URL.String() != tt.want {
			t.Errorf("glob %s: got %v; want %s", tt.method, target, tt.want)
		}
	}
	if target := other.LookupHost("grpc.com", rrPicker); target == nil || target.URL.String() != "http://grpc-default/" {
		t.Errorf("methodless TCP lookup: got %v; want default", target)
	}
	req := &http.Request{Host: "grpc.com", URL: mustParse("/grpc.Service/Method")}
	if target := other.Lookup(req, rrPicker, prefixMatcher, globCache, globEnabled); target == nil || target.URL.String() != "http://grpc-default/" {
		t.Errorf("methodless gRPC lookup: got %v; want default", target)
	}
}

func TestTableMethodPools(t *testing.T) {
	defs := `
route add svc /api http://first/ opts "method=GET,POST"
route add svc /api http://second/ opts "method=POST,GET"
route add svc /api http://first/ opts "method=POST,GET"
route add svc /api http://post/ opts "method=POST"
route add svc /api http://put/ opts "method=PUT"
route add svc /api http://default/
`
	tbl, err := NewTable(bytes.NewBufferString(defs))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(tbl[""]); got != 4 {
		t.Fatalf("got %d method pools, want 4", got)
	}
	methods := tbl[""].findMethods("/api", []string{"GET", "POST"})
	if methods == nil || len(methods.Targets) != 2 {
		t.Fatalf("equivalent method lists should share a two-target pool; got %v", methods)
	}
	for _, target := range methods.Targets {
		if target.Weight != 0.5 {
			t.Errorf("shared method pool weight = %f, want 0.5", target.Weight)
		}
	}
	for i, want := range []string{"http://first/", "http://second/", "http://first/", "http://second/"} {
		req := &http.Request{Method: "GET", URL: mustParse("/api")}
		if got := tbl.Lookup(req, rrPicker, prefixMatcher, globCache, globEnabled).URL.String(); got != want {
			t.Errorf("round robin %d: got %s want %s", i, got, want)
		}
	}
	for range 30 {
		req := &http.Request{Method: "GET", URL: mustParse("/api")}
		got := tbl.Lookup(req, rndPicker, prefixMatcher, globCache, globEnabled).URL.String()
		if got != "http://first/" && got != "http://second/" {
			t.Errorf("random picker selected unrelated method pool: %s", got)
		}
	}
	for _, tt := range []struct{ method, want string }{
		{"POST", "http://post/"},
		{"PUT", "http://put/"},
		{"DELETE", "http://default/"},
	} {
		req := &http.Request{Method: tt.method, URL: mustParse("/api")}
		if got := tbl.Lookup(req, rrPicker, prefixMatcher, globCache, globEnabled).URL.String(); got != tt.want {
			t.Errorf("%s: got %s want %s", tt.method, got, tt.want)
		}
	}
	serialized := tbl.String()
	if !strings.Contains(serialized, "method=POST,GET") {
		t.Fatalf("method options missing from serialized table:\n%s", serialized)
	}
	reparsed, err := NewTable(bytes.NewBufferString(serialized))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(reparsed[""].findMethods("/api", []string{"GET", "POST"}).Targets); got != 2 {
		t.Errorf("round-trip grouped %d targets; want 2", got)
	}
	for _, tt := range []struct{ method, want string }{
		{"GET", "http://first/"},
		{"POST", "http://post/"},
		{"PUT", "http://put/"},
		{"DELETE", "http://default/"},
	} {
		req := &http.Request{Method: tt.method, URL: mustParse("/api")}
		if target := reparsed.Lookup(req, rrPicker, prefixMatcher, globCache, globEnabled); target == nil || target.URL.String() != tt.want {
			t.Errorf("reparsed %s: got %v want %s", tt.method, target, tt.want)
		}
	}

	// Equal-sized overlapping method sets retain their input order.
	equal, err := NewTable(bytes.NewBufferString(`
route add svc /api http://first/ opts "method=GET,POST"
route add svc /api http://second/ opts "method=GET,PUT"
`))
	if err != nil {
		t.Fatal(err)
	}
	req := &http.Request{Method: "GET", URL: mustParse("/api")}
	if got := equal.Lookup(req, rrPicker, prefixMatcher, globCache, globEnabled).URL.String(); got != "http://first/" {
		t.Errorf("equal-sized overlapping conditions: got %s; want first configured target", got)
	}
}

func TestTableMethodMutation(t *testing.T) {
	base := `
route add svc /api http://get-x/ tags "x" opts "method=GET"
route add svc /api http://get-y/ tags "y" opts "method=GET"
route add svc /api http://post-x/ tags "x" opts "method=POST"
route add svc /api http://post-y/ tags "y" opts "method=POST"
route add svc /api http://all-x/ tags "x"
route add svc /api http://all-y/ tags "y"
`

	t.Run("weight applies independently to all method pools", func(t *testing.T) {
		tbl, err := NewTable(bytes.NewBufferString(base + `route weight svc /api weight 0.2 tags "x"`))
		if err != nil {
			t.Fatal(err)
		}
		if len(tbl[""]) != 3 {
			t.Fatalf("got %d pools, want 3", len(tbl[""]))
		}
		for _, r := range tbl[""] {
			if len(r.Targets) != 2 {
				t.Fatalf("pool %v has %d targets, want 2", r.Methods, len(r.Targets))
			}
			for _, target := range r.Targets {
				if contains(target.Tags, []string{"x"}) {
					if target.FixedWeight != 0.2 || target.Weight != 0.2 {
						t.Errorf("pool %v tagged x has weights %v/%v, want 0.2", r.Methods, target.FixedWeight, target.Weight)
					}
				} else if target.FixedWeight != -1 || math.Abs(target.Weight-0.8) > 1e-9 {
					t.Errorf("pool %v tagged y has weights %v/%v, want -1/0.8", r.Methods, target.FixedWeight, target.Weight)
				}
			}
		}
		if err := tbl.weighRoute(&RouteDef{Src: "/api", Service: "absent", Weight: 0.2}); err != errNoMatch {
			t.Errorf("unmatched weight command = %v, want %v", err, errNoMatch)
		}
	})

	t.Run("tag delete touches every method pool", func(t *testing.T) {
		tbl, err := NewTable(bytes.NewBufferString(base + `route del svc tags "x"`))
		if err != nil {
			t.Fatal(err)
		}
		if len(tbl[""]) != 3 {
			t.Fatalf("got %d pools after tag delete, want 3", len(tbl[""]))
		}
		for _, r := range tbl[""] {
			if len(r.Targets) != 1 || !contains(r.Targets[0].Tags, []string{"y"}) {
				t.Errorf("tag delete missed pool %v: %v", r.Methods, r.Targets)
			}
		}
	})

	t.Run("target deletion removes the same destination from every pool", func(t *testing.T) {
		tbl, err := NewTable(bytes.NewBufferString(`
route add svc /api http://shared/ opts "method=GET"
route add svc /api http://shared/ opts "method=POST"
route add svc /api http://shared/
route add other /api http://keep/
route del svc /api http://shared/
`))
		if err != nil {
			t.Fatal(err)
		}
		if len(tbl[""]) != 1 || len(tbl[""][0].Targets) != 1 {
			t.Fatalf("expected only the surviving default pool, got %v", tbl)
		}
		for _, method := range []string{"GET", "POST", "DELETE"} {
			req := &http.Request{Method: method, URL: mustParse("/api")}
			if got := tbl.Lookup(req, rrPicker, prefixMatcher, globCache, globEnabled).URL.String(); got != "http://keep/" {
				t.Errorf("%s: got %s want http://keep/", method, got)
			}
		}
	})

	for _, tt := range []struct {
		name, command string
		want          []string
	}{
		{"delete service and path", "route del svc /api", nil},
		{"delete service globally", "route del svc", nil},
		{"delete tagged targets", `route del tags "x"`, []string{"http://get-y/", "http://post-y/", "http://all-y/"}},
		{"delete matching destination", "route del svc /api http://post-x/", []string{"http://get-x/", "http://get-y/", "http://post-y/", "http://all-x/", "http://all-y/"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tbl, err := NewTable(bytes.NewBufferString(base + tt.command))
			if err != nil {
				t.Fatal(err)
			}
			got := tbl.String()
			for _, dst := range []string{"http://get-x/", "http://get-y/", "http://post-x/", "http://post-y/", "http://all-x/", "http://all-y/"} {
				want := false
				for _, expected := range tt.want {
					if dst == expected {
						want = true
					}
				}
				if strings.Contains(got, dst) != want {
					t.Errorf("destination %s presence %v, want %v; table:\n%s", dst, strings.Contains(got, dst), want, got)
				}
			}
			if tt.want == nil && len(tbl) != 0 {
				t.Errorf("empty method pools remained after delete: %v", tbl)
			}
		})
	}

	t.Run("custom definitions and method-based deletion", func(t *testing.T) {
		defs := []RouteDef{
			{Cmd: RouteAddCmd, Service: "svc", Src: "/api", Dst: "http://get/", Weight: -1, Opts: map[string]string{"method": "GET"}},
			{Cmd: RouteAddCmd, Service: "svc", Src: "/api", Dst: "http://post/", Weight: -1, Opts: map[string]string{"method": "POST"}},
			{Cmd: RouteAddCmd, Service: "other", Src: "/api", Dst: "http://default/", Weight: -1},
			{Cmd: RouteDelCmd, Service: "svc", Src: "/api", Dst: "http://get/"},
		}
		tbl, err := NewTableCustom(&defs)
		if err != nil {
			t.Fatal(err)
		}
		if len(tbl[""]) != 2 {
			t.Errorf("got %d pools, want POST and default only", len(tbl[""]))
		}
		for _, tt := range []struct{ method, want string }{
			{"GET", "http://default/"},
			{"POST", "http://post/"},
		} {
			req := &http.Request{Method: tt.method, URL: mustParse("/api")}
			if got := tbl.Lookup(req, rrPicker, prefixMatcher, globCache, globEnabled).URL.String(); got != tt.want {
				t.Errorf("%s: got %s want %s", tt.method, got, tt.want)
			}
		}
	})
}

func TestNewTableCustom(t *testing.T) {

	var routes []RouteDef
	var tags = []string{"tag1", "tag2"}
	var opts = make(map[string]string)
	opts["tlsskipverify"] = "true"
	opts["proto"] = "http"

	var route1 = RouteDef{
		Cmd:     "route add",
		Service: "service1",
		Src:     "app.com",
		Dst:     "http://10.1.1.1:8080",
		Weight:  0.50,
		Tags:    tags,
		Opts:    opts,
	}
	var route2 = RouteDef{
		Cmd:     "route add",
		Service: "service1",
		Src:     "app.com",
		Dst:     "http://10.1.1.2:8080",
		Weight:  0.50,
		Tags:    tags,
		Opts:    opts,
	}
	var route3 = RouteDef{
		Cmd:     "route add",
		Service: "service2",
		Src:     "app.com",
		Dst:     "http://10.1.1.3:8080",
		Weight:  0.25,
		Tags:    tags,
		Opts:    opts,
	}
	var route4 = RouteDef{
		Cmd:     "route weight",
		Service: "service1",
		Src:     "app.com",
		Weight:  0.10,
		Tags:    tags,
		Opts:    opts,
	}

	var route5 = RouteDef{
		Cmd:     "route weight",
		Service: "service3",
		Src:     "example.com",
		Weight:  1,
	}

	routes = append(routes, route1)
	routes = append(routes, route2)
	routes = append(routes, route3)
	routes = append(routes, route4)
	routes = append(routes, route5)

	table, err := NewTableCustom(&routes)

	if err != nil {
		fmt.Printf("Got error from NewTableCustom - %s", err.Error())
		t.FailNow()
	}

	tableString := table.String()
	if !strings.Contains(tableString, route1.Dst) {
		fmt.Printf("Table Missing Destination %s -- Table -- %s", route1.Dst, tableString)
		t.FailNow()
	}

	if !strings.Contains(tableString, route2.Dst) {
		fmt.Printf("Table Missing Destination %s -- Table -- %s", route2.Dst, tableString)
		t.FailNow()
	}

	if !strings.Contains(tableString, route3.Dst) {
		fmt.Printf("Table Missing Destination %s -- Table -- %s", route3.Dst, tableString)
		t.FailNow()
	}

	want := `route add service1 app.com/ http://10.1.1.1:8080 weight 0.0500 tags "tag1,tag2" opts "proto=http tlsskipverify=true"
route add service1 app.com/ http://10.1.1.2:8080 weight 0.0500 tags "tag1,tag2" opts "proto=http tlsskipverify=true"
route add service2 app.com/ http://10.1.1.3:8080 weight 0.2500 tags "tag1,tag2" opts "proto=http tlsskipverify=true"
`

	if strings.TrimSpace(tableString) != strings.TrimSpace(want) {
		t.Fatalf("mismatch\nwant:\n%s\ngot:\n%s", want, tableString)
	}
}

func TestTable_Dump(t *testing.T) {
	s := `
	route add svc / http://foo.com:800
	route add svc /foo http://foo.com:900
	route add svc abc.com/ http://foo.com:1000
	`

	tbl, err := NewTable(bytes.NewBufferString(s))
	if err != nil {
		t.Fatal(err)
	}

	want := `+-- host=
|   |-- path=/foo
|   |    +-- addr=foo.com:900 weight 1.00 slots 1/1
|   +-- path=/
|       +-- addr=foo.com:800 weight 1.00 slots 1/1
+-- host=abc.com
    +-- path=/
        +-- addr=foo.com:1000 weight 1.00 slots 1/1
`

	got := tbl.Dump()

	if want != got {
		t.Errorf("Unexpected Dump() output:\nwant:\n%s\ngot:\n%s\n", want, got)
	}
}
