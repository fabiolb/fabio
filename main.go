package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/eBay/fabio/consul"
	"github.com/eBay/fabio/metrics"
	"github.com/eBay/fabio/route"
	"github.com/eBay/fabio/ui"
)

var version = "1.0.3"

func main() {
	var cfg string
	var v bool
	flag.StringVar(&cfg, "cfg", "", "path to config file")
	flag.BoolVar(&v, "v", false, "show version")
	flag.Parse()

	if v {
		fmt.Println(version)
		return
	}

	log.Printf("[INFO] Version %s starting", version)

	if cfg != "" {
		if err := loadConfig(cfg); err != nil {
			log.Fatal("[FATAL] ", err)
		}
	}

	if err := metrics.Init(metricsTarget, metricsPrefix, metricsInterval, metricsGraphiteAddr); err != nil {
		log.Fatal("[FATAL] ", err)
	}

	if os.Getenv("GOMAXPROCS") == "" {
		log.Print("[INFO] Setting GOMAXPROCS=", gomaxprocs)
		runtime.GOMAXPROCS(gomaxprocs)
	} else {
		log.Print("[INFO] Using GOMAXPROCS=", os.Getenv("GOMAXPROCS"), " from env")
	}

	if os.Getenv("GOGC") == "" {
		log.Print("[INFO] Setting GOGC=", gogc)
		debug.SetGCPercent(gogc)
	} else {
		log.Print("[INFO] Using GOGC=", os.Getenv("GOGC"), " from env")
	}

	if proxyRoutes == "" {
		useDynamicRoutes()
	} else {
		useStaticRoutes()
	}

	if err := route.SetPickerStrategy(proxyStrategy); err != nil {
		log.Fatal("[FATAL] ", err)
	}

	consul.Addr = consulAddr
	consul.URL = consulURL

	dc, err := consul.Datacenter()
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}

	log.Printf("[INFO] Using routing strategy %q", proxyStrategy)
	log.Printf("[INFO] Connecting to consul on %q in datacenter %q", consulAddr, dc)
	log.Printf("[INFO] Consul can be reached via %q", consulURL)

	log.Printf("[INFO] UI listening on %q", uiAddr)
	go func() {
		if err := ui.Start(uiAddr, consulKVPath); err != nil {
			log.Fatal("[FATAL] ui: ", err)
		}
	}()

	tr := &http.Transport{
		ResponseHeaderTimeout: proxyTimeout,
		MaxIdleConnsPerHost:   proxyMaxConn,
		Dial: (&net.Dialer{
			Timeout:   proxyDialTimeout,
			KeepAlive: proxyTimeout,
		}).Dial,
	}

	proxy := route.NewProxy(tr, proxyHeaderClientIP, proxyHeaderTLS, proxyHeaderTLSValue)
	listen(proxyAddr, proxyShutdownWait, proxy)
}

func useDynamicRoutes() {
	log.Printf("[INFO] Using dynamic routes from consul on %s", consulAddr)
	log.Printf("[INFO] Using tag prefix %q", consulTagPrefix)
	log.Printf("[INFO] Watching KV path %q", consulKVPath)
	go func() {
		w, err := consul.NewWatcher(consulTagPrefix, consulKVPath)
		if err != nil {
			log.Fatal("[FATAL] ", err)
		}
		w.Watch()
	}()
}

func useStaticRoutes() {
	var err error
	var t route.Table

	if strings.HasPrefix(proxyRoutes, "@") {
		proxyRoutes = proxyRoutes[1:]
		log.Print("[INFO] Using static routes from ", proxyRoutes)
		t, err = route.ParseFile(proxyRoutes)
	} else {
		log.Print("[INFO] Using static routes from config file")
		t, err = route.ParseString(proxyRoutes)
	}

	if err != nil {
		log.Fatal("[FATAL] ", err)
	}

	route.SetTable(t)
}
