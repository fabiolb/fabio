package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fabiolb/fabio/cert"
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/proxy/internal"
	"github.com/fabiolb/fabio/proxy/tcp"
	"github.com/fabiolb/fabio/proxy/tcp/tcptest"
	"github.com/fabiolb/fabio/route"
)

var echoHandler tcp.HandlerFunc = func(c net.Conn) error {
	defer c.Close()
	line, _, err := bufio.NewReader(c).ReadLine()
	if err != nil {
		return err
	}
	line = append(line, []byte(" echo")...)
	_, err = c.Write(line)
	return err
}

// TestTCPProxy tests proxying an unencrypted TCP connection
// to a TCP upstream server.
func TestTCPProxy(t *testing.T) {
	srv := tcptest.NewServer(echoHandler)
	defer srv.Close()

	// start proxy
	proxyAddr := "127.0.0.1:57778"
	go func() {
		h := &tcp.Proxy{
			Lookup: func(h string) *route.Target {
				tbl, _ := route.NewTable(bytes.NewBufferString("route add srv :57778 tcp://" + srv.Addr))
				return tbl.LookupHost(h, route.Picker["rr"])
			},
		}
		l := config.Listen{Addr: proxyAddr}
		if err := ListenAndServeTCP(l, h, nil); err != nil {
			t.Log("ListenAndServeTCP: ", err)
		}
	}()
	defer Close()

	// connect to proxy
	out, err := tcptest.NewRetryDialer().Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("net.Dial: %#v", err)
	}
	defer out.Close()

	testRoundtrip(t, out)
}

// TestTCPProxyWithTLS tests proxying an encrypted TCP connection
// to an unencrypted upstream TCP server. The proxy terminates the
// TLS connection.
func TestTCPProxyWithTLS(t *testing.T) {
	srv := tcptest.NewServer(echoHandler)
	defer srv.Close()

	// setup cert source
	dir, err := ioutil.TempDir("", "fabio")
	if err != nil {
		t.Fatal("ioutil.TempDir", err)
	}
	defer os.RemoveAll(dir)

	mustWrite := func(name string, data []byte) {
		path := filepath.Join(dir, name)
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			t.Fatalf("ioutil.WriteFile: %s", err)
		}
	}
	mustWrite("example.com-key.pem", internal.LocalhostKey)
	mustWrite("example.com-cert.pem", internal.LocalhostCert)

	// start tcp proxy
	proxyAddr := "127.0.0.1:57779"
	go func() {
		cs := config.CertSource{Name: "cs", Type: "path", CertPath: dir}
		src, err := cert.NewSource(cs)
		if err != nil {
			t.Fatal("cert.NewSource: ", err)
		}
		cfg, err := cert.TLSConfig(src, false, 0, 0, nil)
		if err != nil {
			t.Fatal("cert.TLSConfig: ", err)
		}

		h := &tcp.Proxy{
			Lookup: func(string) *route.Target {
				return &route.Target{URL: &url.URL{Host: srv.Addr}}
			},
		}

		l := config.Listen{Addr: proxyAddr}
		if err := ListenAndServeTCP(l, h, cfg); err != nil {
			// closing the listener returns this error from the accept loop
			// which we can ignore.
			if err.Error() != "accept tcp 127.0.0.1:57779: use of closed network connection" {
				t.Log("ListenAndServeTCP: ", err)
			}
		}
	}()
	defer Close()

	// give cert store some time to pick up certs
	time.Sleep(250 * time.Millisecond)

	rootCAs := x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(internal.LocalhostCert); !ok {
		t.Fatal("could not parse cert")
	}
	cfg := &tls.Config{
		RootCAs:    rootCAs,
		ServerName: "example.com",
	}

	// connect to proxy
	out, err := tcptest.NewTLSRetryDialer(cfg).Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("tls.Dial: %#v", err)
	}
	defer out.Close()

	testRoundtrip(t, out)
}

