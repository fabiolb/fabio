package consul

import "testing"

func TestParseTag(t *testing.T) {
	prefix := "p-"
	tests := []struct {
		tag  string
		env  map[string]string
		host string
		path string
		ok   bool
	}{
		{tag: "p", host: "", path: "", ok: false},
		{tag: "p-", host: "", path: "", ok: false},
		{tag: "p- ", host: "", path: "", ok: false},
		{tag: "p-/", host: "", path: "/", ok: true},
		{tag: " p-/", host: "", path: "/", ok: true},
		{tag: "p-/ ", host: "", path: "/", ok: true},
		{tag: "p- / ", host: "", path: "/", ok: true},
		{tag: "p-/foo", host: "", path: "/foo", ok: true},
		{tag: "p-bar/foo", host: "bar", path: "/foo", ok: true},
		{tag: "p-bar/foo/foo", host: "bar", path: "/foo/foo", ok: true},
		{tag: "p-www.bar.com/foo/foo", host: "www.bar.com", path: "/foo/foo", ok: true},
		{tag: "p-WWW.BAR.COM/foo/foo", host: "www.bar.com", path: "/foo/foo", ok: true},
		{
			tag:  "p-$x/$y",
			host: "", path: "/",
			ok: true,
		},
		{
			tag:  "p-${x}/${y}",
			host: "", path: "/",
			ok: true,
		},
		{
			tag:  "p-$x/$Y",
			env:  map[string]string{"x": "Xx", "Y": "Yy"},
			host: "xx", path: "/Yy",
			ok: true,
		},
		{
			tag:  "p-${x}/${Y}",
			env:  map[string]string{"x": "Xx", "Y": "Yy"},
			host: "xx", path: "/Yy",
			ok: true,
		},
	}

	for i, tt := range tests {
		host, path, ok := parseURLPrefixTag(tt.tag, prefix, tt.env)
		if got, want := ok, tt.ok; got != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
		if !ok {
			continue
		}
		if got, want := host, tt.host; got != want {
			t.Errorf("%d: got host %q want %q", i, got, want)
		}
		if got, want := path, tt.path; got != want {
			t.Errorf("%d: got path %q want %q", i, got, want)
		}
	}
}
