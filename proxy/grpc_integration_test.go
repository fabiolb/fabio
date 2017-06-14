package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	context "golang.org/x/net/context"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/proxy/internal"
	"github.com/fabiolb/fabio/proxy/internal/echosvc"
	"github.com/fabiolb/fabio/route"
	"github.com/pascaldekloe/goe/verify"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GPRCEchoServer struct {
	sync.Mutex
	Addr string // Addr is the listening address
	URL  string // URL is the service URL
	srv  *grpc.Server
}

func (s *GPRCEchoServer) Send(ctx context.Context, m *echosvc.Msg) (*echosvc.Msg, error) {
	return m, nil
}

func (s *GPRCEchoServer) Start(addr string, cert *tls.Certificate) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	var opts []grpc.ServerOption
	if cert != nil {
		creds := credentials.NewServerTLSFromCert(cert)
		opts = append(opts, grpc.Creds(creds))
	}
	s.Lock()
	s.srv = grpc.NewServer(opts...)
	s.Addr = ln.Addr().String()
	s.URL = "http://"
	if cert != nil {
		s.URL = "https://"
	}
	s.URL += s.Addr + "/echosvc.EchoSvc"
	s.Unlock()
	echosvc.RegisterEchoSvcServer(s.srv, s)
	go s.srv.Serve(ln)
	return nil
}

func (s *GPRCEchoServer) Stop() {
	s.Lock()
	defer s.Unlock()
	if s.srv == nil {
		return
	}
	s.srv.Stop()
}

// TestGRPC starts a gRPC and an HTTP server with TLS
// and an HTTP proxy with TLS which all use and accept
// the same certificate. It then tests connections to
// the gRPC and HTTP server directly and via the proxy.
//
// This test requires go1.9 because of incomplete
// trailer support in the http reverse proxy.
// You can test it with the current tip.
//
// See https://github.com/golang/go/issues/20437
//
func TestGRPC(t *testing.T) {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(internal.LocalhostCert)
	tlsServerConfig := &tls.Config{
		NextProtos:   []string{"h2", "http/1.1"},
		Certificates: []tls.Certificate{internal.LocalhostTLSCert},
	}
	tlsClientConfig := &tls.Config{
		RootCAs:    pool,
		NextProtos: []string{"h2", "http/1.1"},
	}
	http2Transport := &http.Transport{TLSClientConfig: tlsClientConfig}
	http2.ConfigureTransport(http2Transport)
	http2Client := &http.Client{Transport: http2Transport}

	// start a gRPC TLS echo server and bind to a random port.
	grpcServer := &GPRCEchoServer{}
	grpcServer.Start("127.0.0.1:0", &internal.LocalhostTLSCert)
	defer grpcServer.Stop()
	log.Println("gRPC echo server listening on ", grpcServer.URL)

	// start a normal upstream TLS webserver to make sure HTTP/2.0 is
	// working via the proxy
	httpServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Proto)
	}))
	httpServer.TLS = tlsServerConfig
	httpServer.StartTLS()
	defer httpServer.Close()
	log.Println("http/2.0 server listening on ", httpServer.URL)

	// start a TLS proxy which uses the same cert for
	// inbound and outbound connections.
	proxy := httptest.NewUnstartedServer(&HTTPProxy{
		// When NoRouteStatus is not > 0 then the HTTP/2 request fails with
		// "missing status pseudo header"
		Config:    config.Proxy{NoRouteStatus: 999},
		Transport: http2Transport,
		Lookup: func(r *http.Request) *route.Target {
			var x string
			x += "route add grpcsvc /echosvc.EchoSvc/ https://" + grpcServer.Addr + "\n"
			x += "route add httpsvc / " + httpServer.URL + "\n"
			tbl, _ := route.NewTable(x)
			return tbl.Lookup(r, "", route.Picker["rr"], route.Matcher["prefix"])
		},
	})
	proxy.TLS = tlsServerConfig
	proxy.StartTLS()
	defer proxy.Close()
	proxyAddr := proxy.URL[len("https://"):]

	t.Run("https direct", func(t *testing.T) {
		resp, err := http2Client.Get(httpServer.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := resp.StatusCode, 200; got != want {
			t.Fatalf("got response code %d want %d", got, want)
		}
		if got, want := string(body), "HTTP/2.0"; got != want {
			t.Fatalf("got response %q want %q", got, want)
		}
	})
	t.Run("https via proxy", func(t *testing.T) {
		resp, err := http2Client.Get(proxy.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := resp.StatusCode, 200; got != want {
			t.Fatalf("got response code %d want %d", got, want)
		}
		if got, want := string(body), "HTTP/2.0"; got != want {
			t.Fatalf("got response %q want %q", got, want)
		}
	})
	t.Run("gRPC direct", func(t *testing.T) {
		creds := credentials.NewClientTLSFromCert(pool, "")
		conn, err := grpc.Dial(grpcServer.Addr, grpc.WithTransportCredentials(creds))
		if err != nil {
			t.Fatal("client connect failed: ", err)
		}
		defer conn.Close()

		client := echosvc.NewEchoSvcClient(conn)
		msg := &echosvc.Msg{Text: "bla"}
		reply, err := client.Send(context.Background(), msg)
		if err != nil {
			t.Fatal("echosvc.Send failed: ", err)
		}
		if got, want := reply, msg; !verify.Values(t, "", got, want) {
			t.Fail()
		}
	})
	t.Run("gRPC via proxy", func(t *testing.T) {
		creds := credentials.NewClientTLSFromCert(pool, "")
		conn, err := grpc.Dial(proxyAddr, grpc.WithTransportCredentials(creds))
		if err != nil {
			t.Fatal("client connect failed: ", err)
		}
		defer conn.Close()

		client := echosvc.NewEchoSvcClient(conn)
		msg := &echosvc.Msg{Text: "bla"}
		reply, err := client.Send(context.Background(), msg)
		if err != nil {
			t.Fatal("echosvc.Send failed: ", err)
		}
		if got, want := reply, msg; !verify.Values(t, "", got, want) {
			t.Fail()
		}
	})
}
