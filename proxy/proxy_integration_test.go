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
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Forwarded-For")
	}))
	defer mockServer.Close()

	setRoutingTable(mockServer.URL)
	tr := &http.Transport{Dial: (&net.Dialer{}).Dial}
	proxy := New(tr, config.Proxy{LocalIP: "1.1.1.1", ClientIPHeader: "X-Forwarded-For"})

	req := &http.Request{
		RequestURI: "/",
		Header:     http.Header{"X-Forwarded-For": {"3.3.3.3"}},
		RemoteAddr: "2.2.2.2:666",
		URL:        &url.URL{},
	}

	proxy.ServeHTTP(httptest.NewRecorder(), req)

	if want := "3.3.3.3, 2.2.2.2"; want != got {
		t.Errorf("want %v, but got %v", want, got)
	}
}

func setRoutingTable(targetURL string) {
	table := make(route.Table)
	table.AddRoute("mock", "/", targetURL, 1, nil)
	route.SetTable(table)
}
