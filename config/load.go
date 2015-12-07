package config

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/eBay/fabio/_third_party/github.com/magiconair/properties"
)

func FromFile(filename string) (*Config, error) {
	p, err := properties.LoadFile(filename, properties.UTF8)
	if err != nil {
		return nil, err
	}
	return FromProperties(p)
}

func FromProperties(p *properties.Properties) (cfg *Config, err error) {
	cfg = &Config{}

	deprecate := func(key, msg string) {
		_, exists := p.Get(key)
		if exists {
			log.Print("[WARN] config: ", msg)
		}
	}

	cfg.Proxy = Proxy{
		MaxConn:               p.GetInt("proxy.maxconn", DefaultConfig.Proxy.MaxConn),
		Strategy:              p.GetString("proxy.strategy", DefaultConfig.Proxy.Strategy),
		ShutdownWait:          p.GetParsedDuration("proxy.shutdownwait", DefaultConfig.Proxy.ShutdownWait),
		DialTimeout:           p.GetParsedDuration("proxy.dialtimeout", DefaultConfig.Proxy.DialTimeout),
		ResponseHeaderTimeout: p.GetParsedDuration("proxy.timeout", DefaultConfig.Proxy.ResponseHeaderTimeout),
		KeepAliveTimeout:      p.GetParsedDuration("proxy.timeout", DefaultConfig.Proxy.KeepAliveTimeout),
		LocalIP:               p.GetString("proxy.localip", DefaultConfig.Proxy.LocalIP),
		ClientIPHeader:        p.GetString("proxy.header.clientip", DefaultConfig.Proxy.ClientIPHeader),
		TLSHeader:             p.GetString("proxy.header.tls", DefaultConfig.Proxy.TLSHeader),
		TLSHeaderValue:        p.GetString("proxy.header.tls.value", DefaultConfig.Proxy.TLSHeaderValue),
	}

	cfg.Listen, err = parseListen(p.GetString("proxy.addr", DefaultConfig.Listen[0].Addr))
	if err != nil {
		return nil, err
	}

	cfg.Routes = p.GetString("proxy.routes", "")

	cfg.Metrics = parseMetrics(
		p.GetString("metrics.target", ""),
		p.GetString("metrics.prefix", "default"),
		p.GetString("metrics.graphite.addr", ""),
		p.GetParsedDuration("metrics.interval", 30*time.Second),
	)

	cfg.Consul = Consul{
		Addr:          p.GetString("consul.addr", DefaultConfig.Consul.Addr),
		KVPath:        p.GetString("consul.kvpath", DefaultConfig.Consul.KVPath),
		TagPrefix:     p.GetString("consul.tagprefix", DefaultConfig.Consul.TagPrefix),
		ServiceName:   p.GetString("consul.register.name", DefaultConfig.Consul.ServiceName),
		CheckInterval: p.GetParsedDuration("consul.register.checkInterval", DefaultConfig.Consul.CheckInterval),
		CheckTimeout:  p.GetParsedDuration("consul.register.checkTimeout", DefaultConfig.Consul.CheckTimeout),
	}
	deprecate("consul.url", "consul.url is obsolete. Please remove it.")

	cfg.Runtime = Runtime{
		GOGC:       p.GetInt("runtime.gogc", DefaultConfig.Runtime.GOGC),
		GOMAXPROCS: p.GetInt("runtime.gomaxprocs", DefaultConfig.Runtime.GOMAXPROCS),
	}
	if cfg.Runtime.GOMAXPROCS == -1 {
		cfg.Runtime.GOMAXPROCS = runtime.NumCPU()
	}

	cfg.UI = UI{
		Addr:  p.GetString("ui.addr", DefaultConfig.UI.Addr),
		Color: p.GetString("ui.color", DefaultConfig.UI.Color),
		Title: p.GetString("ui.title", DefaultConfig.UI.Title),
	}

	return cfg, nil
}

func parseMetrics(target, prefix, graphiteAddr string, interval time.Duration) []Metrics {
	m := Metrics{Target: target, Prefix: prefix, Interval: interval}
	if target == "graphite" {
		m.Addr = graphiteAddr
	}
	return []Metrics{m}
}

func parseListen(addrs string) ([]Listen, error) {
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
		listen = append(listen, l)
	}
	return listen, nil
}
