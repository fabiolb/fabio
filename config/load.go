package config

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/magiconair/properties"
)

func Load(filename string) (*Config, error) {
	if filename == "" {
		return fromProperties(properties.NewProperties())
	}

	p, err := properties.LoadFile(filename, properties.UTF8)
	if err != nil {
		return nil, err
	}
	return fromProperties(p)
}

func fromProperties(p *properties.Properties) (cfg *Config, err error) {
	cfg = &Config{}

	deprecate := func(key, msg string) {
		_, exists := p.Get(key)
		if exists {
			log.Print("[WARN] config: ", msg)
		}
	}

	cfg.Proxy = Proxy{
		MaxConn:               intVal(p, Default.Proxy.MaxConn, "proxy.maxconn"),
		Strategy:              stringVal(p, Default.Proxy.Strategy, "proxy.strategy"),
		Matcher:               stringVal(p, Default.Proxy.Matcher, "proxy.matcher"),
		ShutdownWait:          durationVal(p, Default.Proxy.ShutdownWait, "proxy.shutdownwait"),
		DialTimeout:           durationVal(p, Default.Proxy.DialTimeout, "proxy.dialtimeout"),
		ResponseHeaderTimeout: durationVal(p, Default.Proxy.ResponseHeaderTimeout, "proxy.timeout"),
		KeepAliveTimeout:      durationVal(p, Default.Proxy.KeepAliveTimeout, "proxy.timeout"),
		LocalIP:               stringVal(p, Default.Proxy.LocalIP, "proxy.localip"),
		ClientIPHeader:        stringVal(p, Default.Proxy.ClientIPHeader, "proxy.header.clientip"),
		TLSHeader:             stringVal(p, Default.Proxy.TLSHeader, "proxy.header.tls"),
		TLSHeaderValue:        stringVal(p, Default.Proxy.TLSHeaderValue, "proxy.header.tls.value"),
	}

	readTimeout := durationVal(p, time.Duration(0), "proxy.readtimeout")
	writeTimeout := durationVal(p, time.Duration(0), "proxy.writetimeout")

	cfg.Listen, err = parseListen(stringVal(p, Default.Listen[0].Addr, "proxy.addr"), readTimeout, writeTimeout)
	if err != nil {
		return nil, err
	}

	cfg.Metrics = parseMetrics(
		stringVal(p, Default.Metrics[0].Target, "metrics.target"),
		stringVal(p, Default.Metrics[0].Prefix, "metrics.prefix"),
		stringVal(p, Default.Metrics[0].Addr, "metrics.graphite.addr"),
		durationVal(p, Default.Metrics[0].Interval, "metrics.interval"),
	)

	cfg.Registry = Registry{
		Backend: stringVal(p, Default.Registry.Backend, "registry.backend"),
		File: File{
			Path: stringVal(p, Default.Registry.File.Path, "registry.file.path"),
		},
		Static: Static{
			Routes: stringVal(p, Default.Registry.Static.Routes, "registry.static.routes"),
		},
		Consul: Consul{
			Addr:          stringVal(p, Default.Registry.Consul.Addr, "registry.consul.addr", "consul.addr"),
			Token:         stringVal(p, Default.Registry.Consul.Token, "registry.consul.token", "consul.token"),
			KVPath:        stringVal(p, Default.Registry.Consul.KVPath, "registry.consul.kvpath", "consul.kvpath"),
			TagPrefix:     stringVal(p, Default.Registry.Consul.TagPrefix, "registry.consul.tagprefix", "consul.tagprefix"),
			ServiceAddr:   stringVal(p, Default.Registry.Consul.ServiceAddr, "registry.consul.register.addr"),
			ServiceName:   stringVal(p, Default.Registry.Consul.ServiceName, "registry.consul.register.name", "consul.register.name"),
			ServiceTags:   stringAVal(p, Default.Registry.Consul.ServiceTags, "registry.consul.register.tags"),
			CheckInterval: durationVal(p, Default.Registry.Consul.CheckInterval, "registry.consul.register.checkInterval", "consul.register.checkInterval"),
			CheckTimeout:  durationVal(p, Default.Registry.Consul.CheckTimeout, "registry.consul.register.checkTimeout", "consul.register.checkTimeout"),
		},
	}
	deprecate("consul.addr", "consul.addr has been replaced by registry.consul.addr")
	deprecate("consul.token", "consul.token has been replaced by registry.consul.token")
	deprecate("consul.kvpath", "consul.kvpath has been replaced by registry.consul.kvpath")
	deprecate("consul.tagprefix", "consul.tagprefix has been replaced by registry.consul.tagprefix")
	deprecate("consul.register.name", "consul.register.name has been replaced by registry.consul.register.name")
	deprecate("consul.register.checkInterval", "consul.register.checkInterval has been replaced by registry.consul.register.checkInterval")
	deprecate("consul.register.checkTimeout", "consul.register.checkTimeout has been replaced by registry.consul.register.checkTimeout")
	deprecate("consul.url", "consul.url is obsolete. Please remove it.")

	proxyRoutes := stringVal(p, "", "proxy.routes")
	if strings.HasPrefix(proxyRoutes, "@") {
		cfg.Registry.Backend = "file"
		cfg.Registry.File.Path = proxyRoutes[1:]
		deprecate("proxy.routes", "Please use registry.backend=file and registry.file.path=<path> instead of proxy.routes=@<path>")
	} else if proxyRoutes != "" {
		cfg.Registry.Backend = "static"
		cfg.Registry.Static.Routes = proxyRoutes
		deprecate("proxy.routes", "Please use registry.backend=static and registry.static.routes=<routes> instead of proxy.routes=<routes>")
	}

	cfg.Runtime = Runtime{
		GOGC:       intVal(p, Default.Runtime.GOGC, "runtime.gogc"),
		GOMAXPROCS: intVal(p, Default.Runtime.GOMAXPROCS, "runtime.gomaxprocs"),
	}
	if cfg.Runtime.GOMAXPROCS == -1 {
		cfg.Runtime.GOMAXPROCS = runtime.NumCPU()
	}

	cfg.UI = UI{
		Addr:  stringVal(p, Default.UI.Addr, "ui.addr"),
		Color: stringVal(p, Default.UI.Color, "ui.color"),
		Title: stringVal(p, Default.UI.Title, "ui.title"),
	}
	return cfg, nil
}

