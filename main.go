package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/route"
	"github.com/eBay/fabio/ui"
)

var version = "1.0.4"

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
	initConsul(cfg)
	initMetrics(cfg)
	initRuntime(cfg)
	initRoutes(cfg)
	startUI(cfg)
	startListeners(cfg.Listen, cfg.Proxy.ShutdownWait, newProxy(cfg))
}

func newProxy(cfg *config.Config) *route.Proxy {
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

	return route.NewProxy(tr, cfg.Proxy)
}

func startUI(cfg *config.Config) {
	log.Printf("[INFO] UI listening on %q", cfg.UI.Addr)
	go func() {
		if err := ui.Start(cfg.UI.Addr, cfg.Consul.KVPath); err != nil {
			log.Fatal("[FATAL] ui: ", err)
		}
	}()
}
