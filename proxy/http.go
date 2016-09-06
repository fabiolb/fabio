package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/eBay/fabio/metrics"
)

func newHTTPProxy(t *url.URL, tr http.RoundTripper, flush time.Duration) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(t)
	rp.Transport = tr
	rp.FlushInterval = flush
	rp.Transport = &meteredRoundTripper{tr}
	return rp
}

type meteredRoundTripper struct {
	tr http.RoundTripper
}

func (m *meteredRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := m.tr.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	metrics.DefaultRegistry.GetTimer(name(resp.StatusCode)).UpdateSince(start)
	return resp, nil
}

func name(code int) string {
	b := []byte("http.status.")
	b = strconv.AppendInt(b, int64(code), 10)
	return string(b)
}
