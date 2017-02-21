package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"net"
	"testing"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/proxy/internal"
	"github.com/eBay/fabio/proxy/tcp"
	"github.com/eBay/fabio/proxy/tcp/tcptest"
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

func TestTCPProxy(t *testing.T) {
	srv := tcptest.NewServer(echoHandler)
	defer srv.Close()

	// start proxy
	proxyAddr := "127.0.0.1:57778"
	go func() {
		h := &tcp.Proxy{
			Lookup: func(string) string { return srv.Addr },
		}
		l := config.Listen{Addr: proxyAddr}
		if err := ListenAndServeTCP(l, h); err != nil {
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

func TestTCPSNIProxy(t *testing.T) {
	srv := tcptest.NewTLSServer(echoHandler)
	defer srv.Close()

	// start tcp proxy
	proxyAddr := "127.0.0.1:57778"
	go func() {
		h := &tcp.SNIProxy{
			Lookup: func(string) string { return srv.Addr },
		}
		l := config.Listen{Addr: proxyAddr}
		if err := ListenAndServeTCP(l, h); err != nil {
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
	out, err := tls.Dial("tcp", proxyAddr, cfg)
	if err != nil {
		t.Fatalf("net.Dial: %#v", err)
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
