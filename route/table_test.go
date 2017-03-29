package route

import (
	"crypto/tls"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestTableParse(t *testing.T) {
	genRoutes := func(n int, format string) (a []string) {
		for i := 0; i < n; i++ {
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

		{"weigh fixed weight 0 -> auto distribution",
			[]string{
				`route add svc / http://bar:111/ weight 0`,
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
			tbl, err := NewTable(strings.Join(tt.in, "\n"))
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
		if got, want := normalizeHost(tt.req), tt.host; got != want {
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
	route add svc *.abc.com/ http://foo.com:4000
	route add svc *.abc.com/foo/ http://foo.com:5000
	`

	tbl, err := NewTable(s)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		req *http.Request
		dst string
	}{
		// match on host and path with and without trailing slash
		{&http.Request{Host: "abc.com", URL: mustParse("/")}, "http://foo.com:1000"},
		{&http.Request{Host: "abc.com", URL: mustParse("/bar")}, "http://foo.com:1000"},
		{&http.Request{Host: "abc.com", URL: mustParse("/foo")}, "http://foo.com:1500"},
		{&http.Request{Host: "abc.com", URL: mustParse("/foo/")}, "http://foo.com:2000"},
		{&http.Request{Host: "abc.com", URL: mustParse("/foo/bar")}, "http://foo.com:2500"},
		{&http.Request{Host: "abc.com", URL: mustParse("/foo/bar/")}, "http://foo.com:3000"},

		// do not match on host but maybe on path
		{&http.Request{Host: "def.com", URL: mustParse("/")}, "http://foo.com:800"},
		{&http.Request{Host: "def.com", URL: mustParse("/bar")}, "http://foo.com:800"},
		{&http.Request{Host: "def.com", URL: mustParse("/foo")}, "http://foo.com:900"},

		// strip default port
		{&http.Request{Host: "abc.com:80", URL: mustParse("/")}, "http://foo.com:1000"},
		{&http.Request{Host: "abc.com:443", URL: mustParse("/"), TLS: &tls.ConnectionState{}}, "http://foo.com:1000"},

		// not using default port
		{&http.Request{Host: "abc.com:443", URL: mustParse("/")}, "http://foo.com:800"},
		{&http.Request{Host: "abc.com:80", URL: mustParse("/"), TLS: &tls.ConnectionState{}}, "http://foo.com:800"},

		// glob match the host
		{&http.Request{Host: "x.abc.com", URL: mustParse("/")}, "http://foo.com:4000"},
		{&http.Request{Host: "y.abc.com", URL: mustParse("/abc")}, "http://foo.com:4000"},
		{&http.Request{Host: "x.abc.com", URL: mustParse("/foo/")}, "http://foo.com:5000"},
		{&http.Request{Host: "y.abc.com", URL: mustParse("/foo/")}, "http://foo.com:5000"},
		{&http.Request{Host: ".abc.com", URL: mustParse("/foo/")}, "http://foo.com:5000"},
		{&http.Request{Host: "x.y.abc.com", URL: mustParse("/foo/")}, "http://foo.com:5000"},
		{&http.Request{Host: "y.abc.com:80", URL: mustParse("/foo/")}, "http://foo.com:5000"},
		{&http.Request{Host: "y.abc.com:443", URL: mustParse("/foo/"), TLS: &tls.ConnectionState{}}, "http://foo.com:5000"},
	}

	for i, tt := range tests {
		if got, want := tbl.Lookup(tt.req, "", rndPicker, prefixMatcher).URL.String(), tt.dst; got != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}
