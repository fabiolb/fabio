package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func newHTTPProxy(t *url.URL, tr http.RoundTripper, flush time.Duration) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(t)
	rp.Transport = tr
	rp.FlushInterval = flush
	return rp
}
