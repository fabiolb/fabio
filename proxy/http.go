package proxy

import (
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/eBay/fabio/config"
)

type httpProxy struct {
	tr  http.RoundTripper
	cfg config.Proxy
}

func newHTTPProxy(tr http.RoundTripper, cfg config.Proxy) http.Handler {
	return &httpProxy{tr, cfg}
}

func (p *httpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := target(r)
	if t == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := addHeaders(r, p.cfg); err != nil {
		http.Error(w, "cannot parse "+r.RemoteAddr, http.StatusInternalServerError)
		return
	}

	start := time.Now()
	rp := httputil.NewSingleHostReverseProxy(t.URL)
	rp.Transport = p.tr
	rp.ServeHTTP(w, r)
	t.Timer.UpdateSince(start)
}
