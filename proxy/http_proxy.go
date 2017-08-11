package proxy

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/logger"
	"github.com/fabiolb/fabio/metrics"
	"github.com/fabiolb/fabio/proxy/gzip"
	"github.com/fabiolb/fabio/route"
	"github.com/fabiolb/fabio/uuid"
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

	// InsecureTransport is the http connection pool configured with
	// InsecureSkipVerify set. This is used for https proxies with
	// self-signed certs.
	InsecureTransport http.RoundTripper

	// Lookup returns a target host for the given request.
	// The proxy will panic if this value is nil.
	Lookup func(*http.Request) *route.Target

	// Requests is the metric name which times every request.
	Requests string

	// Logger is the access logger for the requests.
	Logger logger.Logger

	// UUID returns a unique id in uuid format.
	// If UUID is nil, uuid.NewUUID() is used.
	UUID func() string
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

	if t.Host == "dst" {
		r.Host = targetURL.Host
	}

	// TODO(fs): The HasPrefix check seems redundant since the lookup function should
	// TODO(fs): have found the target based on the prefix but there may be other
	// TODO(fs): matchers which may have different rules. I'll keep this for
	// TODO(fs): a defensive approach.
	if t.StripPath != "" && strings.HasPrefix(r.URL.Path, t.StripPath) {
		targetURL.Path = targetURL.Path[len(t.StripPath):]
	}

	if err := addHeaders(r, p.Config, t.StripPath); err != nil {
		http.Error(w, "cannot parse "+r.RemoteAddr, http.StatusInternalServerError)
		return
	}

	if p.Config.RequestID != "" {
		id := p.UUID
		if id == nil {
			id = uuid.NewUUID
		}
		r.Header.Set(p.Config.RequestID, id())
	}

	upgrade, accept := r.Header.Get("Upgrade"), r.Header.Get("Accept")

	tr := p.Transport
	if t.TLSSkipVerify {
		tr = p.InsecureTransport
	}

	var h http.Handler
	switch {
	case upgrade == "websocket" || upgrade == "Websocket":
		if targetURL.Scheme == "https" || targetURL.Scheme == "wss" {
			h = newRawProxy(targetURL, func(network, address string) (net.Conn, error) {
				return tls.Dial(network, address, tr.(*http.Transport).TLSClientConfig)
			})
		} else {
			h = newRawProxy(targetURL, net.Dial)
		}

	case accept == "text/event-stream":
		// use the flush interval for SSE (server-sent events)
		// must be > 0s to be effective
		h = newHTTPProxy(targetURL, tr, p.Config.FlushInterval)

	default:
		h = newHTTPProxy(targetURL, tr, time.Duration(0))
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

	metrics.TimeService(p.Requests, dur)
	metrics.TimeService(t.TimerName, dur)

	// get response and update metrics
	rp, ok := h.(*httputil.ReverseProxy)
	if !ok {
		return
	}
	rpt, ok := rp.Transport.(*transport)
	if !ok {
		return
	}
	if rpt.resp == nil {
		return
	}
	metrics.TimeDefault(key(rpt.resp.StatusCode), dur)

	// write access log
	if p.Logger != nil {
		p.Logger.Log(&logger.Event{
			Start:           start,
			End:             end,
			Request:         r,
			Response:        rpt.resp,
			RequestURL:      requestURL,
			UpstreamAddr:    targetURL.Host,
			UpstreamService: t.Service,
			UpstreamURL:     targetURL,
		})
	}
}

func key(code int) string {
	b := []byte("http.status.")
	b = strconv.AppendInt(b, int64(code), 10)
	return string(b)
}
