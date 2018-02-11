// Package metrics provides functions for collecting
// and managing metrics through different metrics libraries.
//
// Metrics library implementations must implement the
// Registry interface in the package.
package metrics

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/exit"
)

// DefaultRegistry stores the metrics library provider.
var DefaultRegistry Registry = NoopRegistry{}

// DefaultNames contains the default template for route metric names.
const DefaultNames = "{{clean .Service}}.{{clean .Host}}.{{clean .Path}}.{{clean .TargetURL.Host}}"

// DefaulPrefix contains the default template for metrics prefix.
const DefaultPrefix = "{{clean .Hostname}}.{{clean .Exec}}"

// names stores the template for the route metric names.
var names *template.Template

// prefix stores the final prefix string to use it with metric collectors where applicable, i.e. Graphite/StatsD
var prefix string

func init() {
	// make sure names is initialized to something
	var err error
	if names, err = parseNames(DefaultNames); err != nil {
		panic(err)
	}
}

// NewRegistry creates a new metrics registry.
func NewRegistry(cfg config.Metrics) (r Registry, err error) {

	if prefix, err = parsePrefix(cfg.Prefix); err != nil {
		return nil, fmt.Errorf("metrics: invalid Prefix template. %s", err)
	}

	if names, err = parseNames(cfg.Names); err != nil {
		return nil, fmt.Errorf("metrics: invalid names template. %s", err)
	}

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
		return circonusRegistry(prefix, cfg.Circonus, cfg.Interval)

	default:
		exit.Fatal("[FATAL] Invalid metrics target ", cfg.Target)
	}
	panic("unreachable")
}

// parsePrefix parses the prefix metric template
func parsePrefix(tmpl string) (string, error) {
	// Backward compatibility condition for old metrics.prefix parameter 'default'
	if tmpl == "default" {
		tmpl = DefaultPrefix
	}
	funcMap := template.FuncMap{
		"clean": clean,
	}
	t, err := template.New("prefix").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}
	host, err := hostname()
	if err != nil {
		return "", err
	}
	exe := filepath.Base(os.Args[0])

	b := new(bytes.Buffer)
	data := struct{ Hostname, Exec string }{host, exe}
	if err := t.Execute(b, &data); err != nil {
		return "", err
	}
	return b.String(), nil
}

// parseNames parses the route metric name template.
func parseNames(tmpl string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"clean": clean,
	}
	t, err := template.New("names").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return nil, err
	}
	testURL, err := url.Parse("http://127.0.0.1:12345/")
	if err != nil {
		return nil, err
	}
	if _, err := TargetName("testservice", "test.example.com", "/test", testURL); err != nil {
		return nil, err
	}
	return t, nil
}

// TargetName returns the metrics name from the given parameters.
func TargetName(service, host, path string, targetURL *url.URL) (string, error) {
	if names == nil {
		return "", nil
	}

	var name bytes.Buffer

	data := struct {
		Service, Host, Path string
		TargetURL           *url.URL
	}{service, host, path, targetURL}

	if err := names.Execute(&name, data); err != nil {
		return "", err
	}

	return name.String(), nil
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
