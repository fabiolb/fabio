package consul

import (
	"reflect"
	"testing"

	"github.com/hashicorp/consul/api"
)

func TestRouteCmd(t *testing.T) {
	cases := []struct {
		name string
		r    routecmd
		cfg  []string
	}{
		{
			name: "http",
			r: routecmd{
				prefix: "p-",
				svc: &api.CatalogService{
					ServiceName:    "svc-1",
					ServiceAddress: "1.1.1.1",
					ServicePort:    2222,
					ServiceTags:    []string{`p-foo/bar`},
				},
			},
			cfg: []string{
				`route add svc-1 foo/bar http://1.1.1.1:2222/`,
			},
		},
		{
			name: "http single tag with a space",
			r: routecmd{
				prefix: "p-",
				svc: &api.CatalogService{
					ServiceName:    "svc-1",
					ServiceAddress: "1.1.1.1",
					ServicePort:    2222,
					ServiceTags:    []string{` p-foo/bar`},
				},
			},
			cfg: []string{
				`route add svc-1 foo/bar http://1.1.1.1:2222/`,
			},
		},
		{
			name: "http multiple tags",
			r: routecmd{
				prefix: "p-",
				svc: &api.CatalogService{
					ServiceName:    "svc-1",
					ServiceAddress: "1.1.1.1",
					ServicePort:    2222,
					ServiceTags:    []string{`p-foo/bar`, `p-test/foo`},
				},
			},
			cfg: []string{
				`route add svc-1 foo/bar http://1.1.1.1:2222/`,
				`route add svc-1 test/foo http://1.1.1.1:2222/`,
			},
		},
		{
			name: "http multiple routes with space prefix route",
			r: routecmd{
				prefix: "p-",
				svc: &api.CatalogService{
					ServiceName:    "svc-1",
					ServiceAddress: "1.1.1.1",
					ServicePort:    2222,
					ServiceTags:    []string{`p-foo/bar`, `  p-test/foo`},
				},
			},
			cfg: []string{
				`route add svc-1 foo/bar http://1.1.1.1:2222/`,
				`route add svc-1 test/foo http://1.1.1.1:2222/`,
			},
		},
		{
			name: "tcp",
			r: routecmd{
				prefix: "p-",
				svc: &api.CatalogService{
					ServiceName:    "svc-1",
					ServiceAddress: "1.1.1.1",
					ServicePort:    2222,
					ServiceTags:    []string{`p-:1234 proto=tcp`},
				},
			},
			cfg: []string{
				`route add svc-1 :1234 tcp://1.1.1.1:2222`,
			},
		},
		{
			name: "metadata only with node address fallback",
			r: routecmd{
				prefix: "urlprefix-",
				svc: &api.CatalogService{
					ServiceName: "users",
					Address:     "10.0.0.3",
					ServicePort: 8080,
					ServiceTags: []string{"fabio", "prod"},
					ServiceMeta: map[string]string{"urlprefix-0": "/users"},
				},
			},
			cfg: []string{
				`route add users /users http://10.0.0.3:8080/ tags "fabio,prod"`,
			},
		},
		{
			name: "metadata keys sorted with http https and tcp options",
			r: routecmd{
				prefix: "urlprefix-",
				svc: &api.CatalogService{
					ServiceName:    "svc-2",
					ServiceAddress: "1.1.1.1",
					ServicePort:    1234,
					ServiceMeta: map[string]string{
						"urlprefix-2": ":3306 proto=tcp weight=20",
						"urlprefix-0": "  /catalog  ",
						"urlprefix-1": "EXAMPLE.com/shop proto=https tlsskipverify=true weight=4",
						"urlprefix-3": " \t ",
						"region":      "west",
						"other-route": "/ignored",
					},
				},
			},
			cfg: []string{
				`route add svc-2 /catalog http://1.1.1.1:1234/`,
				`route add svc-2 example.com/shop https://1.1.1.1:1234 weight 4 opts "tlsskipverify=true"`,
				`route add svc-2 :3306 tcp://1.1.1.1:1234 weight 20`,
			},
		},
		{
			name: "mixed tags and metadata with custom prefix",
			r: routecmd{
				prefix: "route-",
				svc: &api.CatalogService{
					ServiceName:    "mixed",
					ServiceAddress: "2.2.2.2",
					ServicePort:    8080,
					ServiceTags: []string{
						"route-/legacy weight=2", "blue", "fabio", "route-/old proto=https",
					},
					ServiceMeta: map[string]string{
						"route-9":     "HOST.EXAMPLE/redirect redirect=302,https://target.example/",
						"route-0":     "/v2 strip=/api prepend=/new",
						"urlprefix-0": "/ignored",
					},
				},
			},
			cfg: []string{
				`route add mixed /legacy http://2.2.2.2:8080/ weight 2 tags "blue,fabio"`,
				`route add mixed /old https://2.2.2.2:8080 tags "blue,fabio"`,
				`route add mixed /v2 http://2.2.2.2:8080/ tags "blue,fabio" opts "strip=/api prepend=/new"`,
				`route add mixed host.example/redirect https://target.example/ tags "blue,fabio" opts "redirect=302"`,
			},
		},
		{
			name: "unrelated and empty metadata without route tags",
			r: routecmd{
				prefix: "urlprefix-",
				svc: &api.CatalogService{
					ServiceName: "unused",
					ServiceTags: []string{"prod"},
					ServiceMeta: map[string]string{
						"urlprefix-0": " \t ",
						"urlprefix-1": "",
						"other":       "/unrelated",
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got, want := c.r.build(), c.cfg; !reflect.DeepEqual(got, want) {
				t.Fatalf("\ngot  %#v\nwant %#v", got, want)
			}
		})
	}
}

func TestParseTag(t *testing.T) {
	prefix := "p-"
	tests := []struct {
		tag   string
		env   map[string]string
		route string
		opts  string
		ok    bool
	}{
		{tag: "p", route: "", ok: false},
		{tag: "p-", route: "", ok: true},
		{tag: "p- ", route: "", ok: true},
		{tag: "p-/", route: "/", ok: true},
		{tag: " p-/", route: "/", ok: true},
		{tag: "p-/ ", route: "/", ok: true},
		{tag: "p- / ", route: "/", ok: true},
		{tag: "p-/foo", route: "/foo", ok: true},
		{tag: "p- /foo", route: "/foo", ok: true},
		{tag: "p-1.1.1.1:999", route: "1.1.1.1:999", ok: true},
		{tag: "p-bar/foo", route: "bar/foo", ok: true},
		{tag: "p-bar/foo/foo", route: "bar/foo/foo", ok: true},
		{tag: "p-www.bar.com/foo/foo", route: "www.bar.com/foo/foo", ok: true},
		{tag: "p-WWW.BAR.COM/foo/foo", route: "www.bar.com/foo/foo", ok: true},
		{tag: "p-bar/foo a b c", route: "bar/foo", opts: "a b c", ok: true},
		{
			tag:   "p-$x/$y",
			route: "/",
			ok:    true,
		},
		{
			tag:   "p-${x}/${y}",
			route: "/",
			ok:    true,
		},
		{
			tag:   "p-$x/$Y",
			env:   map[string]string{"x": "Xx", "Y": "Yy"},
			route: "xx/Yy",
			ok:    true,
		},
		{
			tag:   "p-${x}/${Y}",
			env:   map[string]string{"x": "Xx", "Y": "Yy"},
			route: "xx/Yy",
			ok:    true,
		},
		{
			tag:   "p-www.bar.com:80/foo redirect=302,https://www.bar.com",
			route: "www.bar.com:80/foo",
			opts:  "redirect=302,https://www.bar.com",
			ok:    true,
		},
	}

	for i, tt := range tests {
		uri, opts, ok := parseURLPrefixTag(tt.tag, prefix, tt.env)
		if got, want := ok, tt.ok; got != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
		if !ok {
			continue
		}
		if got, want := uri, tt.route; got != want {
			t.Errorf("%d: got uri %q want %q", i, got, want)
		}
		if got, want := opts, tt.opts; got != want {
			t.Errorf("%d: got opts %q want %q", i, got, want)
		}
	}
}
