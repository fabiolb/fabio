package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
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
		path, err = parseCfg(os.Args, i)
		if err != nil {
			return nil, err
		}
		if path != "" {
			break
		}
	}
	p, err := loadProperties(path)
	if err != nil {
		return nil, err
	}
	return load(p)
}

var errInvalidConfig = errors.New("invalid or missing path to config file")

func parseCfg(args []string, i int) (path string, err error) {
	if len(args) == 0 || i >= len(args) || !strings.HasPrefix(args[i], "-cfg") {
		return "", nil
	}
	arg := args[i]
	if arg == "-cfg" {
		if i >= len(args)-1 {
			return "", errInvalidConfig
		}
		return args[i+1], nil
	}

	if !strings.HasPrefix(arg, "-cfg=") {
		return "", errInvalidConfig
	}

	path = arg[len("-cfg="):]
	switch {
	case path == "":
		return "", errInvalidConfig
	case path[0] == '\'':
		path = strings.Trim(path, "'")
	case path[0] == '"':
		path = strings.Trim(path, "\"")
	}
	if path == "" {
		return "", errInvalidConfig
	}
	return path, nil
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
	f.StringSliceVar(&cfg.ListenerValue, "proxy.addr", Default.ListenerValue, "listener config")
	f.KVSliceVar(&cfg.CertSourcesValue, "proxy.cs", Default.CertSourcesValue, "certificate sources")
	f.DurationVar(&cfg.Proxy.ReadTimeout, "proxy.readtimeout", Default.Proxy.ReadTimeout, "read timeout for incoming requests")
	f.DurationVar(&cfg.Proxy.WriteTimeout, "proxy.writetimeout", Default.Proxy.WriteTimeout, "write timeout for outgoing responses")
	f.DurationVar(&cfg.Proxy.FlushInterval, "proxy.flushinterval", Default.Proxy.FlushInterval, "flush interval for streaming responses")
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
	f.StringSliceVar(&cfg.Registry.Consul.ServiceTags, "registry.consul.register.tags", Default.Registry.Consul.ServiceTags, "service registration tags")
	f.StringSliceVar(&cfg.Registry.Consul.ServiceStatus, "registry.consul.service.status", Default.Registry.Consul.ServiceStatus, "valid service status values")
	f.DurationVar(&cfg.Registry.Consul.CheckInterval, "registry.consul.register.checkInterval", Default.Registry.Consul.CheckInterval, "service check interval")
	f.DurationVar(&cfg.Registry.Consul.CheckTimeout, "registry.consul.register.checkTimeout", Default.Registry.Consul.CheckTimeout, "service check timeout")
	f.IntVar(&cfg.Runtime.GOGC, "runtime.gogc", Default.Runtime.GOGC, "sets runtime.GOGC")
	f.IntVar(&cfg.Runtime.GOMAXPROCS, "runtime.gomaxprocs", Default.Runtime.GOMAXPROCS, "sets runtime.GOMAXPROCS")
	f.StringVar(&cfg.UI.Addr, "ui.addr", Default.UI.Addr, "address the UI/API is listening on")
	f.StringVar(&cfg.UI.Color, "ui.color", Default.UI.Color, "background color of the UI")
	f.StringVar(&cfg.UI.Title, "ui.title", Default.UI.Title, "optional title for the UI")

	var awsApiGWCertCN string
	f.StringVar(&awsApiGWCertCN, "aws.apigw.cert.cn", "", "deprecated. use caupgcn=<CN> for cert source")

	// filter out -test flags
	var args []string
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "-test.") {
			continue
		}
		args = append(args, a)
	}

	// parse configuration
	prefixes := []string{"FABIO_", ""}
	if err := f.ParseFlags(args, os.Environ(), prefixes, p); err != nil {
		return nil, err
	}

	// post configuration
	if cfg.Runtime.GOMAXPROCS == -1 {
		cfg.Runtime.GOMAXPROCS = runtime.NumCPU()
	}

	cfg.Registry.Consul.Scheme, cfg.Registry.Consul.Addr = parseScheme(cfg.Registry.Consul.Addr)

	cfg.CertSources, err = parseCertSources(cfg.CertSourcesValue)
	if err != nil {
		return nil, err
	}

	cfg.Listen, err = parseListeners(cfg.ListenerValue, cfg.CertSources, cfg.Proxy.ReadTimeout, cfg.Proxy.WriteTimeout)
	if err != nil {
		return nil, err
	}

	// handle deprecations
	// deprecate := func(name, msg string) {
	// 	if f.IsSet(name) {
	// 		log.Print("[WARN] ", msg)
	// 	}
	// }

	return cfg, nil
}

