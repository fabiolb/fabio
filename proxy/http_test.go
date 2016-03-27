package proxy

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// Test, that the proxy calls the target
// with the target host as host header,
// and not with the ip of the request to the proxy itself
func TestCorrectHostHeader(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Host))
	}))

	defer server.Close()
	serverUrl, _ := url.Parse(server.URL)

	tr := &http.Transport{
		Dial: (&net.Dialer{}).Dial,
	}
	proxy := newHTTPProxy(serverUrl, tr)

	req, _ := http.NewRequest("GET", "http://example.com:666", nil)
	res := httptest.NewRecorder()

	proxy.ServeHTTP(res, req)

	if serverUrl.Host != res.Body.String() {
		t.Errorf("expected host was %v, but got %v", serverUrl.Host, res.Body.String())
	}
}
