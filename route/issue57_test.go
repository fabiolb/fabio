package route

import (
	"github.com/fabiolb/fabio/config"
	"net/http"
	"testing"
)

// TestIssue57 tests that after deleting a all targets for
// a route requests to that route are handled by the next
// matching route.
func TestIssue57(t *testing.T) {
	tests := []string{
		`
		route add svca / http://foo.com:800
		route add svcb /foo http://foo.com:900
		route del svcb /foo http://foo.com:900`,
		`
		route add svca / http://foo.com:800
	 	route add svcb /foo http://foo.com:900
	 	route del svcb /foo`,
		`
		route add svca / http://foo.com:800
	 	route add svcb /foo http://foo.com:900
	 	route del svcb`,
	}

	//Glob Matching True
	globMatching := config.Config{GlobMatching: true}

	req := &http.Request{URL: mustParse("/foo")}
	want := "http://foo.com:800"

	for i, tt := range tests {
		tbl, err := NewTable(tt)
		if err != nil {
			t.Fatalf("%d: got %v want nil", i, err)
		}
		target := tbl.Lookup(req, "", rrPicker, prefixMatcher, &globMatching)
		if target == nil {
			t.Fatalf("%d: got %v want %v", i, target, want)
		}
		if got := target.URL.String(); got != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}
