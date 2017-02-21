package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"io"
	"github.com/Graylog2/go-gelf/gelf"
	"github.com/eBay/fabio/admin"
	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/exit"
	"github.com/eBay/fabio/metrics"
	"github.com/eBay/fabio/proxy"
	"github.com/eBay/fabio/proxy/tcp"
	"github.com/eBay/fabio/registry"
	"github.com/eBay/fabio/registry/consul"
	"github.com/eBay/fabio/registry/file"
	"github.com/eBay/fabio/registry/static"
	"github.com/eBay/fabio/route"
	dmp "github.com/sergi/go-diff/diffmatchpatch"
)

// version contains the version number
//
// It is set by build/release.sh for tagged releases
// so that 'go get' just works.
//
// It is also set by the linker when fabio
// is built via the Makefile or the build/docker.sh
// script to ensure the correct version nubmer
var version = "1.4beta1"

var shuttingDown int32

func main() {
	cfg, err := config.Load(os.Args, os.Environ())
	if err != nil {
		exit.Fatalf("[FATAL] %s. %s", version, err)
	}
	if cfg == nil {
		fmt.Println(version)
		return
	}

	log.Printf("[INFO] Runtime config\n" + toJSON(cfg))
	log.Printf("[INFO] Version %s starting", version)
	log.Printf("[INFO] Go runtime is %s", runtime.Version())

	exit.Listen(func(s os.Signal) {
		atomic.StoreInt32(&shuttingDown, 1)
		proxy.Shutdown(cfg.Proxy.ShutdownWait)
		if registry.Default == nil {
			return
		}
		registry.Default.Deregister()
	})

	// init metrics early since that create the global metric registries
	// that are used by other parts of the code.
	initMetrics(cfg)
	logExternal(cfg)
	initRuntime(cfg)
	initBackend(cfg)
	startAdmin(cfg)

	first := make(chan bool)
	go watchBackend(cfg, first)
	log.Print("[INFO] Waiting for first routing table")
	<-first

	// create proxies after metrics since they use the metrics registry.
	startServers(cfg)
	exit.Wait()
	log.Print("[INFO] Down")
}

func newHTTPProxy(cfg *config.Config) http.Handler {
	pick := route.Picker[cfg.Proxy.Strategy]
	match := route.Matcher[cfg.Proxy.Matcher]
	notFound := metrics.DefaultRegistry.GetCounter("notfound")
	log.Printf("[INFO] Using routing strategy %q", cfg.Proxy.Strategy)
	log.Printf("[INFO] Using route matching %q", cfg.Proxy.Matcher)

	return &proxy.HTTPProxy{
		Config: cfg.Proxy,
		Transport: &http.Transport{
			ResponseHeaderTimeout: cfg.Proxy.ResponseHeaderTimeout,
			MaxIdleConnsPerHost:   cfg.Proxy.MaxConn,
			Dial: (&net.Dialer{
				Timeout:   cfg.Proxy.DialTimeout,
				KeepAlive: cfg.Proxy.KeepAliveTimeout,
			}).Dial,
		},
		Lookup: func(r *http.Request) *route.Target {
			t := route.GetTable().Lookup(r, r.Header.Get("trace"), pick, match)
			if t == nil {
				notFound.Inc(1)
				log.Print("[WARN] No route for ", r.Host, r.URL)
			}
			return t
		},
		Requests: metrics.DefaultRegistry.GetTimer("requests"),
	}
}

func lookupHostFn(cfg *config.Config) func(string) string {
	pick := route.Picker[cfg.Proxy.Strategy]
	notFound := metrics.DefaultRegistry.GetCounter("notfound")
	return func(host string) string {
		t := route.GetTable().LookupHost(host, pick)
		if t == nil {
			notFound.Inc(1)
			log.Print("[WARN] No route for ", host)
			return ""
		}
		return t.URL.Host
	}
}

func startAdmin(cfg *config.Config) {
	log.Printf("[INFO] Admin server listening on %q", cfg.UI.Addr)
	go func() {
		srv := &admin.Server{
			Color:    cfg.UI.Color,
			Title:    cfg.UI.Title,
			Version:  version,
			Commands: route.Commands,
			Cfg:      cfg,
		}
		if err := srv.ListenAndServe(cfg.UI.Addr); err != nil {
			exit.Fatal("[FATAL] ui: ", err)
		}
	}()
}

