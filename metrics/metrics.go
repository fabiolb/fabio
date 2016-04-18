package metrics

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/cyberdelia/go-metrics-graphite"
	"github.com/eBay/fabio/config"
	gometrics "github.com/rcrowley/go-metrics"
)

var pfx string

type Prefix struct {
	Fqdn     string
	Hostname string
	Exe      string
}

// ServiceRegistry contains a separate metrics registry for
// the timers for all targets to avoid conflicts
// with globally registered timers.
var ServiceRegistry = gometrics.NewRegistry()

func Init(cfgs []config.Metrics) error {
	for _, cfg := range cfgs {
		if err := initMetrics(cfg); err != nil {
			return err
		}
	}
	return nil
}

func initMetrics(cfg config.Metrics) error {
	pfx = cfg.Prefix
	prefix := initPrefixTemplate()
	if strings.Contains(pfx, "default") {
		pfx = strings.Replace(pfx, "default", defaultPrefix(), 1)
	} else {

		t := template.New("Prefix template")
		t, err := t.Parse(pfx)
		if err != nil {
			fmt.Println("Fatal error ", err.Error())
			os.Exit(1)
		}
		var doc bytes.Buffer
		t.Execute(&doc, prefix)
		pfx = doc.String()
	}

	switch cfg.Target {
	case "stdout":
		log.Printf("[INFO] Sending metrics to stdout")
		return initStdout(cfg.Interval)
	case "graphite":
		if cfg.Addr == "" {
			return errors.New("metrics: graphite addr missing")
		}

		log.Printf("[INFO] Sending metrics to Graphite on %s as %q", cfg.Addr, pfx)
		return initGraphite(cfg.Addr, cfg.Interval)
	case "":
		log.Printf("[INFO] Metrics disabled")
	default:
		log.Fatal("[FATAL] Invalid metrics target ", cfg.Target)
	}
	return nil
}

func TargetName(service, host, path string, targetURL *url.URL) string {
	return strings.Join([]string{
		clean(service),
		clean(host),
		clean(path),
		clean(targetURL.Host),
	}, ".")
}

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

func defaultPrefix() string {
	host, err := hostname()
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}
	exe := filepath.Base(os.Args[0])
	return clean(host) + "." + clean(exe)
}

func initPrefixTemplate() *Prefix {
	// Init Template variables
	var fqdn string
	hostname, err := hostname()
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}
	out, err := exec.Command("hostname", "-f").Output()
	if err != nil {
		log.Printf("[WARN] Couldn't determine hostname fqdn, failing back to os.Hostname()", err)
		fqdn = hostname
	} else {
		fqdn = strings.Trim(string(out), "\n\r\t")
	}
	exe := filepath.Base(os.Args[0])

	// Assign template variables to Prefix structure
	prefix := Prefix{
		Fqdn:     clean(fqdn),
		Hostname: clean(hostname),
		Exe:      clean(exe),
	}
	return &prefix
}

func initStdout(interval time.Duration) error {
	logger := log.New(os.Stderr, "localhost: ", log.Lmicroseconds)
	go gometrics.Log(gometrics.DefaultRegistry, interval, logger)
	go gometrics.Log(ServiceRegistry, interval, logger)
	return nil
}

func initGraphite(addr string, interval time.Duration) error {
	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return fmt.Errorf("metrics: cannot connect to Graphite: %s", err)
	}

	go graphite.Graphite(gometrics.DefaultRegistry, interval, pfx, a)
	go graphite.Graphite(ServiceRegistry, interval, pfx, a)
	return nil
}
