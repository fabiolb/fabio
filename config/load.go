package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/magiconair/properties"
)

func Load() (cfg *Config, err error) {
	var path string
	for i, arg := range os.Args {
		if arg == "-v" {
			return nil, nil
		}
		if arg == "-cfg" {
			if i == len(os.Args)-1 {
				return nil, errors.New("missing path or url to config file")
			}
			path = os.Args[i+1]
			break
		}
	}
	p, err := loadProperties(path)
	if err != nil {
		return nil, err
	}
	return load(p)
}

func loadProperties(path string) (p *properties.Properties, err error) {
	if path == "" {
		return properties.NewProperties(), nil
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return properties.LoadURL(path)
	}
	return properties.LoadFile(path, properties.UTF8)
}

func load(p *properties.Properties) (cfg *Config, err error) {
	cfg = &Config{}

	f := NewFlagSet(os.Args[0], flag.ExitOnError)

	// dummy values which were parsed earlier
	f.String("cfg", "", "Path or URL to config file")
	f.Bool("v", false, "Show version")

	// config values
	f.IntVar(&cfg.Proxy.MaxConn, "proxy.maxconn", Default.Proxy.MaxConn, "maximum number of cached connections")
	f.StringVar(&cfg.Proxy.Strategy, "proxy.strategy", Default.Proxy.Strategy, "load balancing strategy")
	f.StringVar(&cfg.Proxy.Matcher, "proxy.matcher", Default.Proxy.Matcher, "path matching algorithm")
	f.IntVar(&cfg.Proxy.NoRouteStatus, "proxy.noroutestatus", Default.Proxy.NoRouteStatus, "status code for invalid route")
	f.DurationVar(&cfg.Proxy.ShutdownWait, "proxy.shutdownwait", Default.Proxy.ShutdownWait, "time for graceful shutdown")
	f.DurationVar(&cfg.Proxy.DialTimeout, "proxy.dialtimeout", Default.Proxy.DialTimeout, "connection timeout for backend connections")
	f.DurationVar(&cfg.Proxy.ResponseHeaderTimeout, "proxy.responseheadertimeout", Default.Proxy.ResponseHeaderTimeout, "response header timeout")
	f.DurationVar(&cfg.Proxy.KeepAliveTimeout, "proxy.keepalivetimeout", Default.Proxy.KeepAliveTimeout, "keep-alive timeout")
	f.StringVar(&cfg.Proxy.LocalIP, "proxy.localip", Default.Proxy.LocalIP, "fabio address in Forward headers")
	f.StringVar(&cfg.Proxy.ClientIPHeader, "proxy.header.clientip", Default.Proxy.ClientIPHeader, "header for the request ip")
	f.StringVar(&cfg.Proxy.TLSHeader, "proxy.header.tls", Default.Proxy.TLSHeader, "header for TLS connections")
	f.StringVar(&cfg.Proxy.TLSHeaderValue, "proxy.header.tls.value", Default.Proxy.TLSHeaderValue, "value for TLS connection header")
	f.StringVar(&cfg.Proxy.ListenerAddr, "proxy.addr", Default.Proxy.ListenerAddr, "listener config")
	f.DurationVar(&cfg.Proxy.ReadTimeout, "proxy.readtimeout", Default.Proxy.ReadTimeout, "read timeout for incoming requests")
	f.DurationVar(&cfg.Proxy.WriteTimeout, "proxy.writetimeout", Default.Proxy.WriteTimeout, "write timeout for outgoing responses")
	f.StringVar(&cfg.Metrics.Target, "metrics.target", Default.Metrics.Target, "metrics backend")
	f.StringVar(&cfg.Metrics.Prefix, "metrics.prefix", Default.Metrics.Prefix, "prefix for reported metrics")
	f.DurationVar(&cfg.Metrics.Interval, "metrics.interval", Default.Metrics.Interval, "metrics reporting interval")
	f.StringVar(&cfg.Metrics.GraphiteAddr, "metrics.graphite.addr", Default.Metrics.GraphiteAddr, "graphite server address")
	f.StringVar(&cfg.Registry.Backend, "registry.backend", Default.Registry.Backend, "registry backend")
	f.StringVar(&cfg.Registry.File.Path, "registry.file.path", Default.Registry.File.Path, "path to file based routing table")
	f.StringVar(&cfg.Registry.Static.Routes, "registry.static.routes", Default.Registry.Static.Routes, "static routes")
	f.StringVar(&cfg.Registry.Consul.Addr, "registry.consul.addr", Default.Registry.Consul.Addr, "address of the consul agent")
	f.StringVar(&cfg.Registry.Consul.Token, "registry.consul.token", Default.Registry.Consul.Token, "token for consul agent")
	f.StringVar(&cfg.Registry.Consul.KVPath, "registry.consul.kvpath", Default.Registry.Consul.KVPath, "consul KV path for manual overrides")
	f.StringVar(&cfg.Registry.Consul.TagPrefix, "registry.consul.tagprefix", Default.Registry.Consul.TagPrefix, "prefix for consul tags")
	f.BoolVar(&cfg.Registry.Consul.Register, "registry.consul.register.enabled", Default.Registry.Consul.Register, "register fabio in consul")
	f.StringVar(&cfg.Registry.Consul.ServiceAddr, "registry.consul.register.addr", Default.Registry.Consul.ServiceAddr, "service registration address")
	f.StringVar(&cfg.Registry.Consul.ServiceName, "registry.consul.register.name", Default.Registry.Consul.ServiceName, "service registration name")
	f.Var((*tags)(&cfg.Registry.Consul.ServiceTags), "registry.consul.register.tags", "service registration tags")
	f.DurationVar(&cfg.Registry.Consul.CheckInterval, "registry.consul.register.checkInterval", Default.Registry.Consul.CheckInterval, "service check interval")
	f.DurationVar(&cfg.Registry.Consul.CheckTimeout, "registry.consul.register.checkTimeout", Default.Registry.Consul.CheckTimeout, "service check timeout")
	f.IntVar(&cfg.Runtime.GOGC, "runtime.gogc", Default.Runtime.GOGC, "sets runtime.GOGC")
	f.IntVar(&cfg.Runtime.GOMAXPROCS, "runtime.gomaxprocs", Default.Runtime.GOMAXPROCS, "sets runtime.GOMAXPROCS")
	f.StringVar(&cfg.UI.Addr, "ui.addr", Default.UI.Addr, "address the UI/API is listening on")
	f.StringVar(&cfg.UI.Color, "ui.color", Default.UI.Color, "background color of the UI")
	f.StringVar(&cfg.UI.Title, "ui.title", Default.UI.Title, "optional title for the UI")

	var proxyTimeout time.Duration
	f.DurationVar(&proxyTimeout, "proxy.timeout", time.Duration(0), "deprecated")

	// parse configuration
	prefixes := []string{"FABIO_", ""}
	if err := f.ParseFlags(os.Args[1:], os.Environ(), prefixes, p); err != nil {
		return nil, err
	}

	// post configuration
	if cfg.Runtime.GOMAXPROCS == -1 {
		cfg.Runtime.GOMAXPROCS = runtime.NumCPU()
	}

	cfg.Registry.Consul.Scheme, cfg.Registry.Consul.Addr = parseScheme(cfg.Registry.Consul.Addr)

	cfg.Listen, err = parseListen(cfg.Proxy.ListenerAddr, cfg.Proxy.ReadTimeout, cfg.Proxy.WriteTimeout)
	if err != nil {
		return nil, err
	}

	// handle deprecations
	deprecate := func(name, msg string) {
		if f.IsSet(name) {
			log.Print("[WARN] ", msg)
		}
	}
	deprecate("proxy.timeout", "proxy.timeout has been replaced by proxy.responseheadertimeout and proxy.keepalivetimeout")
	deprecate("consul.addr", "consul.addr has been replaced by registry.consul.addr")
	deprecate("consul.token", "consul.token has been replaced by registry.consul.token")
	deprecate("consul.kvpath", "consul.kvpath has been replaced by registry.consul.kvpath")
	deprecate("consul.tagprefix", "consul.tagprefix has been replaced by registry.consul.tagprefix")
	deprecate("consul.register.name", "consul.register.name has been replaced by registry.consul.register.name")
	deprecate("consul.register.checkInterval", "consul.register.checkInterval has been replaced by registry.consul.register.checkInterval")
	deprecate("consul.register.checkTimeout", "consul.register.checkTimeout has been replaced by registry.consul.register.checkTimeout")
	deprecate("consul.url", "consul.url is obsolete. Please remove it.")

	if proxyTimeout > 0 {
		if cfg.Proxy.ResponseHeaderTimeout == 0 {
			cfg.Proxy.ResponseHeaderTimeout = proxyTimeout
		}
		if cfg.Proxy.KeepAliveTimeout == 0 {
			cfg.Proxy.KeepAliveTimeout = proxyTimeout
		}
	}

	return cfg, nil
}

func parseScheme(s string) (scheme, addr string) {
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "https://") {
		return "https", s[len("https://"):]
	}
	if strings.HasPrefix(s, "http://") {
		return "http", s[len("http://"):]
	}
	return "http", s
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

type tags []string

func (t *tags) String() string {
	return strings.Join(*t, ",")
}

func (t *tags) Set(value string) error {
	for _, v := range splitSkipEmpty(value, ",") {
		*t = append(*t, v)
	}
	return nil
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
