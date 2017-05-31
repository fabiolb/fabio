package route

import (
	"net/url"

	"github.com/fabiolb/fabio/metrics"
)

type Target struct {
	// Service is the name of the service the targetURL points to
	Service string

	// Tags are the list of tags for this target
	Tags []string

	// StripPath will be removed from the front of the outgoing
	// request path
	StripPath string

	// TLSSkipVerify disables certificate validation for upstream
	// TLS connections.
	TLSSkipVerify bool

	// Host signifies what the proxy will set the host header to.
	// The proxy does nothing to the host header by default. (src)
	// Setting Host=dst will signify to the proxy that the host header
	// should be reset to the destination service.
	Host string

	// URL is the endpoint the service instance listens on
	URL *url.URL

	// FixedWeight is the weight assigned to this target.
	// If the value is 0 the targets weight is dynamic.
	FixedWeight float64

	// Weight is the actual weight for this service in percent.
	Weight float64

	// Timer measures throughput and latency of this target
	Timer metrics.Timer

	// timerName is the name of the timer in the metrics registry
	timerName string
}
