package config

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/magiconair/properties"
	"github.com/pascaldekloe/goe/verify"
)

func TestFromProperties(t *testing.T) {
	in := `
proxy.addr = :1234
proxy.localip = 4.4.4.4
proxy.strategy = rr
proxy.matcher = prefix
proxy.shutdownwait = 500ms
proxy.timeout = 3s
proxy.dialtimeout = 60s
proxy.readtimeout = 5s
proxy.writetimeout = 10s
proxy.maxconn = 666
proxy.header.clientip = clientip
proxy.header.tls = tls
proxy.header.tls.value = tls-true
registry.backend = something
registry.file.path = /foo/bar
registry.static.routes = route add svc / http://127.0.0.1:6666/
registry.consul.addr = 1.2.3.4:5678
registry.consul.token = consul-token
registry.consul.kvpath = /some/path
registry.consul.tagprefix = p-
registry.consul.register.addr = 6.6.6.6:7777
registry.consul.register.name = fab
registry.consul.register.tags = a, b, c ,
registry.consul.register.checkInterval = 5s
registry.consul.register.checkTimeout = 10s
metrics.target = graphite
metrics.prefix = someprefix
metrics.interval = 5s
metrics.graphite.addr = 5.6.7.8:9999
runtime.gogc = 666
runtime.gomaxprocs = 12
ui.addr = 7.8.9.0:1234
ui.color = fonzy
ui.title = fabfab
	`
	out := &Config{
		Proxy: Proxy{
			MaxConn:               666,
			LocalIP:               "4.4.4.4",
			Strategy:              "rr",
			Matcher:               "prefix",
			ShutdownWait:          500 * time.Millisecond,
			DialTimeout:           60 * time.Second,
			KeepAliveTimeout:      3 * time.Second,
			ResponseHeaderTimeout: 3 * time.Second,
			ClientIPHeader:        "clientip",
			TLSHeader:             "tls",
			TLSHeaderValue:        "tls-true",
		},
		Registry: Registry{
			Backend: "something",
			File: File{
				Path: "/foo/bar",
			},
			Static: Static{
				Routes: "route add svc / http://127.0.0.1:6666/",
			},
			Consul: Consul{
				Addr:          "1.2.3.4:5678",
				Token:         "consul-token",
				KVPath:        "/some/path",
				TagPrefix:     "p-",
				ServiceAddr:   "6.6.6.6:7777",
				ServiceName:   "fab",
				ServiceTags:   []string{"a", "b", "c"},
				CheckInterval: 5 * time.Second,
				CheckTimeout:  10 * time.Second,
			},
		},
		Listen: []Listen{
			{
				Addr:         ":1234",
				ReadTimeout:  5 * time.Second,
				WriteTimeout: 10 * time.Second,
			},
		},
		Metrics: []Metrics{
			{
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
			Addr:  "7.8.9.0:1234",
			Color: "fonzy",
			Title: "fabfab",
		},
	}

	p, err := properties.Load([]byte(in), properties.UTF8)
	if err != nil {
		t.Fatalf("got %v want nil", err)
	}

	cfg, err := fromProperties(p)
	if err != nil {
		t.Fatalf("got %v want nil", err)
	}

	got, want := cfg, out
	verify.Values(t, "cfg", got, want)
}

func TestStringVal(t *testing.T) {
	props := func(s string) *properties.Properties {
		p, err := properties.Load([]byte(s), properties.UTF8)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}

	tests := []struct {
		env   map[string]string
		props *properties.Properties
		keys  []string
		val   string
		def   string
	}{
		{
			env:   nil,
			props: nil,
			keys:  []string{"key"}, val: "default", def: "default",
		},
		{
			env:   map[string]string{"key": "env"},
			props: nil,
			keys:  []string{"key"}, val: "env",
		},
		{
			env:   nil,
			props: props("key=props"),
			keys:  []string{"key"}, val: "props",
		},
		{
			env:   map[string]string{"key": "env"},
			props: props("key=props"),
			keys:  []string{"key"}, val: "env",
		},
		{
			env:   map[string]string{"key": "env"},
			props: props("other=props"),
			keys:  []string{"other"}, val: "props",
		},
		{
			env:   map[string]string{"key": "env"},
			props: props("other=props"),
			keys:  []string{"key", "other"}, val: "env",
		},
	}

	for i, tt := range tests {
		for k, v := range tt.env {
			os.Setenv(k, v)
		}
		if got, want := stringVal(tt.props, tt.def, tt.keys...), tt.val; got != want {
			t.Errorf("%d: got %s want %s", i, got, want)
		}
		for k := range tt.env {
			os.Unsetenv(k)
		}
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
				{Addr: ":123"},
			},
			"",
		},
		{
			":123;cert.pem",
			[]Listen{
				{Addr: ":123", CertFile: "cert.pem", KeyFile: "cert.pem", TLS: true},
			},
			"",
		},
		{
			":123;cert.pem;key.pem",
			[]Listen{
				{Addr: ":123", CertFile: "cert.pem", KeyFile: "key.pem", TLS: true},
			},
			"",
		},
		{
			":123;cert.pem;key.pem;client.pem",
			[]Listen{
				{Addr: ":123", CertFile: "cert.pem", KeyFile: "key.pem", ClientAuthFile: "client.pem", TLS: true},
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
		l, err := parseListen(tt.in, time.Duration(0), time.Duration(0))
		if got, want := err, tt.err; (got != nil || want != "") && got.Error() != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
		if got, want := l, tt.out; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}
