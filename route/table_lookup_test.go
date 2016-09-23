package route

import (
	"crypto/tls"
	"net/http"
	"testing"
)

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
	`

	tbl, err := ParseTable(s)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		req *http.Request
		dst string
	}{
		// match on host and path with and without trailing slash
		{&http.Request{Host: "abc.com", RequestURI: "/"}, "http://foo.com:1000"},
		{&http.Request{Host: "abc.com", RequestURI: "/bar"}, "http://foo.com:1000"},
		{&http.Request{Host: "abc.com", RequestURI: "/foo"}, "http://foo.com:1500"},
		{&http.Request{Host: "abc.com", RequestURI: "/foo/"}, "http://foo.com:2000"},
		{&http.Request{Host: "abc.com", RequestURI: "/foo/bar"}, "http://foo.com:2500"},
		{&http.Request{Host: "abc.com", RequestURI: "/foo/bar/"}, "http://foo.com:3000"},

		// do not match on host but maybe on path
		{&http.Request{Host: "def.com", RequestURI: "/"}, "http://foo.com:800"},
		{&http.Request{Host: "def.com", RequestURI: "/bar"}, "http://foo.com:800"},
		{&http.Request{Host: "def.com", RequestURI: "/foo"}, "http://foo.com:900"},

		// strip default port
		{&http.Request{Host: "abc.com:80", RequestURI: "/"}, "http://foo.com:1000"},
		{&http.Request{Host: "abc.com:443", RequestURI: "/", TLS: &tls.ConnectionState{}}, "http://foo.com:1000"},

		// not using default port
		{&http.Request{Host: "abc.com:443", RequestURI: "/"}, "http://foo.com:800"},
		{&http.Request{Host: "abc.com:80", RequestURI: "/", TLS: &tls.ConnectionState{}}, "http://foo.com:800"},
	}

	for i, tt := range tests {
		if got, want := tbl.Lookup(tt.req, "").URL.String(), tt.dst; got != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}