// parseScheme splits a url into scheme and address and defaults
// to "http" if no scheme was given.
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

func parseListeners(cfgs []string, cs map[string]CertSource, readTimeout, writeTimeout time.Duration) (listen []Listen, err error) {
	for _, cfg := range cfgs {
		l, err := parseListen(cfg, cs, readTimeout, writeTimeout)
		if err != nil {
			return nil, err
		}
		listen = append(listen, l)
	}
	return
}

func parseListen(cfg string, cs map[string]CertSource, readTimeout, writeTimeout time.Duration) (l Listen, err error) {
	if cfg == "" {
		return Listen{}, nil
	}

	opts := strings.Split(cfg, ";")
	if len(opts) > 1 && !strings.Contains(opts[1], "=") {
		return parseLegacyListen(cfg, readTimeout, writeTimeout)
	}

	l = Listen{
		Addr:         opts[0],
		Scheme:       "http",
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	for k, v := range kvParse(cfg) {
		switch k {
		case "rt": // read timeout
			d, err := time.ParseDuration(v)
			if err != nil {
				return Listen{}, err
			}
			l.ReadTimeout = d
		case "wt": // write timeout
			d, err := time.ParseDuration(v)
			if err != nil {
				return Listen{}, err
			}
			l.WriteTimeout = d
		case "cs": // cert source
			c, ok := cs[v]
			if !ok {
				return Listen{}, fmt.Errorf("unknown certificate source %s", v)
			}
			l.CertSource = c
			l.Scheme = "https"
		}
	}
	return
}

func parseLegacyListen(cfg string, readTimeout, writeTimeout time.Duration) (l Listen, err error) {
	opts := strings.Split(cfg, ";")

	l = Listen{
		Addr:         opts[0],
		Scheme:       "http",
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	if len(opts) > 1 {
		l.Scheme = "https"
		l.CertSource.Type = "file"
		l.CertSource.CertPath = opts[1]
	}
	if len(opts) > 2 {
		l.CertSource.KeyPath = opts[2]
	}
	if len(opts) > 3 {
		l.CertSource.ClientCAPath = opts[3]
	}
	if len(opts) > 4 {
		return Listen{}, fmt.Errorf("invalid listener configuration")
	}

	log.Printf("[WARN] proxy.addr legacy configuration for certificates is deprecated. Use cs=path configuration")
	return l, nil
}

func parseCertSources(cfgs []map[string]string) (cs map[string]CertSource, err error) {
	cs = map[string]CertSource{}
	for _, cfg := range cfgs {
		src, err := parseCertSource(cfg)
		if err != nil {
			return nil, err
		}
		cs[src.Name] = src
	}
	return
}

func parseCertSource(cfg map[string]string) (c CertSource, err error) {
	if cfg == nil {
		return CertSource{}, nil
	}

	c.Refresh = 3 * time.Second

	for k, v := range cfg {
		switch k {
		case "cs":
			c.Name = v
		case "type":
			c.Type = v
		case "cert":
			c.CertPath = v
		case "key":
			c.KeyPath = v
		case "clientca":
			c.ClientCAPath = v
		case "caupgcn":
			c.CAUpgradeCN = v
		case "refresh":
			d, err := time.ParseDuration(v)
			if err != nil {
				return CertSource{}, err
			}
			c.Refresh = d
		case "hdr":
			p := strings.SplitN(v, ": ", 2)
			if len(p) != 2 {
				return CertSource{}, fmt.Errorf("invalid header %s", v)
			}
			if c.Header == nil {
				c.Header = http.Header{}
			}
			c.Header.Set(p[0], p[1])
		}
	}
	if c.Name == "" {
		return CertSource{}, fmt.Errorf("missing 'cs' in %s", cfg)
	}
	if c.Type == "" {
		return CertSource{}, fmt.Errorf("missing 'type' in %s", cfg)
	}
	if c.CertPath == "" {
		return CertSource{}, fmt.Errorf("missing 'cert' in %s", cfg)
	}
	if c.Type != "file" && c.Type != "path" && c.Type != "http" && c.Type != "consul" && c.Type != "vault" {
		return CertSource{}, fmt.Errorf("unknown cert source type %s", c.Type)
	}
	if c.Type == "file" {
		c.Refresh = 0
	}
	return
}
