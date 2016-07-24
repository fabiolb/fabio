package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func newHTTPProxy(target *url.URL, tr http.RoundTripper, flush time.Duration) http.Handler {
	rp := &httputil.ReverseProxy{
		// this is a simplified director function based on the
		// httputil.NewSingleHostReverseProxy() which does not
		// mangle the request and target URL since the target
		// URL is already in the correct format.
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = target.Path
			req.URL.RawQuery = target.RawQuery
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		},
		FlushInterval: flush,
		Transport:     &transport{tr, nil},
	}
	return &httpHandler{rp}
}

// responseKeeper exposes the response from an HTTP request.
type responseKeeper interface {
	response() *http.Response
}

// httpHandler is a simple wrapper around a reverse proxy to access the
// captured response object in the underlying transport object. There
// may be a better way of doing this.
type httpHandler struct {
	rp *httputil.ReverseProxy
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.rp.ServeHTTP(w, r)
}

func (h *httpHandler) response() *http.Response {
	return h.rp.Transport.(*transport).resp
}

// transport executes the roundtrip and captures the response. It is not
// safe for multiple or concurrent use since it only captures a single
// response.
type transport struct {
	http.RoundTripper
	resp *http.Response
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := t.RoundTripper.RoundTrip(r)
	t.resp = resp
	return resp, err
}
