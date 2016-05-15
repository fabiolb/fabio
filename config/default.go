package config

import (
	"runtime"
	"time"
)

var Default = &Config{
	Proxy: Proxy{
		MaxConn:     10000,
		Strategy:    "rnd",
		Matcher:     "prefix",
		DialTimeout: 30 * time.Second,
		LocalIP:     LocalIPString(),
	},
	Registry: Registry{
		Backend: "consul",
		Consul: Consul{
			Addr:          "localhost:8500",
			KVPath:        "/fabio/config",
			TagPrefix:     "urlprefix-",
			Register:      true,
			ServiceAddr:   ":9998",
			ServiceName:   "fabio",
			CheckInterval: time.Second,
			CheckTimeout:  3 * time.Second,
		},
	},
	Listen: []Listen{
		{
			Addr: ":9999",
		},
	},
	Runtime: Runtime{
		GOGC:       800,
		GOMAXPROCS: runtime.NumCPU(),
	},
	UI: UI{
		Addr:  ":9998",
		Color: "light-green",
	},
	Metrics: []Metrics{
		{
			Target:   "",
			Prefix:   "default",
			Addr:     "",
			Interval: 30 * time.Second,
		},
	},
}
