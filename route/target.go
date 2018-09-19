package route

import (
	"net/url"
	"strings"

	"github.com/fabiolb/fabio/metrics"
)

type Target struct {
	// Service is the name of the service the targetURL points to
	Service string

	// Tags are the list of tags for this target
	Tags []string

	// Opts is the raw options for the target.
	Opts map[string]string

	// StripPath will be removed from the front of the outgoing
	// request path
	StripPath string

	// TLSSkipVerify disables certificate validation for upstream
	// TLS connections.
	TLSSkipVerify bool

	// Host signifies what the proxy will set the Host header to.
	// The proxy does not modify the Host header by default.
	// When Host is set to 'dst' the proxy will use the host name
	// of the target host for the outgoing request.
	Host string

	// URL is the endpoint the service instance listens on
	URL *url.URL

	// RedirectCode is the HTTP status code used for redirects.
	// When set to a value > 0 the client is redirected to the target url.
	RedirectCode int

	// RedirectURL is the redirect target based on the request.
	// This is cached here to prevent multiple generations per request.
	RedirectURL *url.URL

	// FixedWeight is the weight assigned to this target.
	// If the value is 0 the targets weight is dynamic.
	FixedWeight float64

	// Weight is the actual weight for this service in percent.
	Weight float64

	// Timer measures throughput and latency of this target
	Timer metrics.Timer

	// TimerName is the name of the timer in the metrics registry
	TimerName string

	// accessRules is map of access information for the target.
	accessRules map[string][]interface{}
}

func (t *Target) BuildRedirectURL(requestURL *url.URL) {
	t.RedirectURL = &url.URL{
		Scheme:   t.URL.Scheme,
		Host:     t.URL.Host,
		Path:     t.URL.Path,
		RawQuery: t.URL.RawQuery,
	}
	if strings.HasSuffix(t.RedirectURL.Host, "$path") {
		t.RedirectURL.Host = t.RedirectURL.Host[:len(t.RedirectURL.Host)-len("$path")]
		t.RedirectURL.Path = "$path"
	}
	if strings.Contains(t.RedirectURL.Path, "/$path") {
		t.RedirectURL.Path = strings.Replace(t.RedirectURL.Path, "/$path", "$path", 1)
	}
	if strings.Contains(t.RedirectURL.Path, "$path") {
		t.RedirectURL.Path = strings.Replace(t.RedirectURL.Path, "$path", requestURL.Path, 1)
		if t.StripPath != "" && strings.HasPrefix(t.RedirectURL.Path, t.StripPath) {
			t.RedirectURL.Path = t.RedirectURL.Path[len(t.StripPath):]
		}
		if t.RedirectURL.RawQuery == "" && requestURL.RawQuery != "" {
			t.RedirectURL.RawQuery = requestURL.RawQuery
		}
	}
	if t.RedirectURL.Path == "" {
		t.RedirectURL.Path = "/"
	}
	if strings.Contains(t.RedirectURL.Host, "$host") {
		t.RedirectURL.Host = strings.Replace(t.RedirectURL.Host, "$host", requestURL.Host, 1)
	}
}
