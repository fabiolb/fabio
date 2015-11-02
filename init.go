package main

import (
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/consul"
	"github.com/eBay/fabio/metrics"
	"github.com/eBay/fabio/route"
)

func loadConfig(filename string) *config.Config {
	cfg, err := config.FromFile(filename)
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}
	return cfg
}

func initConsul(cfg *config.Config) {
	consul.Addr = cfg.Consul.Addr
	consul.URL = cfg.Consul.URL

	dc, err := consul.Datacenter()
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}

	log.Printf("[INFO] Connecting to consul on %q in datacenter %q", cfg.Consul.Addr, dc)
	log.Printf("[INFO] Consul can be reached via %q", cfg.Consul.URL)
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
		initDynamicRoutes(cfg.Consul)
	} else {
		initStaticRoutes(cfg.Routes)
	}
}

func initDynamicRoutes(cfg config.Consul) {
	log.Printf("[INFO] Using dynamic routes from consul on %s", cfg.Addr)
	log.Printf("[INFO] Using tag prefix %q", cfg.TagPrefix)
	log.Printf("[INFO] Watching KV path %q", cfg.KVPath)
	go func() {
		w, err := consul.NewWatcher(cfg.TagPrefix, cfg.KVPath)
		if err != nil {
			log.Fatal("[FATAL] ", err)
		}
		w.Watch()
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
