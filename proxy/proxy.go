package proxy

import (
	"log"
	"net/http"
	"time"

	gometrics "github.com/eBay/fabio/_third_party/github.com/rcrowley/go-metrics"
	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/route"
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

	h := p.httpProxy
	if r.Header.Get("Upgrade") == "websocket" {
		h = p.wsProxy
	}

	start := time.Now()
	h.ServeHTTP(w, r)
	p.requests.UpdateSince(start)
}

func target(r *http.Request) *route.Target {
	t := route.GetTable().Lookup(r, r.Header.Get("trace"))
	if t == nil {
		log.Print("[WARN] No route for ", r.URL)
	}
	return t
}
