package config

import (
	"runtime"
	"time"
)

var defaultValues = struct {
	ListenerValue         []string
	CertSourcesValue      []map[string]string
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	GZIPContentTypesValue string
}{
	ListenerValue:    []string{":9999"},
	CertSourcesValue: []map[string]string{},
}

var defaultConfig = &Config{
	Log: Log{
		AccessFormat: "common",
		RoutesFormat: "delta",
	},
	Metrics: Metrics{
		Prefix:   "{{clean .Hostname}}.{{clean .Exec}}",
		Names:    "{{clean .Service}}.{{clean .Host}}.{{clean .Path}}.{{clean .TargetURL.Host}}",
		Interval: 30 * time.Second,
		Circonus: Circonus{
			APIApp: "fabio",
		},
	},
	Proxy: Proxy{
		MaxConn:       10000,
		Strategy:      "rnd",
		Matcher:       "prefix",
		NoRouteStatus: 404,
		DialTimeout:   30 * time.Second,
		FlushInterval: time.Second,
		LocalIP:       LocalIPString(),
	},
	Registry: Registry{
		Backend: "consul",
		Consul: Consul{
			Addr:          "localhost:8500",
			Scheme:        "http",
			KVPath:        "/fabio/config",
			TagPrefix:     "urlprefix-",
			Register:      true,
			ServiceAddr:   ":9998",
			ServiceName:   "fabio",
			ServiceStatus: []string{"passing"},
			CheckInterval: time.Second,
			CheckTimeout:  3 * time.Second,
		},
		Timeout: 10 * time.Second,
		Retry:   500 * time.Millisecond,
	},
	Runtime: Runtime{
		GOGC:       800,
		GOMAXPROCS: runtime.NumCPU(),
	},
	UI: UI{
		Addr:  ":9998",
		Color: "light-green",
	},
}
