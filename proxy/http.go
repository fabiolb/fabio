package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func newHTTPProxy(t *url.URL, tr http.RoundTripper) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(t)
	rp.Transport = tr
	return rp
}
