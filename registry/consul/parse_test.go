package consul

import "testing"

func TestParseTag(t *testing.T) {
	prefix := "p-"
	tests := []struct {
		tag, host, path string
		ok              bool
	}{
		{"p", "", "", false},
		{"p-", "", "", false},
		{"p- ", "", "", false},
		{"p-/", "", "/", true},
		{"p-/ ", "", "/", true},
		{"p- / ", "", "/", true},
		{"p-/foo", "", "/foo", true},
		{"p-bar/foo", "bar", "/foo", true},
		{"p-bar/foo/foo", "bar", "/foo/foo", true},
		{"p-www.bar.com/foo/foo", "www.bar.com", "/foo/foo", true},
		{"p-WWW.BAR.COM/foo/foo", "www.bar.com", "/foo/foo", true},
	}

	for i, tt := range tests {
		host, path, ok := parseURLPrefixTag(tt.tag, prefix)
		if got, want := ok, tt.ok; got != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
		if !ok {
			continue
		}
		if got, want := host, tt.host; got != want {
			t.Errorf("%d: got %s want %s", i, got, want)
		}
		if got, want := path, tt.path; got != want {
			t.Errorf("%d: got %s want %s", i, got, want)
		}
	}
}
