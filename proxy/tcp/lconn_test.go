package tcp

import (
	"crypto/tls"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/fabiolb/fabio/route"
)

func TestLeastConnTCPSessions(t *testing.T) {
	for _, name := range []string{"tcp", "dynamic", "sni"} {
		t.Run(name, func(t *testing.T) {
			backend, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer backend.Close()
			target := &route.Target{URL: &url.URL{Host: backend.Addr().String()}}
			accepted := make(chan net.Conn, 1)
			go func() {
				conn, err := backend.Accept()
				if err == nil {
					accepted <- conn
				}
			}()

			in, server := net.Pipe()
			defer in.Close()
			var handler Handler
			switch name {
			case "tcp":
				handler = &Proxy{ListenAddr: ":12345", Lookup: func(string) *route.Target { return target }}
			case "dynamic":
				handler = &DynamicProxy{Lookup: func([]string) *route.Target { return target }}
			case "sni":
				handler = &SNIProxy{Lookup: func(string) *route.Target { return target }}
			}
			finished := make(chan error, 1)
			go func() { finished <- handler.ServeTCP(server) }()
			var handshakeDone chan error
			if name == "sni" {
				handshakeDone = make(chan error, 1)
				go func() {
					client := tls.Client(in, &tls.Config{ServerName: "example.com", InsecureSkipVerify: true})
					handshakeDone <- client.Handshake() // the raw backend deliberately does not finish TLS
				}()
			}

			var out net.Conn
			select {
			case out = <-accepted:
			case err := <-finished:
				t.Fatalf("proxy returned before upstream connection: %v", err)
			case <-time.After(2 * time.Second):
				t.Fatal("upstream connection not established")
			}
			defer out.Close()
			if got := target.Active(); got != 1 {
				t.Fatalf("connected target has %d active sessions, want 1", got)
			}
			in.Close()
			if handshakeDone != nil {
				select {
				case <-handshakeDone:
				case <-time.After(2 * time.Second):
					t.Fatal("TLS client did not finish after close")
				}
			}
			select {
			case <-finished:
			case <-time.After(2 * time.Second):
				t.Fatal("proxy did not finish after client closed")
			}
			if got := target.Active(); got != 0 {
				t.Fatalf("closed session left %d active operations", got)
			}
		})
	}
}

func TestLeastConnTCPFailures(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	target := &route.Target{URL: &url.URL{Host: addr}}
	for _, name := range []string{"tcp", "dynamic"} {
		t.Run(name, func(t *testing.T) {
			in, server := net.Pipe()
			defer in.Close()
			var handler Handler
			switch name {
			case "tcp":
				handler = &Proxy{ListenAddr: ":12345", Lookup: func(string) *route.Target { return target }}
			case "dynamic":
				handler = &DynamicProxy{Lookup: func([]string) *route.Target { return target }}
			}
			if err := handler.ServeTCP(server); err == nil {
				t.Fatal("expected dial error")
			}
			if got := target.Active(); got != 0 {
				t.Fatalf("failed dial left %d active operations", got)
			}
		})
	}
}
