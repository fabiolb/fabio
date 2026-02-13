package config

import (
	"net/http"
	"regexp"
	"time"
)

type Config struct {
	Log                  Log
	ProfileMode          string
	ProfilePath          string
	Listen               []Listen
	Metrics              Metrics
	BGP                  BGP
	UI                   UI
	Registry             Registry
	Proxy                Proxy
	Runtime              Runtime
	GlobCacheSize        int
	Insecure             bool
	GlobMatchingDisabled bool
}

type CertSource struct {
	Header          http.Header
	Name            string
	Type            string
	CertPath        string
	KeyPath         string
	ClientCAPath    string
	CAUpgradeCN     string
	VaultFetchToken string
	Refresh         time.Duration
}

type Listen struct {
	CertSource         CertSource
	Addr               string
	Proto              string
	TLSCiphers         []uint16
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	ProxyHeaderTimeout time.Duration
	Refresh            time.Duration
	TLSMinVersion      uint16
	TLSMaxVersion      uint16
	StrictMatch        bool
	ProxyProto         bool
}

type Source struct {
	Scheme      string
	Host        string
	Port        string
	LinkEnabled bool
	NewTab      bool
}

type RoutingTable struct {
	Source Source
}

type UI struct {
	RoutingTable RoutingTable
	Color        string
	Title        string
	Access       string
	Listen       Listen
}

type Proxy struct {
	GZIPContentTypes      *regexp.Regexp
	AuthSchemes           map[string]AuthScheme
	Strategy              string
	Matcher               string
	LocalIP               string
	ClientIPHeader        string
	TLSHeader             string
	TLSHeaderValue        string
	RequestID             string
	STSHeader             STSHeader
	NoRouteStatus         int
	MaxConn               int
	ShutdownWait          time.Duration
	DeregisterGracePeriod time.Duration
	DialTimeout           time.Duration
	ResponseHeaderTimeout time.Duration
	KeepAliveTimeout      time.Duration
	IdleConnTimeout       time.Duration
	FlushInterval         time.Duration
	GlobalFlushInterval   time.Duration
	GRPCMaxRxMsgSize      int
	GRPCMaxTxMsgSize      int
	GRPCGShutdownTimeout  time.Duration
}

type STSHeader struct {
	MaxAge     int
	Subdomains bool
	Preload    bool
}

type Runtime struct {
	GOGC       int
	GOMAXPROCS int
}

type Circonus struct {
	APIKey        string
	APIApp        string
	APIURL        string
	CheckID       string
	BrokerID      string
	SubmissionURL string
}

type Prometheus struct {
	Subsystem string
	Path      string
	Buckets   []float64
}

type Log struct {
	AccessFormat string
	AccessTarget string
	RoutesFormat string
	Level        string
}

type Metrics struct {
	Circonus      Circonus
	Target        string
	Prefix        string
	Names         string
	GraphiteAddr  string
	StatsDAddr    string
	DogstatsdAddr string
	Prometheus    Prometheus
	Interval      time.Duration
	Timeout       time.Duration
	Retry         time.Duration
}

type Registry struct {
	Static  Static
	File    File
	Backend string
	Custom  Custom
	Consul  Consul
	Timeout time.Duration
	Retry   time.Duration
}

type Static struct {
	NoRouteHTML string
	Routes      string
}

type File struct {
	NoRouteHTMLPath string
	RoutesPath      string
}

type Consul struct {
	Addr               string
	Scheme             string
	Token              string
	KVPath             string
	NoRouteHTMLPath    string
	TagPrefix          string
	Namespace          string
	ServiceAddr        string
	ServiceName        string
	CheckScheme        string
	ChecksRequired     string
	TLS                ConsulTlS
	ServiceTags        []string
	ServiceStatus      []string
	CheckInterval      time.Duration
	CheckTimeout       time.Duration
	ServiceMonitors    int
	PollInterval       time.Duration
	Register           bool
	CheckTLSSkipVerify bool
	RequireConsistent  bool
	AllowStale         bool
}

type Custom struct {
	Host               string
	Path               string
	QueryParams        string
	Scheme             string
	NoRouteHTML        string
	PollInterval       time.Duration
	Timeout            time.Duration
	CheckTLSSkipVerify bool
}

type AuthScheme struct {
	Name  string
	Type  string
	Basic BasicAuth
}

type BasicAuth struct {
	ModTime time.Time // the htpasswd file last modification time
	Realm   string
	File    string
	Refresh time.Duration
}

type ConsulTlS struct {
	KeyFile            string
	CertFile           string
	CAFile             string
	CAPath             string
	InsecureSkipVerify bool
}

type BGP struct {
	RouterID          string
	GRPCListenAddress string
	CertFile          string
	KeyFile           string
	GOBGPDCfgFile     string
	NextHop           string
	AnycastAddresses  []string
	ListenAddresses   []string
	Peers             []BGPPeer
	Asn               uint
	ListenPort        int
	BGPEnabled        bool
	EnableGRPC        bool
	GRPCTLS           bool
}

type BGPPeer struct {
	NeighborAddress string
	Password        string
	NeighborPort    uint
	Asn             uint
	MultiHopLength  uint
	MultiHop        bool
}
