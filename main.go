package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/eBay/fabio/admin"
	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/proxy"
	"github.com/eBay/fabio/registry"
	"github.com/eBay/fabio/route"
)

// version contains the version number
//
// It is set by build/release.sh for tagged releases
// so that 'go get' just works.
//
// It is also set by the linker when fabio
// is built via the Makefile or the build/docker.sh
// script to ensure the correct version nubmer
var version = "1.0.7-dev"

func main() {
	var filename string
	var v bool
	flag.StringVar(&filename, "cfg", "", "path to config file")
	flag.BoolVar(&v, "v", false, "show version")
	flag.Parse()

	if v {
		fmt.Println(version)
		return
	}
	log.Printf("[INFO] Version %s starting", version)

	cfg := config.DefaultConfig
	if filename != "" {
		cfg = loadConfig(filename)
	}

	initBackend(cfg)
	initMetrics(cfg)
	initRuntime(cfg)
	initRoutes(cfg)
	startAdmin(cfg)
	startListeners(cfg.Listen, cfg.Proxy.ShutdownWait, newProxy(cfg))
	registry.Default.Deregister()
}

func newProxy(cfg *config.Config) *proxy.Proxy {
	if err := route.SetPickerStrategy(cfg.Proxy.Strategy); err != nil {
		log.Fatal("[FATAL] ", err)
	}
	log.Printf("[INFO] Using routing strategy %q", cfg.Proxy.Strategy)

	tr := &http.Transport{
		ResponseHeaderTimeout: cfg.Proxy.ResponseHeaderTimeout,
		MaxIdleConnsPerHost:   cfg.Proxy.MaxConn,
		Dial: (&net.Dialer{
			Timeout:   cfg.Proxy.DialTimeout,
			KeepAlive: cfg.Proxy.KeepAliveTimeout,
		}).Dial,
	}

	return proxy.New(tr, cfg.Proxy)
}

func startAdmin(cfg *config.Config) {
	log.Printf("[INFO] Admin server listening on %q", cfg.UI.Addr)
	go func() {
		if err := admin.ListenAndServe(cfg.UI.Addr, version); err != nil {
			log.Fatal("[FATAL] ui: ", err)
		}
	}()
}
