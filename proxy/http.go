package proxy

import (
	"net"
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

	if p.cfg.ClientIPHeader != "" {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "cannot parse "+r.RemoteAddr, http.StatusInternalServerError)
			return
		}
		r.Header.Set(p.cfg.ClientIPHeader, ip)
	}

	if p.cfg.TLSHeader != "" && r.TLS != nil {
		r.Header.Set(p.cfg.TLSHeader, p.cfg.TLSHeaderValue)
	}

	start := time.Now()
	rp := httputil.NewSingleHostReverseProxy(t.URL)
	rp.Transport = p.tr
	rp.ServeHTTP(w, r)
	t.Timer.UpdateSince(start)
}
