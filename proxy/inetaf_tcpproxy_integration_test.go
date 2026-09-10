package proxy

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/proxy/internal"
	"github.com/fabiolb/fabio/proxy/tcp"
	"github.com/fabiolb/fabio/proxy/tcp/tcptest"
	"github.com/fabiolb/fabio/route"
)

// to run this test, add the following to /etc/hosts:
// 127.0.0.1	example.com
// 127.0.0.1	example2.com
// and then set the environment FABIO_IHAVEHOSTENTRIES=true
// This also runs in Github Actions by default, since the workflow adds these aliases.
func TestProxyTCPAndHTTPS(t *testing.T) {
	if os.Getenv("TRAVIS") != "true" &&
		os.Getenv("CI") != "true" &&
		os.Getenv("FABIO_IHAVEHOSTENTRIES") != "true" {
		t.Skip("skipping because env FABIO_IHAVEHOSTENTRIES is not set to true")
	}

	tlsCfg1 := tlsServerConfig()
	tlsCfg2 := tlsServerConfig2()
	tcpServer := httptest.NewUnstartedServer(okHandler)
	tcpServer.TLS = tlsCfg2
	tcpServer.StartTLS()
	defer tcpServer.Close()

	httpPayload := []byte(`OK HTTP`)

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(httpPayload)
	}))
	defer httpServer.Close()

	tpl := `route add srv / %s opts "proto=https"
route add tcproute example2.com/ tcp://%s opts "proto=tcp"`

	table, _ := route.NewTable(bytes.NewBufferString(fmt.Sprintf(tpl, httpServer.URL, tcpServer.Listener.Addr())))
	hp := &HTTPProxy{
		Lookup: func(r *http.Request) *route.Target {
			return table.Lookup(r, route.Picker["rr"], route.Matcher["prefix"], globCache, globEnabled)
		},
	}

	tp := &tcp.SNIProxy{
		Lookup: func(h string) *route.Target {
			return table.LookupHost(h, route.Picker["rr"])
		},
	}
	m := func(_ context.Context, h string) bool {
		// TODO - matcher needs to move out of main
		// so we can test it more easily.  Probably
		// the other functions too.
		t := table.LookupHost(h, route.Picker["rr"])
		if t == nil {
			return false
		}
		// Make sure this is supposed to be a tcp proxy.
		// opts proto= overrides scheme if present.
		var (
			ok    bool
			proto string
		)
		if proto, ok = t.Opts["proto"]; !ok && t.URL != nil {
			proto = t.URL.Scheme
		}
		return proto == "tcp"
	}

	// get an unused port for use for the proxy.  the rest of the tests just
	// pick a high-numbered port, but this should be safer, if ugly.  could
	// also just fire up a listener with 0 as the port and let the stack
	// pick one - which is what httptest does - but this is less lines and
	// I'm lazy. --NJ
	tmp := httptest.NewServer(okHandler)
	proxyAddr := tmp.Listener.Addr().String()
	tmp.Close()
	_, port, err := net.SplitHostPort(proxyAddr)
	if err != nil {
		t.Fatalf("error determining port from addr: %s", err)
	}

	l := config.Listen{Addr: proxyAddr}
	go func() {
		err := ListenAndServeHTTPSTCPSNI(l, hp, tp, tlsCfg1, m)
		if err != nil {
			t.Logf("error shutting down: %s", err)
		}
	}()
	defer Close()
	// retry until listener is responding.
	d, err := tcptest.NewRetryDialer().Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("error connecting to proxy: %s", err)
	}
	d.Close()
	// At this point, the proxy should up and listening and will do
	// tcp proxy to https://example2.com, and terminate TLS for
	// https://example.com

	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   tlsClientConfig(),
			DisableKeepAlives: true,
			MaxConnsPerHost:   -1,
		},
	}

	// make sure tcp steering happens for https://example2.com/
	// and https proxying happens for https://example.com/
	for _, data := range []struct {
		name string
		u    string
		h    string
		body []byte
	}{{
		name: "https proxy for example.com",
		u:    "https://example.com:" + port,
		h:    "example.com",
		body: httpPayload,
	}, {
		name: "tcp proxy for example2.com serving https",
		u:    "https://example2.com:" + port,
		h:    "example2.com",
		body: []byte(`OK`),
	}} {
		t.Run(data.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, data.u, nil)
			if err != nil {
				t.Fatalf("unexpected error creating req: %s", err)
			}
			resp, err := c.Do(req)
			if err != nil {
				t.Errorf("error on request %s", err)
				return
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				t.Errorf("error reading body: %s", err)
				return
			}
			if !bytes.Equal(body, data.body) {
				t.Error("http body not equal")
			}
			if len(resp.TLS.PeerCertificates) != 1 {
				t.Errorf("unexpected peer certs")
				return
			}
			if !foundDNSName(resp.TLS.PeerCertificates[0], data.h) {
				t.Error("wrong certificate returned")
			}
		})
	}

}

