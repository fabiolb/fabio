package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func newHTTPProxy(t *url.URL, tr http.RoundTripper) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(t)

	d := rp.Director

	rp.Director = func(req *http.Request) {
		d(req)
		// Set the target Host as Host header, to allow virtual servers as upstream.
		// Required in HTTP1.1, rfc2616-sec14:
		//     An HTTP/1.1 proxy MUST ensure that any request message
		//     it forwards does contain an appropriate Host header field
		//     that identifies the service being requested by the proxy.
		req.Host = t.Host
	}

	rp.Transport = tr
	return rp
}
