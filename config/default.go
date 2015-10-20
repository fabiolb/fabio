package config

import (
	"runtime"
	"time"
)

var DefaultConfig = &Config{
	Proxy: Proxy{
		MaxConn:     10000,
		Strategy:    "rnd",
		DialTimeout: 30 * time.Second,
		LocalIP:     localIP(),
	},
	Listen: []Listen{
		{
			Addr: ":9999",
		},
	},
	Consul: Consul{
		Addr:      "localhost:8500",
		KVPath:    "/fabio/config",
		TagPrefix: "urlprefix-",
		URL:       "http://localhost:8500/",
	},
	Runtime: Runtime{
		GOGC:       800,
		GOMAXPROCS: runtime.NumCPU(),
	},
	UI: UI{
		Addr: ":9998",
	},
}
