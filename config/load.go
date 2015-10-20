package config

import (
	"fmt"
	"net"
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

func FromProperties(p *properties.Properties) (*Config, error) {
	var cfg *Config = &Config{}
	var err error

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
		Addr:      p.GetString("consul.addr", DefaultConfig.Consul.Addr),
		KVPath:    p.GetString("consul.kvpath", DefaultConfig.Consul.KVPath),
		TagPrefix: p.GetString("consul.tagprefix", DefaultConfig.Consul.TagPrefix),
		URL:       p.GetString("consul.url", DefaultConfig.Consul.URL),
	}

	cfg.Runtime = Runtime{
		GOGC:       p.GetInt("runtime.gogc", DefaultConfig.Runtime.GOGC),
		GOMAXPROCS: p.GetInt("runtime.gomaxprocs", DefaultConfig.Runtime.GOMAXPROCS),
	}
	if cfg.Runtime.GOMAXPROCS == -1 {
		cfg.Runtime.GOMAXPROCS = runtime.NumCPU()
	}

	cfg.UI = UI{
		Addr: p.GetString("ui.addr", DefaultConfig.UI.Addr),
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

// localIP tries to determine a non-loopback address for the local machine
func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
			if ipnet.IP.To4() != nil || ipnet.IP.To16() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
