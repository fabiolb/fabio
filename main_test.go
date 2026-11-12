package main

import (
	"sync"
	"testing"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/proxy"
)

func TestNewHTTPProxyCopiesProtectHeaders(t *testing.T) {
	cfg, err := config.Load([]string{"fabio"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Proxy.RequestID = "X-Request-Id"
	cfg.Proxy.ClientIPHeader = "X-Client-Ip"
	cfg.Proxy.TLSHeader = "X-Tls"

	// startServers builds one proxy per HTTP listener, each in its own goroutine.
	proxies := make([]*proxy.HTTPProxy, 4)
	var wg sync.WaitGroup
	for i := range proxies {
		wg.Go(func() {
			proxies[i] = newHTTPProxy(cfg, &proxy.HttpStatsHandler{})
		})
	}
	wg.Wait()

	for _, h := range []string{"X-Request-Id", "X-Client-Ip", "X-Tls"} {
		if proxy.DefaultProtectHeaders[h] {
			t.Errorf("DefaultProtectHeaders was modified: contains %s", h)
		}
		for i, p := range proxies {
			if !p.ProtectHeaders[h] {
				t.Errorf("proxy %d does not protect %s", i, h)
			}
		}
	}
}
