package config

import (
	"net/http"
	"time"
)

type Config struct {
	Proxy       Proxy
	Registry    Registry
	Listen      []Listen
	CertSources map[string]CertSource
	Metrics     Metrics
	UI          UI
	Runtime     Runtime

	ListenerValue    []string
	CertSourcesValue []map[string]string
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
	Scheme       string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	CertSource   CertSource
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
	FlushInterval         time.Duration
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
	ServiceStatus []string
	CheckInterval time.Duration
	CheckTimeout  time.Duration
}
