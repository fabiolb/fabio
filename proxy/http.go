package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func newHTTPProxy(t *url.URL, tr http.RoundTripper) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(t)

	originalDirector := rp.Director

	rp.Director = func(req *http.Request) {
		originalDirector(req)
		// set the target Host as Host header,
		// to allow virtual servers as upstream.
		req.Host = t.Host
	}

	rp.Transport = tr
	return rp
}
