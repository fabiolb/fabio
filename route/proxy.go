package route

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	gometrics "github.com/eBay/fabio/_third_party/github.com/rcrowley/go-metrics"
)

// Proxy is a dynamic reverse proxy.
type Proxy struct {
	tr             http.RoundTripper
	clientIPHeader string
	tlsHeader      string
	tlsHeaderValue string
	requests       gometrics.Timer
}

func NewProxy(tr http.RoundTripper, clientIPHeader, tlsHeader, tlsHeaderValue string) *Proxy {
	return &Proxy{
		tr:             tr,
		clientIPHeader: clientIPHeader,
		tlsHeader:      tlsHeader,
		tlsHeaderValue: tlsHeaderValue,
		requests:       gometrics.GetOrRegisterTimer("requests", gometrics.DefaultRegistry),
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if ShuttingDown() {
		http.Error(w, "shutting down", http.StatusServiceUnavailable)
		return
	}

	target := GetTable().lookup(req, req.Header.Get("trace"))
	if target == nil {
		log.Print("[WARN] No route for ", req.URL)
		w.WriteHeader(404)
		return
	}

	if p.clientIPHeader != "" {
		ip, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			http.Error(w, "cannot parse "+req.RemoteAddr, http.StatusInternalServerError)
			return
		}
		req.Header.Set(p.clientIPHeader, ip)
	}

	if p.tlsHeader != "" && req.TLS != nil {
		req.Header.Set(p.tlsHeader, p.tlsHeaderValue)
	}

	start := time.Now()
	rp := httputil.NewSingleHostReverseProxy(target.URL)
	rp.Transport = p.tr
	rp.ServeHTTP(w, req)
	target.timer.UpdateSince(start)
	p.requests.UpdateSince(start)
}
