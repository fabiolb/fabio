// Package metrics provides functions for collecting
// and managing metrics through different metrics libraries.
//
// Metrics library implementations must implement the
// Registry interface in the package.
package metrics

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/exit"
)

// DefaultRegistry stores the metrics library provider.
var DefaultRegistry Registry = NoopRegistry{}

var routeMetricNameTemplate string

// NewRegistry creates a new metrics registry.
func NewRegistry(cfg config.Metrics) (r Registry, err error) {
	prefix := cfg.Prefix
	if prefix == "default" {
		prefix = defaultPrefix()
	}

	routeMetricNameTemplate = cfg.RouteMetricNameTemplate

	switch cfg.Target {
	case "stdout":
		log.Printf("[INFO] Sending metrics to stdout")
		return gmStdoutRegistry(cfg.Interval)

	case "graphite":
		log.Printf("[INFO] Sending metrics to Graphite on %s as %q", cfg.GraphiteAddr, prefix)
		return gmGraphiteRegistry(prefix, cfg.GraphiteAddr, cfg.Interval)

	case "statsd":
		log.Printf("[INFO] Sending metrics to StatsD on %s as %q", cfg.StatsDAddr, prefix)
		return gmStatsDRegistry(prefix, cfg.StatsDAddr, cfg.Interval)

	case "circonus":
		return circonusRegistry(prefix,
			cfg.CirconusAPIKey,
			cfg.CirconusAPIApp,
			cfg.CirconusAPIURL,
			cfg.CirconusBrokerID,
			cfg.CirconusCheckID,
			cfg.Interval)

	default:
		exit.Fatal("[FATAL] Invalid metrics target ", cfg.Target)
	}
	panic("unreachable")
}

// TargetName returns the metrics name from the given parameters.
func TargetName(service, host, path string, targetURL *url.URL) string {
	name := routeMetricNameTemplate
	name = strings.Replace(name, "$service$", clean(service), -1)
	name = strings.Replace(name, "$host$", clean(host), -1)
	name = strings.Replace(name, "$path$", clean(path), -1)
	name = strings.Replace(name, "$target$", clean(targetURL.Host), -1)

	return name
}

// clean creates safe names for graphite reporting by replacing
// some characters with underscores.
// TODO(fs): This may need updating for other metrics backends.
func clean(s string) string {
	if s == "" {
		return "_"
	}
	s = strings.Replace(s, ".", "_", -1)
	s = strings.Replace(s, ":", "_", -1)
	return strings.ToLower(s)
}

// stubbed out for testing
var hostname = os.Hostname

// defaultPrefix determines the default metrics prefix from
// the current hostname and the name of the executable.
func defaultPrefix() string {
	host, err := hostname()
	if err != nil {
		exit.Fatal("[FATAL] ", err)
	}
	exe := filepath.Base(os.Args[0])
	return clean(host) + "." + clean(exe)
}
