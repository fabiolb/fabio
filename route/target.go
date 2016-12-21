package route

import (
	"net/url"

	"github.com/eBay/fabio/metrics"
)

type Target struct {
	// Service is the name of the service the targetURL points to
	Service string

	// Tags are the list of tags for this target
	Tags []string

	// StripPath will be removed from the front of the outgoing
	// request path
	StripPath string

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
