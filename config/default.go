package config

import (
	"runtime"
	"time"
)

var Default = &Config{
	Proxy: Proxy{
		MaxConn:      10000,
		Strategy:     "rnd",
		Matcher:      "prefix",
		NoRouteStatus: 404,
		DialTimeout:  30 * time.Second,
		LocalIP:      LocalIPString(),
		ListenerAddr: ":9999",
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
			CheckInterval: time.Second,
			CheckTimeout:  3 * time.Second,
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
	Metrics: Metrics{
		Prefix:   "default",
		Interval: 30 * time.Second,
	},
}