// stringVal returns the first non-empty value found or the default value.
// Keys are checked in order and environment variables take precedence over
// properties values.  Environment varaible names are derived from property
// names by replacing the dots with underscores.
func stringVal(p *properties.Properties, def string, keys ...string) string {
	for _, key := range keys {
		if v := os.Getenv(strings.Replace(key, ".", "_", -1)); v != "" {
			return v
		}
		if p == nil {
			continue
		}
		if v, ok := p.Get(key); ok {
			return v
		}
	}
	return def
}

func stringAVal(p *properties.Properties, def []string, keys ...string) []string {
	v := stringVal(p, "", keys...)
	if v == "" {
		return def
	}
	return splitSkipEmpty(v, ",")
}

func splitSkipEmpty(s, sep string) (vals []string) {
	for _, v := range strings.Split(s, sep) {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		vals = append(vals, v)
	}
	return vals
}

func intVal(p *properties.Properties, def int, keys ...string) int {
	v := stringVal(p, "", keys...)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("[WARN] Invalid value %s for %v", v, keys)
		return def
	}
	return n
}

func durationVal(p *properties.Properties, def time.Duration, keys ...string) time.Duration {
	v := stringVal(p, "", keys...)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Printf("[WARN] Invalid duration %s for %v", v, keys)
		return def
	}
	return d
}

func parseMetrics(target, prefix, graphiteAddr string, interval time.Duration) []Metrics {
	m := Metrics{Target: target, Prefix: prefix, Interval: interval}
	if target == "graphite" {
		m.Addr = graphiteAddr
	}
	return []Metrics{m}
}

func parseListen(addrs string, readTimeout, writeTimeout time.Duration) ([]Listen, error) {
	listen := []Listen{}
	for _, addr := range strings.Split(addrs, ",") {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}

		var l Listen
		p := strings.Split(addr, ";")
		switch len(p) {
		case 1:
			l.Addr = p[0]
		case 2:
			l.Addr, l.CertFile, l.KeyFile, l.TLS = p[0], p[1], p[1], true
		case 3:
			l.Addr, l.CertFile, l.KeyFile, l.TLS = p[0], p[1], p[2], true
		case 4:
			l.Addr, l.CertFile, l.KeyFile, l.ClientAuthFile, l.TLS = p[0], p[1], p[2], p[3], true
		default:
			return nil, fmt.Errorf("invalid address %s", addr)
		}
		l.ReadTimeout = readTimeout
		l.WriteTimeout = writeTimeout
		listen = append(listen, l)
	}
	return listen, nil
}
