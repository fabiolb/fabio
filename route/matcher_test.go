package route

import (
	"testing"
)

func TestPrefixMatcher(t *testing.T) {
	routeFoo := &Route{Host: "www.example.com", Path: "/foo"}

	tests := []struct {
		uri   string
		want  bool
		route *Route
	}{
		{"/fo", false, routeFoo},
		{"/foo", true, routeFoo},
		{"/fools", true, routeFoo},
		{"/bar", false, routeFoo},
	}

	for _, tt := range tests {
		if got := prefixMatcher(tt.uri, tt.route); got != tt.want {
			t.Errorf("%s: got %v want %v", tt.uri, got, tt.want)
		}
	}
}

func TestGlobMatcher(t *testing.T) {
	routeFoo := &Route{Host: "www.example.com", Path: "/foo"}
	routeFooWild := &Route{Host: "www.example.com", Path: "/foo.*"}

	tests := []struct {
		uri   string
		want  bool
		route *Route
	}{
		{"/fo", false, routeFoo},
		{"/foo", true, routeFoo},
		{"/fools", false, routeFoo},
		{"/bar", false, routeFoo},

		{"/fo", false, routeFooWild},
		{"/foo", false, routeFooWild},
		{"/fools", false, routeFooWild},
		{"/foo.", true, routeFooWild},
		{"/foo.a", true, routeFooWild},
		{"/foo.bar", true, routeFooWild},
		{"/foo.bar.baz", true, routeFooWild},
	}

	for _, tt := range tests {
		if got := globMatcher(tt.uri, tt.route); got != tt.want {
			t.Errorf("%s: got %v want %v", tt.uri, got, tt.want)
		}
	}
}
