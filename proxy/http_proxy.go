package proxy

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/logger"
	"github.com/eBay/fabio/metrics"
	"github.com/eBay/fabio/proxy/gzip"
	"github.com/eBay/fabio/route"
)

// HTTPProxy is a dynamic reverse proxy for HTTP and HTTPS protocols.
type HTTPProxy struct {
	// Config is the proxy configuration as provided during startup.
	Config config.Proxy

	// Time returns the current time as the number of seconds since the epoch.
	// If Time is nil, time.Now is used.
	Time func() time.Time

	// Transport is the http connection pool configured with timeouts.
	// The proxy will panic if this value is nil.
	Transport http.RoundTripper

	// Lookup returns a target host for the given request.
	// The proxy will panic if this value is nil.
	Lookup func(*http.Request) *route.Target

	// Requests is a timer metric which is updated for every request.
	Requests metrics.Timer

	// Noroute is a counter metric which is updated for every request
	// where Lookup() returns nil.
	Noroute metrics.Counter

	// Logger is the access logger for the requests.
	Logger logger.Logger
}

func (p *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.Lookup == nil {
		panic("no lookup function")
	}

	t := p.Lookup(r)
	if t == nil {
		w.WriteHeader(p.Config.NoRouteStatus)
		return
	}

	if err := addHeaders(r, p.Config); err != nil {
		http.Error(w, "cannot parse "+r.RemoteAddr, http.StatusInternalServerError)
		return
	}

	// build the request url since r.URL will get modified
	// by the reverse proxy and contains only the RequestURI anyway
	requestURL := &url.URL{
		Scheme:   scheme(r),
		Host:     r.Host,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	// build the real target url that is passed to the proxy
	targetURL := &url.URL{
		Scheme: t.URL.Scheme,
		Host:   t.URL.Host,
		Path:   r.URL.Path,
	}
	if t.URL.RawQuery == "" || r.URL.RawQuery == "" {
		targetURL.RawQuery = t.URL.RawQuery + r.URL.RawQuery
	} else {
		targetURL.RawQuery = t.URL.RawQuery + "&" + r.URL.RawQuery
	}

	// TODO(fs): The HasPrefix check seems redundant since the lookup function should
	// TODO(fs): have found the target based on the prefix but there may be other
	// TODO(fs): matchers which may have different rules. I'll keep this for
	// TODO(fs): a defensive approach.
	if t.StripPath != "" && strings.HasPrefix(r.URL.Path, t.StripPath) {
		targetURL.Path = targetURL.Path[len(t.StripPath):]
	}

	upgrade, accept := r.Header.Get("Upgrade"), r.Header.Get("Accept")

	var h http.Handler
	switch {
	case upgrade == "websocket" || upgrade == "Websocket":
		h = newRawProxy(targetURL)

	case accept == "text/event-stream":
		// use the flush interval for SSE (server-sent events)
		// must be > 0s to be effective
		h = newHTTPProxy(targetURL, p.Transport, p.Config.FlushInterval)

	default:
		h = newHTTPProxy(targetURL, p.Transport, time.Duration(0))
	}

	if p.Config.GZIPContentTypes != nil {
		h = gzip.NewGzipHandler(h, p.Config.GZIPContentTypes)
	}

	timeNow := p.Time
	if timeNow == nil {
		timeNow = time.Now
	}

	start := timeNow()
	h.ServeHTTP(w, r)
	end := timeNow()
	dur := end.Sub(start)

	if p.Requests != nil {
		p.Requests.Update(dur)
	}
	if t.Timer != nil {
		t.Timer.Update(dur)
	}

	// get response and update metrics
	hr, ok := h.(responseKeeper)
	if !ok {
		return
	}
	resp := hr.response()
	if resp == nil {
		return
	}
	metrics.DefaultRegistry.GetTimer(key(resp.StatusCode)).Update(dur)

	// write access log
	if p.Logger != nil {
		p.Logger.Log(&logger.Event{
			Start:        start,
			End:          end,
			Request:      r,
			Response:     resp,
			RequestURL:   requestURL,
			UpstreamAddr: targetURL.Host,
			UpstreamURL:  targetURL,
		})
	}
}

func key(code int) string {
	b := []byte("http.status.")
	b = strconv.AppendInt(b, int64(code), 10)
	return string(b)
}
