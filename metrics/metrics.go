package metrics

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eBay/fabio/_third_party/github.com/cyberdelia/go-metrics-graphite"
	gometrics "github.com/eBay/fabio/_third_party/github.com/rcrowley/go-metrics"
)

var pfx string

func Init(target, prefix string, interval time.Duration, graphteAddr string) error {
	pfx = prefix
	if pfx == "default" {
		pfx = defaultPrefix()
	}

	switch target {
	case "stdout":
		log.Printf("[INFO] Sending metrics to stdout")
		return initStdout(interval)
	case "graphite":
		if graphteAddr == "" {
			return errors.New("metrics: graphite addr missing")
		}

		log.Printf("[INFO] Sending metrics to Graphite on %s as %q", graphteAddr, pfx)
		return initGraphite(graphteAddr, interval)
	case "":
		log.Printf("[INFO] Metrics disabled")
	default:
		log.Fatal("[FATAL] Invalid metrics target ", target)
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

func initStdout(interval time.Duration) error {
	logger := log.New(os.Stderr, "localhost: ", log.Lmicroseconds)
	go gometrics.Log(gometrics.DefaultRegistry, interval, logger)
	return nil
}

func initGraphite(addr string, interval time.Duration) error {
	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return fmt.Errorf("metrics: cannot connect to Graphite: %s", err)
	}

	go graphite.Graphite(gometrics.DefaultRegistry, interval, pfx, a)
	return nil
}
