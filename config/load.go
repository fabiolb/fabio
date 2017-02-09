package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/magiconair/properties"
)

func Load(args, environ []string) (cfg *Config, err error) {
	var props *properties.Properties

	cmdline, path, version, err := parse(args)
	switch {
	case err != nil:
		return nil, err
	case version:
		return nil, nil
	case path != "":
		switch {
		case strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://"):
			props, err = properties.LoadURL(path)
		case path != "":
			props, err = properties.LoadFile(path, properties.UTF8)
		}
		if err != nil {
			return nil, err
		}
	}
	envprefix := []string{"FABIO_", ""}
	return load(cmdline, environ, envprefix, props)
}

var errInvalidConfig = errors.New("invalid or missing path to config file")

// parse extracts the version and config file flags from the command
// line arguments and returns the individual parts. Test flags are
// ignored.
func parse(args []string) (cmdline []string, path string, version bool, err error) {
	if len(args) < 1 {
		panic("missing exec name")
	}

	// always copy the name of the executable
	cmdline = args[:1]

	// parse rest of the arguments
	for i := 1; i < len(args); i++ {
		arg := args[i]

		switch {
		// version flag
		case arg == "-v" || arg == "--version":
			return nil, "", true, nil

		// config file without '='
		case arg == "-cfg" || arg == "--cfg":
			if i >= len(args)-1 {
				return nil, "", false, errInvalidConfig
			}
			path = args[i+1]
			i++

		// config file with '='. needs unquoting
		case strings.HasPrefix(arg, "-cfg=") || strings.HasPrefix(arg, "--cfg="):
			if strings.HasPrefix(arg, "-cfg=") {
				path = arg[len("-cfg="):]
			} else {
				path = arg[len("--cfg="):]
			}
			switch {
			case path == "":
				return nil, "", false, errInvalidConfig
			case path[0] == '\'':
				path = strings.Trim(path, "'")
			case path[0] == '"':
				path = strings.Trim(path, "\"")
			}
			if path == "" {
				return nil, "", false, errInvalidConfig
			}

		// ignore test flags
		case strings.HasPrefix(arg, "-test."):
			continue

		default:
			cmdline = append(cmdline, arg)
		}
	}
	return cmdline, path, false, nil
}