func startServers(cfg *config.Config) {
	for _, l := range cfg.Listen {
		switch l.Proto {
		case "http", "https":
			h := newHTTPProxy(cfg)
			go proxy.ListenAndServeHTTP(l, h)
		case "tcp":
			h := &tcp.Proxy{cfg.Proxy.DialTimeout, lookupHostFn(cfg)}
			go proxy.ListenAndServeTCP(l, h)
		case "tcp+sni":
			h := &tcp.SNIProxy{cfg.Proxy.DialTimeout, lookupHostFn(cfg)}
			go proxy.ListenAndServeTCP(l, h)
		default:
			exit.Fatal("[FATAL] Invalid protocol ", l.Proto)
		}
	}
}

func initMetrics(cfg *config.Config) {
	if cfg.Metrics.Target == "" {
		log.Printf("[INFO] Metrics disabled")
		return
	}

	var err error
	if metrics.DefaultRegistry, err = metrics.NewRegistry(cfg.Metrics); err != nil {
		exit.Fatal("[FATAL] ", err)
	}
	if route.ServiceRegistry, err = metrics.NewRegistry(cfg.Metrics); err != nil {
		exit.Fatal("[FATAL] ", err)
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

func initBackend(cfg *config.Config) {
	var deadline = time.Now().Add(cfg.Registry.Timeout)

	var err error
	for {
		switch cfg.Registry.Backend {
		case "file":
			registry.Default, err = file.NewBackend(cfg.Registry.File.Path)
		case "static":
			registry.Default, err = static.NewBackend(cfg.Registry.Static.Routes)
		case "consul":
			registry.Default, err = consul.NewBackend(&cfg.Registry.Consul)
		default:
			exit.Fatal("[FATAL] Unknown registry backend ", cfg.Registry.Backend)
		}

		if err == nil {
			if err = registry.Default.Register(); err == nil {
				return
			}
		}
		log.Print("[WARN] Error initializing backend. ", err)

		if time.Now().After(deadline) {
			exit.Fatal("[FATAL] Timeout registering backend.")
		}

		time.Sleep(cfg.Registry.Retry)
		if atomic.LoadInt32(&shuttingDown) > 0 {
			exit.Exit(1)
		}
	}
}

func watchBackend(cfg *config.Config, first chan bool) {
	var (
		last   string
		svccfg string
		mancfg string

		once sync.Once
	)

	svc := registry.Default.WatchServices()
	man := registry.Default.WatchManual()

	for {
		select {
		case svccfg = <-svc:
		case mancfg = <-man:
		}

		// manual config overrides service config
		// order matters
		next := svccfg + "\n" + mancfg
		if next == last {
			continue
		}

		t, err := route.NewTable(next)
		if err != nil {
			log.Printf("[WARN] %s", err)
			continue
		}
		route.SetTable(t)
		logRoutes(last, next, cfg.Proxy.LogRoutes)
		last = next

		once.Do(func() { close(first) })
	}
}

func logRoutes(last, next, format string) {
	fmtDiff := func(diffs []dmp.Diff) string {
		var b bytes.Buffer
		for _, d := range diffs {
			t := strings.TrimSpace(d.Text)
			if t == "" {
				continue
			}
			switch d.Type {
			case dmp.DiffDelete:
				b.WriteString("- ")
				b.WriteString(strings.Replace(t, "\n", "\n- ", -1))
			case dmp.DiffInsert:
				b.WriteString("+ ")
				b.WriteString(strings.Replace(t, "\n", "\n+ ", -1))
			}
		}
		return b.String()
	}

	const defFormat = "delta"
	switch format {
	case "delta":
		if delta := fmtDiff(dmp.New().DiffMain(last, next, true)); delta != "" {
			log.Printf("[INFO] Config updates\n%s", delta)
		}

	case "all":
		log.Printf("[INFO] Updated config to\n%s", next)

	default:
		log.Printf("[WARN] Invalid route format %q. Defaulting to %q", format, defFormat)
		logRoutes(last, next, defFormat)
	}
}

func toJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		panic("json: " + err.Error())
	}
	return string(data)
}

func logExternal(cfg *config.Config) {
   if cfg.LogServer.Enabled {
     if cfg.LogServer.Protocol == "gelf" {
	graylogAddr := cfg.LogServer.Address + ":" + cfg.LogServer.Port
        gelfWriter, err := gelf.NewWriter(graylogAddr)
        if err != nil {
            log.Fatalf("gelf.NewWriter: %s", err)
        }
        // log to both stderr and graylog2
        log.SetOutput(io.MultiWriter(os.Stderr, gelfWriter))
        log.Printf("logging to stderr & graylog2@'%s'", graylogAddr)
     }
  }
}
