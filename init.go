package main

import (
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/metrics"
	"github.com/eBay/fabio/registry"
	"github.com/eBay/fabio/registry/consul"
	"github.com/eBay/fabio/route"
)

var be registry.Backend

func loadConfig(filename string) *config.Config {
	cfg, err := config.FromFile(filename)
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}
	return cfg
}

func initBackend(cfg *config.Config) {
	var err error
	be, err = consul.NewBackend(&cfg.Consul)
	if err != nil {
		log.Fatal("[FATAL] Error initializing backend. ", err)
	}
}

func initMetrics(cfg *config.Config) {
	if err := metrics.Init(cfg.Metrics); err != nil {
		log.Fatal("[FATAL] ", err)
	}
}

func initRuntime(cfg *config.Config) {
	if os.Getenv("GOGC") == "" {
		log.Print("[INFO] Setting GOGC=", cfg.Runtime.GOGC)
		debug.SetGCPercent(cfg.Runtime.GOGC)
	} else {
		log.Print("[INFO] Using GOGC=", os.Getenv("GOGC"), " from env")
	}

	if os.Getenv("GOMAXPROCS") == "" {
		log.Print("[INFO] Setting GOMAXPROCS=", cfg.Runtime.GOMAXPROCS)
		runtime.GOMAXPROCS(cfg.Runtime.GOMAXPROCS)
	} else {
		log.Print("[INFO] Using GOMAXPROCS=", os.Getenv("GOMAXPROCS"), " from env")
	}
}

func initRoutes(cfg *config.Config) {
	if cfg.Routes == "" {
		initDynamicRoutes()
	} else {
		initStaticRoutes(cfg.Routes)
	}
}

func initDynamicRoutes() {
	go func() {
		ch := be.Watch()
		for {
			r := <-ch
			t, err := route.ParseString(r)
			if err != nil {
				log.Printf("[WARN] %s", err)
				continue
			}
			route.SetTable(t)
		}
	}()
}

func initStaticRoutes(routes string) {
	var err error
	var t route.Table

	if strings.HasPrefix(routes, "@") {
		routes = routes[1:]
		log.Print("[INFO] Using static routes from ", routes)
		t, err = route.ParseFile(routes)
	} else {
		log.Print("[INFO] Using static routes from config file")
		t, err = route.ParseString(routes)
	}

	if err != nil {
		log.Fatal("[FATAL] ", err)
	}

	route.SetTable(t)
}