func load(cmdline, environ, envprefix []string, props *properties.Properties) (cfg *Config, err error) {
	cfg = &Config{}
	f := NewFlagSet(cmdline[0], flag.ExitOnError)

	// dummy values which were parsed earlier
	f.String("cfg", "", "Path or URL to config file")
	f.Bool("v", false, "Show version")
	f.Bool("version", false, "Show version")

	// config values
	var listenerValue []string
	var certSourcesValue []map[string]string
	var readTimeout, writeTimeout time.Duration
	var gzipContentTypesValue string

	f.IntVar(&cfg.Proxy.MaxConn, "proxy.maxconn", defaultConfig.Proxy.MaxConn, "maximum number of cached connections")
	f.StringVar(&cfg.Proxy.Strategy, "proxy.strategy", defaultConfig.Proxy.Strategy, "load balancing strategy")
	f.StringVar(&cfg.Proxy.Matcher, "proxy.matcher", defaultConfig.Proxy.Matcher, "path matching algorithm")
	f.IntVar(&cfg.Proxy.NoRouteStatus, "proxy.noroutestatus", defaultConfig.Proxy.NoRouteStatus, "status code for invalid route")
	f.DurationVar(&cfg.Proxy.ShutdownWait, "proxy.shutdownwait", defaultConfig.Proxy.ShutdownWait, "time for graceful shutdown")
	f.DurationVar(&cfg.Proxy.DialTimeout, "proxy.dialtimeout", defaultConfig.Proxy.DialTimeout, "connection timeout for backend connections")
	f.DurationVar(&cfg.Proxy.ResponseHeaderTimeout, "proxy.responseheadertimeout", defaultConfig.Proxy.ResponseHeaderTimeout, "response header timeout")
	f.DurationVar(&cfg.Proxy.KeepAliveTimeout, "proxy.keepalivetimeout", defaultConfig.Proxy.KeepAliveTimeout, "keep-alive timeout")
	f.StringVar(&cfg.Proxy.LocalIP, "proxy.localip", defaultConfig.Proxy.LocalIP, "fabio address in Forward headers")
	f.StringVar(&cfg.Proxy.ClientIPHeader, "proxy.header.clientip", defaultConfig.Proxy.ClientIPHeader, "header for the request ip")
	f.StringVar(&cfg.Proxy.TLSHeader, "proxy.header.tls", defaultConfig.Proxy.TLSHeader, "header for TLS connections")
	f.StringVar(&cfg.Proxy.TLSHeaderValue, "proxy.header.tls.value", defaultConfig.Proxy.TLSHeaderValue, "value for TLS connection header")
	f.StringVar(&gzipContentTypesValue, "proxy.gzip.contenttype", defaultValues.GZIPContentTypesValue, "regexp of content types to compress")
	f.StringSliceVar(&listenerValue, "proxy.addr", defaultValues.ListenerValue, "listener config")
	f.KVSliceVar(&certSourcesValue, "proxy.cs", defaultValues.CertSourcesValue, "certificate sources")
	f.DurationVar(&readTimeout, "proxy.readtimeout", defaultValues.ReadTimeout, "read timeout for incoming requests")
	f.DurationVar(&writeTimeout, "proxy.writetimeout", defaultValues.WriteTimeout, "write timeout for outgoing responses")
	f.DurationVar(&cfg.Proxy.FlushInterval, "proxy.flushinterval", defaultConfig.Proxy.FlushInterval, "flush interval for streaming responses")
	f.StringVar(&cfg.Proxy.LogRoutes, "proxy.log.routes", defaultConfig.Proxy.LogRoutes, "log format of routing table updates")
	f.StringVar(&cfg.Metrics.Target, "metrics.target", defaultConfig.Metrics.Target, "metrics backend")
	f.StringVar(&cfg.Metrics.Prefix, "metrics.prefix", defaultConfig.Metrics.Prefix, "prefix for reported metrics")
	f.StringVar(&cfg.Metrics.Names, "metrics.names", defaultConfig.Metrics.Names, "route metric name template")
	f.DurationVar(&cfg.Metrics.Interval, "metrics.interval", defaultConfig.Metrics.Interval, "metrics reporting interval")
	f.StringVar(&cfg.Metrics.GraphiteAddr, "metrics.graphite.addr", defaultConfig.Metrics.GraphiteAddr, "graphite server address")
	f.StringVar(&cfg.Metrics.StatsDAddr, "metrics.statsd.addr", defaultConfig.Metrics.StatsDAddr, "statsd server address")
	f.StringVar(&cfg.Metrics.Circonus.APIKey, "metrics.circonus.apikey", defaultConfig.Metrics.Circonus.APIKey, "Circonus API token key")
	f.StringVar(&cfg.Metrics.Circonus.APIApp, "metrics.circonus.apiapp", defaultConfig.Metrics.Circonus.APIApp, "Circonus API token app")
	f.StringVar(&cfg.Metrics.Circonus.APIURL, "metrics.circonus.apiurl", defaultConfig.Metrics.Circonus.APIURL, "Circonus API URL")
	f.StringVar(&cfg.Metrics.Circonus.BrokerID, "metrics.circonus.brokerid", defaultConfig.Metrics.Circonus.BrokerID, "Circonus Broker ID")
	f.StringVar(&cfg.Metrics.Circonus.CheckID, "metrics.circonus.checkid", defaultConfig.Metrics.Circonus.CheckID, "Circonus Check ID")
	f.StringVar(&cfg.Registry.Backend, "registry.backend", defaultConfig.Registry.Backend, "registry backend")
	f.StringVar(&cfg.Registry.File.Path, "registry.file.path", defaultConfig.Registry.File.Path, "path to file based routing table")
	f.StringVar(&cfg.Registry.Static.Routes, "registry.static.routes", defaultConfig.Registry.Static.Routes, "static routes")
	f.StringVar(&cfg.Registry.Consul.Addr, "registry.consul.addr", defaultConfig.Registry.Consul.Addr, "address of the consul agent")
	f.StringVar(&cfg.Registry.Consul.Token, "registry.consul.token", defaultConfig.Registry.Consul.Token, "token for consul agent")
	f.StringVar(&cfg.Registry.Consul.KVPath, "registry.consul.kvpath", defaultConfig.Registry.Consul.KVPath, "consul KV path for manual overrides")
	f.StringVar(&cfg.Registry.Consul.TagPrefix, "registry.consul.tagprefix", defaultConfig.Registry.Consul.TagPrefix, "prefix for consul tags")
	f.BoolVar(&cfg.Registry.Consul.Register, "registry.consul.register.enabled", defaultConfig.Registry.Consul.Register, "register fabio in consul")
	f.StringVar(&cfg.Registry.Consul.ServiceAddr, "registry.consul.register.addr", defaultConfig.Registry.Consul.ServiceAddr, "service registration address")
	f.StringVar(&cfg.Registry.Consul.ServiceName, "registry.consul.register.name", defaultConfig.Registry.Consul.ServiceName, "service registration name")
	f.StringSliceVar(&cfg.Registry.Consul.ServiceTags, "registry.consul.register.tags", defaultConfig.Registry.Consul.ServiceTags, "service registration tags")
	f.StringSliceVar(&cfg.Registry.Consul.ServiceStatus, "registry.consul.service.status", defaultConfig.Registry.Consul.ServiceStatus, "valid service status values")
	f.DurationVar(&cfg.Registry.Consul.CheckInterval, "registry.consul.register.checkInterval", defaultConfig.Registry.Consul.CheckInterval, "service check interval")
	f.DurationVar(&cfg.Registry.Consul.CheckTimeout, "registry.consul.register.checkTimeout", defaultConfig.Registry.Consul.CheckTimeout, "service check timeout")
	f.IntVar(&cfg.Runtime.GOGC, "runtime.gogc", defaultConfig.Runtime.GOGC, "sets runtime.GOGC")
	f.IntVar(&cfg.Runtime.GOMAXPROCS, "runtime.gomaxprocs", defaultConfig.Runtime.GOMAXPROCS, "sets runtime.GOMAXPROCS")
	f.StringVar(&cfg.UI.Addr, "ui.addr", defaultConfig.UI.Addr, "address the UI/API is listening on")
	f.StringVar(&cfg.UI.Color, "ui.color", defaultConfig.UI.Color, "background color of the UI")
	f.StringVar(&cfg.UI.Title, "ui.title", defaultConfig.UI.Title, "optional title for the UI")

	var awsApiGWCertCN string
	f.StringVar(&awsApiGWCertCN, "aws.apigw.cert.cn", "", "deprecated. use caupgcn=<CN> for cert source")

	// parse configuration
	if err := f.ParseFlags(cmdline[1:], environ, envprefix, props); err != nil {
		return nil, err
	}

	// post configuration
	if cfg.Runtime.GOMAXPROCS == -1 {
		cfg.Runtime.GOMAXPROCS = runtime.NumCPU()
	}

	cfg.Registry.Consul.Scheme, cfg.Registry.Consul.Addr = parseScheme(cfg.Registry.Consul.Addr)

	certSources, err := parseCertSources(certSourcesValue)
	if err != nil {
		return nil, err
	}

	cfg.Listen, err = parseListeners(listenerValue, certSources, readTimeout, writeTimeout)
	if err != nil {
		return nil, err
	}

	if gzipContentTypesValue != "" {
		cfg.Proxy.GZIPContentTypes, err = regexp.Compile(gzipContentTypesValue)
		if err != nil {
			return nil, fmt.Errorf("invalid expression for content types: %s", err)
		}
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
	switch {
	case strings.HasPrefix(s, "https://"):
		scheme, addr = "https", s[len("https://"):]
	case strings.HasPrefix(s, "http://"):
		scheme, addr = "http", s[len("http://"):]
	default:
		scheme, addr = "http", s
	}

	// strip off anything after a final slash
	if n := strings.Index(addr, "/"); n >= 0 {
		addr = addr[:n]
	}
	return
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
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	var csName string
	for k, v := range kvParse(cfg) {
		switch k {
		case "proto":
			l.Proto = v
			if l.Proto != "http" && l.Proto != "https" && l.Proto != "tcp+sni" {
				return Listen{}, fmt.Errorf("unknown protocol %q", v)
			}
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
			csName = v
			c, ok := cs[v]
			if !ok {
				return Listen{}, fmt.Errorf("unknown certificate source %q", v)
			}
			l.CertSource = c
			if l.Proto == "" {
				l.Proto = "https"
			}
		case "strictmatch":
			l.StrictMatch = (v == "true")
		}
	}

	if l.Proto == "" {
		l.Proto = "http"
	}
	if csName != "" && l.Proto != "https" {
		return Listen{}, fmt.Errorf("cert source requires proto 'https'")
	}
	if csName == "" && l.Proto == "https" {
		return Listen{}, fmt.Errorf("proto 'https' requires cert source")
	}

	return
}

func parseLegacyListen(cfg string, readTimeout, writeTimeout time.Duration) (l Listen, err error) {
	opts := strings.Split(cfg, ";")

	l = Listen{
		Addr:         opts[0],
		Proto:        "http",
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	if len(opts) > 1 {
		l.Proto = "https"
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
