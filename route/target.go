package route

import (
	"net/url"

	gometrics "github.com/eBay/fabio/_third_party/github.com/rcrowley/go-metrics"
)

type Target struct {
	// service is the name of the service the targetURL points to
	service string

	// tags are the list of tags for this target
	tags []string

	// URL is the endpoint the service instance listens on
	URL *url.URL

	// fixedWeight is the weight assigned to this target.
	// If the value is 0 the targets weight is dynamic.
	fixedWeight float64

	// weight is the actual weight for this service in percent.
	weight float64

	// timer measures throughput and latency of this target
	Timer gometrics.Timer
}
