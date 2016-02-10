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

	route add svc /del http://foo.com:950
	route del svc /del
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

		{&http.Request{Host: "", RequestURI: "/del"}, ""},
	}

	for i, tt := range tests {
		var targetURL string
		tg := tbl.Lookup(tt.req, "")
		if tg != nil {
			targetURL = tg.URL.String()
		}

		if got, want := targetURL, tt.dst; got != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}
