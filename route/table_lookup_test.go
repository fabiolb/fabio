package route

import (
	"net/http"
	"testing"
)

func TestTableLookup(t *testing.T) {
	s := `
	route add svc / http://foo.com:800
	route add svc /foo http://foo.com:900
	route add svc abc.com/ http://foo.com:1000
	route add svc abc.com/foo http://foo.com:1500
	route add svc abc.com/foo/ http://foo.com:2000
	route add svc abc.com/foo/bar http://foo.com:2500
	route add svc abc.com/foo/bar/ http://foo.com:3000
	`

	tbl, err := ParseString(s)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		req *http.Request
		dst string
	}{
		{&http.Request{Host: "abc.com", RequestURI: "/"}, "http://foo.com:1000"},
		{&http.Request{Host: "abc.com", RequestURI: "/bar"}, "http://foo.com:1000"},
		{&http.Request{Host: "abc.com", RequestURI: "/foo"}, "http://foo.com:1500"},
		{&http.Request{Host: "abc.com", RequestURI: "/foo/"}, "http://foo.com:2000"},
		{&http.Request{Host: "abc.com", RequestURI: "/foo/bar"}, "http://foo.com:2500"},
		{&http.Request{Host: "abc.com", RequestURI: "/foo/bar/"}, "http://foo.com:3000"},

		{&http.Request{Host: "def.com", RequestURI: "/"}, "http://foo.com:800"},
		{&http.Request{Host: "def.com", RequestURI: "/bar"}, "http://foo.com:800"},
		{&http.Request{Host: "def.com", RequestURI: "/baz"}, "http://foo.com:800"},

		{&http.Request{Host: "def.com", RequestURI: "/foo"}, "http://foo.com:900"},
	}

	for i, tt := range tests {
		if got, want := tbl.Lookup(tt.req, "").URL.String(), tt.dst; got != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}
