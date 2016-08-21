package proxy

import (
	"net/http"
	"time"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/metrics"
)

// httpProxy is a dynamic reverse proxy for HTTP and HTTPS protocols.
type httpProxy struct {
	tr       http.RoundTripper
	cfg      config.Proxy
	requests metrics.Timer
	noroute  metrics.Counter
}

func NewHTTPProxy(tr http.RoundTripper, cfg config.Proxy) http.Handler {
	return &httpProxy{
		tr:       tr,
		cfg:      cfg,
		requests: metrics.DefaultRegistry.GetTimer("requests"),
		noroute:  metrics.DefaultRegistry.GetCounter("notfound"),
	}
}

func (p *httpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ShuttingDown() {
		http.Error(w, "shutting down", http.StatusServiceUnavailable)
		return
	}

	t := target(r)
	if t == nil {
		p.noroute.Inc(1)
		w.WriteHeader(p.cfg.NoRouteStatus)
		return
	}

	if err := addHeaders(r, p.cfg); err != nil {
		http.Error(w, "cannot parse "+r.RemoteAddr, http.StatusInternalServerError)
		return
	}

	var h http.Handler
	switch {
	case r.Header.Get("Upgrade") == "websocket":
		h = newRawProxy(t.URL)

		// To use the filtered proxy use
		// h = newWSProxy(t.URL)

	case r.Header.Get("Accept") == "text/event-stream":
		// use the flush interval for SSE (server-sent events)
		// must be > 0s to be effective
		h = newHTTPProxy(t.URL, p.tr, p.cfg.FlushInterval)

	default:
		h = newHTTPProxy(t.URL, p.tr, time.Duration(0))
	}

	start := time.Now()
	h.ServeHTTP(w, r)
	p.requests.UpdateSince(start)
	t.Timer.UpdateSince(start)
}
