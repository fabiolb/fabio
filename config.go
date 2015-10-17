package main

import (
	"runtime"
	"time"

	"github.com/eBay/fabio/_third_party/github.com/magiconair/properties"
)

var (
	proxyAddr           = ":9999"
	proxyMaxConn        = 10000
	proxyRoutes         = ""
	proxyStrategy       = "rnd"
	proxyShutdownWait   = time.Duration(0)
	proxyDialTimeout    = 30 * time.Second
	proxyTimeout        = time.Duration(0)
	proxyHeaderClientIP = ""
	proxyHeaderTLS      = ""
	proxyHeaderTLSValue = ""
	consulAddr          = "localhost:8500"
	consulKVPath        = "/fabio/config"
	consulTagPrefix     = "urlprefix-"
	consulURL           = "http://" + consulAddr + "/"
	metricsTarget       = ""
	metricsInterval     = 30 * time.Second
	metricsPrefix       = "default"
	metricsGraphiteAddr = ""
	gogc                = 800
	gomaxprocs          = runtime.NumCPU()
	uiAddr              = ":9998"
)

func loadConfig(filename string) error {
	p, err := properties.LoadFile(filename, properties.UTF8)
	if err != nil {
		return err
	}

	proxyAddr = p.GetString("proxy.addr", proxyAddr)
	proxyMaxConn = p.GetInt("proxy.maxconn", proxyMaxConn)
	proxyRoutes = p.GetString("proxy.routes", proxyRoutes)
	proxyStrategy = p.GetString("proxy.strategy", proxyStrategy)
	proxyShutdownWait = p.GetParsedDuration("proxy.shutdownWait", proxyShutdownWait)
	proxyDialTimeout = p.GetParsedDuration("proxy.dialtimeout", proxyDialTimeout)
	proxyTimeout = p.GetParsedDuration("proxy.timeout", proxyTimeout)
	proxyHeaderClientIP = p.GetString("proxy.header.clientip", proxyHeaderClientIP)
	proxyHeaderTLS = p.GetString("proxy.header.tls", proxyHeaderTLS)
	proxyHeaderTLSValue = p.GetString("proxy.header.tls.value", proxyHeaderTLSValue)
	consulAddr = p.GetString("consul.addr", consulAddr)
	consulKVPath = p.GetString("consul.kvpath", consulKVPath)
	consulTagPrefix = p.GetString("consul.tagprefix", consulTagPrefix)
	consulURL = p.GetString("consul.url", "http://"+consulAddr+"/")
	metricsTarget = p.GetString("metrics.target", metricsTarget)
	metricsInterval = p.GetParsedDuration("metrics.interval", metricsInterval)
	metricsPrefix = p.GetString("metrics.prefix", metricsPrefix)
	metricsGraphiteAddr = p.GetString("metrics.graphite.addr", metricsGraphiteAddr)
	gogc = p.GetInt("runtime.gogc", gogc)
	gomaxprocs = p.GetInt("runtime.gomaxprocs", gomaxprocs)
	uiAddr = p.GetString("ui.addr", uiAddr)

	return nil
}
