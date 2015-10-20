package config

import "time"

type Config struct {
	Proxy   Proxy
	Listen  []Listen
	Routes  string
	Metrics []Metrics
	Consul  Consul
	UI      UI
	Runtime Runtime
}

type Listen struct {
	Addr           string
	KeyFile        string
	CertFile       string
	ClientAuthFile string
	TLS            bool
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

type UI struct {
	Addr string
}

type Proxy struct {
	Strategy              string
	MaxConn               int
	ShutdownWait          time.Duration
	DialTimeout           time.Duration
	ResponseHeaderTimeout time.Duration
	KeepAliveTimeout      time.Duration
	LocalIP               string
	ClientIPHeader        string
	TLSHeader             string
	TLSHeaderValue        string
}

type Runtime struct {
	GOGC       int
	GOMAXPROCS int
}

type Metrics struct {
	Target   string
	Prefix   string
	Interval time.Duration
	Addr     string
}

type Consul struct {
	Addr      string
	KVPath    string
	TagPrefix string
	URL       string
}