// TestProxyHTTPSTCPSNIWithProxyProto verifies that the combined listener
// consumes the PROXY header before tcpproxy inspects the TLS ClientHello.
func TestProxyHTTPSTCPSNIWithProxyProto(t *testing.T) {
	for _, version := range proxyProtoVersions {
		t.Run(version.name, func(t *testing.T) {
			srv := tcptest.NewTLSServerWithProxyProto(proxyHandler)
			defer srv.Close()

			tmp, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("net.Listen: %v", err)
			}
			proxyAddr := tmp.Addr().String()
			tmp.Close()

			tp := &tcp.SNIProxy{
				Lookup: func(string) *route.Target {
					return &route.Target{URL: &url.URL{Host: srv.Addr}, ProxyProto: true}
				},
			}
			matcher := func(_ context.Context, host string) bool {
				return host == "example.com"
			}
			go func() {
				l := config.Listen{Addr: proxyAddr, ProxyProto: true}
				if err := ListenAndServeHTTPSTCPSNI(l, okHandler, tp, tlsServerConfig(), matcher); err != nil {
					t.Logf("ListenAndServeHTTPSTCPSNI: %v", err)
				}
			}()
			defer Close()

			cfg := tlsClientConfig()
			cfg.ServerName = "example.com"
			dialer := tcptest.NewTLSRetryDialer(cfg)
			dialer.ProxyProto = true
			dialer.ProxyProtoVersion = version.version
			out, err := dialer.Dial("tcp", proxyAddr)
			if err != nil {
				t.Fatalf("tls.Dial: %#v", err)
			}
			defer out.Close()

			testProxyProto(t, out)
		})
	}
}

// TestProxyHTTPSWithProxyProtoListenerNoHeader verifies that a plain HTTPS
// connection succeeds when the listener has pxyproto enabled but the client
// does not send a PROXY protocol header. go-proxyproto defaults to the USE
// policy which falls through to normal connection handling when no header is
// detected, so the request must complete successfully.
func TestProxyHTTPSWithProxyProtoListenerNoHeader(t *testing.T) {
	tmp, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	proxyAddr := tmp.Addr().String()
	tmp.Close()

	go func() {
		l := config.Listen{
			Addr:               proxyAddr,
			ProxyProto:         true,
			ProxyHeaderTimeout: 250 * time.Millisecond,
		}
		if err := ListenAndServeHTTPSTCPSNI(l, okHandler, &tcp.SNIProxy{
			Lookup: func(string) *route.Target { return nil },
		}, tlsServerConfig(), func(_ context.Context, _ string) bool { return false }); err != nil {
			t.Logf("ListenAndServeHTTPSTCPSNI: %v", err)
		}
	}()
	defer Close()

	rootCAs := x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(internal.LocalhostCert); !ok {
		t.Fatal("could not parse cert")
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: rootCAs},
			DialContext:     (&net.Dialer{}).DialContext,
		},
	}

	// Retry until the listener is up, then send a plain HTTPS request with no
	// PROXY protocol header.
	var resp *http.Response
	deadline := time.Now().Add(2 * time.Second)
	for {
		resp, err = client.Get("https://" + proxyAddr)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("https GET failed: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Errorf("status: got %d want %d", got, want)
	}
	if got, want := body, []byte("OK"); !bytes.Equal(got, want) {
		t.Errorf("body: got %q want %q", got, want)
	}
}

func foundDNSName(crt *x509.Certificate, dnsName string) bool {
	found := slices.Contains(crt.DNSNames, dnsName)
	return found
}
