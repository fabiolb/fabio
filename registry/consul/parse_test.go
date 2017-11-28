package consul

import "testing"

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
		{tag: "p-", route: "", ok: false},
		{tag: "p- ", route: "", ok: false},
		{tag: "p-/", route: "/", ok: true},
		{tag: " p-/", route: "/", ok: true},
		{tag: "p-/ ", route: "/", ok: true},
		{tag: "p- / ", route: "/", ok: true},
		{tag: "p-/foo", route: "/foo", ok: true},
		{tag: "p- /foo", route: "/foo", ok: true},
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
