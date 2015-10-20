package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/eBay/fabio/_third_party/github.com/magiconair/properties"
)

// TODO(fs): refactor with struct merging
func TestFromProperties(t *testing.T) {
	in := `
proxy.addr = :1234
proxy.localip = 4.4.4.4
proxy.strategy = rr
proxy.shutdownwait = 500ms
proxy.timeout     = 3s
proxy.dialtimeout = 60s
proxy.maxconn = 666
proxy.routes = route add svc / http://127.0.0.1:6666/
proxy.header.clientip = clientip
proxy.header.tls = tls
proxy.header.tls.value = tls-true
consul.addr = 1.2.3.4:5678
consul.url = http://hooray.com/
consul.kvpath = /some/path
consul.tagprefix = p-
metrics.target = graphite
metrics.prefix = someprefix
metrics.interval = 5s
metrics.graphite.addr = 5.6.7.8:9999
runtime.gogc = 666
runtime.gomaxprocs = 12
ui.addr = 7.8.9.0:1234
	`
	out := &Config{
		Proxy: Proxy{
			MaxConn:               666,
			LocalIP:               "4.4.4.4",
			Strategy:              "rr",
			ShutdownWait:          500 * time.Millisecond,
			DialTimeout:           60 * time.Second,
			KeepAliveTimeout:      3 * time.Second,
			ResponseHeaderTimeout: 3 * time.Second,
			ClientIPHeader:        "clientip",
			TLSHeader:             "tls",
			TLSHeaderValue:        "tls-true",
		},
		Listen: []Listen{
			Listen{
				Addr: ":1234",
			},
		},
		Routes: "route add svc / http://127.0.0.1:6666/",
		Consul: Consul{
			Addr:      "1.2.3.4:5678",
			URL:       "http://hooray.com/",
			KVPath:    "/some/path",
			TagPrefix: "p-",
		},
		Metrics: []Metrics{
			Metrics{
				Target:   "graphite",
				Prefix:   "someprefix",
				Interval: 5 * time.Second,
				Addr:     "5.6.7.8:9999",
			},
		},
		Runtime: Runtime{
			GOGC:       666,
			GOMAXPROCS: 12,
		},
		UI: UI{
			Addr: "7.8.9.0:1234",
		},
	}

	p, err := properties.Load([]byte(in), properties.UTF8)
	if err != nil {
		t.Fatalf("got %v want nil", err)
	}

	cfg, err := FromProperties(p)
	if err != nil {
		t.Fatalf("got %v want nil", err)
	}

	if got, want := cfg, out; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestParseAddr(t *testing.T) {
	tests := []struct {
		in  string
		out []Listen
		err string
	}{
		{
			"",
			[]Listen{},
			"",
		},
		{
			":123",
			[]Listen{
				Listen{Addr: ":123"},
			},
			"",
		},
		{
			":123;cert.pem",
			[]Listen{
				Listen{Addr: ":123", CertFile: "cert.pem", KeyFile: "cert.pem", TLS: true},
			},
			"",
		},
		{
			":123;cert.pem;key.pem",
			[]Listen{
				Listen{Addr: ":123", CertFile: "cert.pem", KeyFile: "key.pem", TLS: true},
			},
			"",
		},
		{
			":123;cert.pem;key.pem;client.pem",
			[]Listen{
				Listen{Addr: ":123", CertFile: "cert.pem", KeyFile: "key.pem", ClientAuthFile: "client.pem", TLS: true},
			},
			"",
		},
		{
			":123;cert.pem;key.pem;client.pem;",
			nil,
			"invalid address :123;cert.pem;key.pem;client.pem;",
		},
	}

	for i, tt := range tests {
		l, err := parseListen(tt.in)
		if got, want := err, tt.err; (got != nil || want != "") && got.Error() != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
		if got, want := l, tt.out; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}
