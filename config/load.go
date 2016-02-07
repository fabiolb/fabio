package config

import (
	"encoding/json"
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
		ShutdownWait:          durationVal(p, Default.Proxy.ShutdownWait, "proxy.shutdownwait"),
		DialTimeout:           durationVal(p, Default.Proxy.DialTimeout, "proxy.dialtimeout"),
		ResponseHeaderTimeout: durationVal(p, Default.Proxy.ResponseHeaderTimeout, "proxy.timeout"),
		KeepAliveTimeout:      durationVal(p, Default.Proxy.KeepAliveTimeout, "proxy.timeout"),
		LocalIP:               stringVal(p, Default.Proxy.LocalIP, "proxy.localip"),
		ClientIPHeader:        stringVal(p, Default.Proxy.ClientIPHeader, "proxy.header.clientip"),
		TLSHeader:             stringVal(p, Default.Proxy.TLSHeader, "proxy.header.tls"),
		TLSHeaderValue:        stringVal(p, Default.Proxy.TLSHeaderValue, "proxy.header.tls.value"),
	}

	cfg.Listen, err = parseListen(stringVal(p, Default.Listen[0].Addr, "proxy.addr"))
	if err != nil {
		return nil, err
	}

	cfg.Routes = stringVal(p, Default.Routes, "proxy.routes")

	cfg.Metrics = parseMetrics(
		stringVal(p, Default.Metrics[0].Target, "metrics.target"),
		stringVal(p, Default.Metrics[0].Prefix, "metrics.prefix"),
		stringVal(p, Default.Metrics[0].Addr, "metrics.graphite.addr"),
		durationVal(p, Default.Metrics[0].Interval, "metrics.interval"),
	)

	cfg.Consul = Consul{
		Addr:          stringVal(p, Default.Consul.Addr, "consul.addr"),
		Token:         stringVal(p, Default.Consul.Token, "consul.token"),
		KVPath:        stringVal(p, Default.Consul.KVPath, "consul.kvpath"),
		TagPrefix:     stringVal(p, Default.Consul.TagPrefix, "consul.tagprefix"),
		ServiceName:   stringVal(p, Default.Consul.ServiceName, "consul.register.name"),
		CheckInterval: durationVal(p, Default.Consul.CheckInterval, "consul.register.checkInterval"),
		CheckTimeout:  durationVal(p, Default.Consul.CheckTimeout, "consul.register.checkTimeout"),
	}
	deprecate("consul.url", "consul.url is obsolete. Please remove it.")

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

	dump(cfg)
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

func dump(cfg *Config) {
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		log.Fatal("[FATAL] Cannot dump runtime config. ", err)
	}
	log.Println("[INFO] Runtime config\n" + string(data))
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
