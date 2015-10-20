package proxy

import (
	"net/http"
	"time"

	gometrics "github.com/eBay/fabio/_third_party/github.com/rcrowley/go-metrics"
	"github.com/eBay/fabio/config"
)

// Proxy is a dynamic reverse proxy.
type Proxy struct {
	httpProxy http.Handler
	wsProxy   http.Handler
	requests  gometrics.Timer
}

func New(tr http.RoundTripper, cfg config.Proxy) *Proxy {
	return &Proxy{
		httpProxy: newHTTPProxy(tr, cfg),
		wsProxy:   newWSProxy(),
		requests:  gometrics.GetOrRegisterTimer("requests", gometrics.DefaultRegistry),
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ShuttingDown() {
		http.Error(w, "shutting down", http.StatusServiceUnavailable)
		return
	}

	// http or ws?
	h := p.httpProxy
	if r.Header.Get("Upgrade") == "websocket" {
		h = p.wsProxy
	}

	start := time.Now()
	h.ServeHTTP(w, r)
	p.requests.UpdateSince(start)
}
