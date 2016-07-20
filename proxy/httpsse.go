package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func newHTTPSSEProxy(t *url.URL, tr http.RoundTripper) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(t)
	rp.Transport = tr
	rp.FlushInterval = 5 * time.Second
	return rp
}
