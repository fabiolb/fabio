package route

import (
	"testing"

	"github.com/gobwas/glob"
)

func TestPrefixMatcher(t *testing.T) {
	tests := []struct {
		uri     string
		matches bool
		route   *Route
	}{
		{uri: "/foo", matches: true, route: &Route{Path: "/foo"}},
		{uri: "/fools", matches: true, route: &Route{Path: "/foo"}},
		{uri: "/fo", matches: false, route: &Route{Path: "/foo"}},
		{uri: "/bar", matches: false, route: &Route{Path: "/foo"}},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			if got, want := prefixMatcher(tt.uri, tt.route), tt.matches; got != want {
				t.Fatalf("got %v want %v", got, want)
			}
		})
	}
}

func TestGlobMatcher(t *testing.T) {
	tests := []struct {
		uri     string
		matches bool
		route   *Route
	}{
		// happy flows
		{uri: "/foo", matches: true, route: &Route{Path: "/foo"}},
		{uri: "/fool", matches: true, route: &Route{Path: "/foo?"}},
		{uri: "/fool", matches: true, route: &Route{Path: "/foo*"}},
		{uri: "/fools", matches: true, route: &Route{Path: "/foo*"}},
		{uri: "/fools", matches: true, route: &Route{Path: "/foo*"}},
		{uri: "/foo/x/bar", matches: true, route: &Route{Path: "/foo/*/bar"}},
		{uri: "/foo/x/y/z/w/bar", matches: true, route: &Route{Path: "/foo/**"}},
		{uri: "/foo/x/y/z/w/bar", matches: true, route: &Route{Path: "/foo/**/bar"}},

		// error flows
		{uri: "/fo", matches: false, route: &Route{Path: "/foo"}},
		{uri: "/fools", matches: false, route: &Route{Path: "/foo"}},
		{uri: "/fo", matches: false, route: &Route{Path: "/foo*"}},
		{uri: "/fools", matches: false, route: &Route{Path: "/foo.*"}},
		{uri: "/foo/x/y/z/w/baz", matches: false, route: &Route{Path: "/foo/**/bar"}},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			tt.route.Glob = glob.MustCompile(tt.route.Path)
			if got, want := globMatcher(tt.uri, tt.route), tt.matches; got != want {
				t.Fatalf("got %v want %v", got, want)
			}
		})
	}
}

func TestIPrefixMatcher(t *testing.T) {
	tests := []struct {
		uri     string
		matches bool
		route   *Route
	}{
		{uri: "/foo", matches: false, route: &Route{Path: "/fool"}},
		{uri: "/foo", matches: true, route: &Route{Path: "/foo"}},
		{uri: "/Fool", matches: true, route: &Route{Path: "/foo"}},
		{uri: "/foo", matches: true, route: &Route{Path: "/Foo"}},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			if got, want := iPrefixMatcher(tt.uri, tt.route), tt.matches; got != want {
				t.Fatalf("got %v want %v", got, want)
			}
		})
	}
}
