package config

import "time"

type Config struct {
	Proxy    Proxy
	Registry Registry
	Listen   []Listen
	Metrics  Metrics
	UI       UI
	Runtime  Runtime
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
	Addr  string
	Color string
	Title string
}

type Proxy struct {
	Strategy              string
	Matcher               string
	NoRouteStatus         int
	MaxConn               int
	ShutdownWait          time.Duration
	DialTimeout           time.Duration
	ResponseHeaderTimeout time.Duration
	KeepAliveTimeout      time.Duration
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	LocalIP               string
	ClientIPHeader        string
	TLSHeader             string
	TLSHeaderValue        string
	ListenerAddr          string
}

type Runtime struct {
	GOGC       int
	GOMAXPROCS int
}

type Metrics struct {
	Target       string
	Prefix       string
	Interval     time.Duration
	GraphiteAddr string
}

type Registry struct {
	Backend string
	Static  Static
	File    File
	Consul  Consul
}

type Static struct {
	Routes string
}

type File struct {
	Path string
}

type Consul struct {
	Addr          string
	Scheme        string
	Token         string
	KVPath        string
	TagPrefix     string
	Register      bool
	ServiceAddr   string
	ServiceName   string
	ServiceTags   []string
	CheckInterval time.Duration
	CheckTimeout  time.Duration
}
