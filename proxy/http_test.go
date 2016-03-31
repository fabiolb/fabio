package proxy

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestCorrectHostHeader checks that the proxy calls the target
// with the target host as host header,
// and not with the ip of the request to the proxy itself
func TestCorrectHostHeader(t *testing.T) {
	got := "not called"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Host
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	tr := &http.Transport{
		Dial: (&net.Dialer{}).Dial,
	}
	proxy := newHTTPProxy(serverURL, tr)

	req, err := http.NewRequest("GET", "http://example.com:666", nil)
	if err != nil {
		t.Fatal(err)
	}

	proxy.ServeHTTP(httptest.NewRecorder(), req)

	if want := serverURL.Host; want != got {
		t.Errorf("want host %q, but got %q", want, got)
	}
}
