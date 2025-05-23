package transport

import (
	"crypto/tls"
	"github.com/fabiolb/fabio/config"
	"net"
	"net/http"
)

var (
	cfg *config.Config = &config.Config{}
)

func NewTransport(tlscfg *tls.Config) *http.Transport {
	return &http.Transport{
		ResponseHeaderTimeout: cfg.Proxy.ResponseHeaderTimeout,
		IdleConnTimeout:       cfg.Proxy.IdleConnTimeout,
		MaxIdleConnsPerHost:   cfg.Proxy.MaxConn,
		Dial: (&net.Dialer{
			Timeout:   cfg.Proxy.DialTimeout,
			KeepAlive: cfg.Proxy.KeepAliveTimeout,
		}).Dial,
		TLSClientConfig: tlscfg,
	}
}

func SetConfig(ncfg *config.Config) {
	cfg = ncfg
}