// TestTCPSNIProxy tests proxying an encrypted TCP connection
// to an upstream TCP service without decrypting the traffic.
// The upstream server terminates the TLS connection.
func TestTCPSNIProxy(t *testing.T) {
	srv := tcptest.NewTLSServer(echoHandler)
	defer srv.Close()

	// start tcp proxy
	proxyAddr := "127.0.0.1:57778"
	go func() {
		h := &tcp.SNIProxy{
			Lookup: func(string) *route.Target {
				return &route.Target{URL: &url.URL{Host: srv.Addr}}
			},
		}
		l := config.Listen{Addr: proxyAddr}
		if err := ListenAndServeTCP(l, h, nil); err != nil {
			t.Log("ListenAndServeTCP: ", err)
		}
	}()
	defer Close()

	rootCAs := x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(internal.LocalhostCert); !ok {
		t.Fatal("could not parse cert")
	}
	cfg := &tls.Config{
		RootCAs:    rootCAs,
		ServerName: "example.com",
	}

	// connect to proxy
	out, err := tcptest.NewTLSRetryDialer(cfg).Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("tls.Dial: %#v", err)
	}
	defer out.Close()

	testRoundtrip(t, out)
}

func testRoundtrip(t *testing.T, c net.Conn) {
	// send data to server
	_, err := c.Write([]byte("foo\n"))
	if err != nil {
		t.Fatal("out.Write: ", err)
	}

	// read response which should be
	// src data + " echo"
	line, _, err := bufio.NewReader(c).ReadLine()
	if err != nil {
		t.Fatal("readLine: ", err)
	}

	// compare
	if got, want := line, []byte("foo echo"); !bytes.Equal(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

var proxyHandler tcp.HandlerFunc = func(c net.Conn) error {
	defer c.Close()
	line, _, err := bufio.NewReader(c).ReadLine()
	if err != nil {
		return err
	}

	str := " " + c.RemoteAddr().String()
	line = append(line, []byte(str)...)
	_, err = c.Write(line)
	return err
}

// TestTCPProxyWithProxyProtoEnables tests proxying an unencrypted TCP connection
// to a TCP upstream server with proxy protocol enabed on upstream connection
func TestTCPProxyWithProxyProto(t *testing.T) {
	srv := tcptest.NewServerWithProxyProto(proxyHandler)
	defer srv.Close()

	// start proxy
	proxyAddr := "127.0.0.1:57778"
	go func() {
		h := &tcp.Proxy{
			Lookup: func(h string) *route.Target {
				tbl, _ := route.NewTable(bytes.NewBufferString("route add srv :57778 tcp://" + srv.Addr + " opts \"pxyproto=true\""))
				tgt := tbl.LookupHost(h, route.Picker["rr"])
				return tgt
			},
		}
		l := config.Listen{Addr: proxyAddr, ProxyProto: true}
		if err := ListenAndServeTCP(l, h, nil); err != nil {
			t.Log("ListenAndServeTCP: ", err)
		}
	}()
	defer Close()

	// connect to proxy
	dialer := tcptest.NewRetryDialer()
	dialer.ProxyProto = true
	out, err := dialer.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("net.Dial: %#v", err)
	}
	defer out.Close()

	testProxyProto(t, out)
}

// TestTCPProxyWithTLSWithProxyProto tests proxying an encrypted TCP connection
// to an unencrypted upstream TCP server with proxy protocol enabled.
// The proxy extract the proxy protocl header and terminates the TLS connection.
func TestTCPProxyWithTLSWithProxyProto(t *testing.T) {
	srv := tcptest.NewServerWithProxyProto(proxyHandler)
	defer srv.Close()

	// setup cert source
	dir, err := ioutil.TempDir("", "fabio")
	if err != nil {
		t.Fatal("ioutil.TempDir", err)
	}
	defer os.RemoveAll(dir)

	mustWrite := func(name string, data []byte) {
		path := filepath.Join(dir, name)
		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			t.Fatalf("ioutil.WriteFile: %s", err)
		}
	}
	mustWrite("example.com-key.pem", internal.LocalhostKey)
	mustWrite("example.com-cert.pem", internal.LocalhostCert)

	// start tcp proxy
	proxyAddr := "127.0.0.1:57779"
	go func() {
		cs := config.CertSource{Name: "cs", Type: "path", CertPath: dir}
		src, err := cert.NewSource(cs)
		if err != nil {
			t.Fatal("cert.NewSource: ", err)
		}
		cfg, err := cert.TLSConfig(src, false, 0, 0, nil)
		if err != nil {
			t.Fatal("cert.TLSConfig: ", err)
		}

		h := &tcp.Proxy{
			Lookup: func(string) *route.Target {
				return &route.Target{URL: &url.URL{Host: srv.Addr}, ProxyProto: true}
			},
		}

		l := config.Listen{Addr: proxyAddr, ProxyProto: true}
		if err := ListenAndServeTCP(l, h, cfg); err != nil {
			// closing the listener returns this error from the accept loop
			// which we can ignore.
			if err.Error() != "accept tcp 127.0.0.1:57779: use of closed network connection" {
				t.Log("ListenAndServeTCP: ", err)
			}
		}
	}()
	defer Close()

	// give cert store some time to pick up certs
	time.Sleep(250 * time.Millisecond)

	rootCAs := x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(internal.LocalhostCert); !ok {
		t.Fatal("could not parse cert")
	}
	cfg := &tls.Config{
		RootCAs:    rootCAs,
		ServerName: "example.com",
	}

	// connect to proxy
	dialer := tcptest.NewTLSRetryDialer(cfg)
	dialer.ProxyProto = true
	out, err := dialer.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("tls.Dial: %#v", err)
	}
	defer out.Close()

	testProxyProto(t, out)
}

