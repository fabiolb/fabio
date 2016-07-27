package proxy

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/route"
)

func TestProxyProducesCorrectXffHeader(t *testing.T) {
	got := "not called"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Forwarded-For")
	}))
	defer server.Close()

	table := make(route.Table)
	table.AddRoute("mock", "/", server.URL, 1, nil)
	route.SetTable(table)

	tr := &http.Transport{Dial: (&net.Dialer{}).Dial}
	proxy := New(tr, config.Proxy{LocalIP: "1.1.1.1", ClientIPHeader: "X-Forwarded-For"})

	req := &http.Request{
		RequestURI: "/",
		Header:     http.Header{"X-Forwarded-For": {"3.3.3.3"}},
		RemoteAddr: "2.2.2.2:666",
		URL:        &url.URL{},
	}

	proxy.ServeHTTP(httptest.NewRecorder(), req)

	if want := "3.3.3.3, 2.2.2.2"; got != want {
		t.Errorf("got %v, but want %v", got, want)
	}
}

func TestProxyNoRouteStaus(t *testing.T) {
	route.SetTable(make(route.Table))
	tr := &http.Transport{Dial: (&net.Dialer{}).Dial}
	cfg := config.Proxy{NoRouteStatus: 999}
	proxy := New(tr, cfg)
	req := &http.Request{
		RequestURI: "/",
		URL:        &url.URL{},
	}

	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)
	if got, want := rec.Code, cfg.NoRouteStatus; got != want {
		t.Fatalf("got %d want %d", got, want)
	}
}

func BenchmarkProxyLogger(b *testing.B) {
	got := "not called"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Forwarded-For")
	}))
	defer server.Close()

	table := make(route.Table)
	table.AddRoute("mock", "/", server.URL, 1, nil)
	route.SetTable(table)

	tr := &http.Transport{Dial: (&net.Dialer{}).Dial}
	proxy := New(tr, config.Proxy{LocalIP: "1.1.1.1", ClientIPHeader: "X-Forwarded-For", Log: config.Log{
		Format: "remote_addr time request body_bytes_sent http_referer http_user_agent server_name proxy_endpoint response_time request_args ",
		Target: "stdout",
	}})

	req := &http.Request{
		RequestURI: "/",
		Header:     http.Header{"X-Forwarded-For": {"1.2.3.4"}},
		RemoteAddr: "2.2.2.2:666",
		URL:        &url.URL{},
		Method:     "GET",
		Proto:      "HTTP/1.1",
	}

	for i := 0; i < b.N; i++ {
		proxy.ServeHTTP(httptest.NewRecorder(), req)
	}

}
