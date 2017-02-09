package config

import (
	"net/http"
	"regexp"
	"time"
)

type Config struct {
	Proxy    Proxy
	Registry Registry
	Listen   []Listen
	Metrics  Metrics
	UI       UI
	Runtime  Runtime
}

type CertSource struct {
	Name         string
	Type         string
	CertPath     string
	KeyPath      string
	ClientCAPath string
	CAUpgradeCN  string
	Refresh      time.Duration
	Header       http.Header
}

type Listen struct {
	Addr         string
	Proto        string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	CertSource   CertSource
	StrictMatch  bool
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
	FlushInterval         time.Duration
	LocalIP               string
	ClientIPHeader        string
	TLSHeader             string
	TLSHeaderValue        string
	GZIPContentTypes      *regexp.Regexp
	LogRoutes             string
}

type Runtime struct {
	GOGC       int
	GOMAXPROCS int
}

type Circonus struct {
	APIKey   string
	APIApp   string
	APIURL   string
	CheckID  string
	BrokerID string
}

type Metrics struct {
	Target       string
	Prefix       string
	Names        string
	Interval     time.Duration
	GraphiteAddr string
	StatsDAddr   string
	Circonus     Circonus
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
	ServiceStatus []string
	CheckInterval time.Duration
	CheckTimeout  time.Duration
}