// TestTCPSNIProxyWithProxyProto tests proxying an encrypted TCP connection adding
// proxy protocol header to an upstream TCP service without decrypting the traffic.
// The upstream server extracts the proxy protocol and terminates the TLS connection.
func TestTCPSNIProxyWithProxyProto(t *testing.T) {
	srv := tcptest.NewTLSServerWithProxyProto(proxyHandler)
	defer srv.Close()

	// start tcp proxy
	proxyAddr := "127.0.0.1:57778"
	go func() {
		h := &tcp.SNIProxy{
			Lookup: func(string) *route.Target {
				return &route.Target{URL: &url.URL{Host: srv.Addr}, ProxyProto: true}
			},
		}
		l := config.Listen{Addr: proxyAddr, ProxyProto: true}
		if err := ListenAndServeTCP(l, h, nil); err != nil {
			t.Log("ListenAndServeTCP: ", err)
		}
	}()
	defer Close()

	rootCAs := x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(internal.LocalhostCert); !ok {
		t.Fatal("could not parse cert")
	}
	cfg := &tls.Config{
		RootCAs:    rootCAs,
		ServerName: "example.com",
	}

	// connect to proxy
	dialer := tcptest.NewTLSRetryDialer(cfg)
	dialer.ProxyProto = true
	out, err := dialer.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("tls.Dial: %#v", err)
	}
	defer out.Close()

	testProxyProto(t, out)
}

func testProxyProto(t *testing.T, c net.Conn) {
	// send data to server
	_, err := c.Write([]byte("foo\n"))
	if err != nil {
		t.Fatal("out.Write: ", err)
	}

	// read response which should be
	// PROXY proto header
	line, _, err := bufio.NewReader(c).ReadLine()
	if err != nil {
		t.Fatal("readLine: ", err)
	}

	// remote := c.RemoteAddr().String()
	// local := c.LocalAddr().String()

	// compare
	if got, want := line, []byte("foo 1.2.3.4:12345"); !bytes.Equal(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}
