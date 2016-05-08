package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/armon/go-proxyproto"
	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/exit"
	"github.com/eBay/fabio/proxy"
)

var quit = make(chan bool)
var commas = regexp.MustCompile(`\s*,\s*`)
var semicolons = regexp.MustCompile(`\s*;\s*`)

func init() {
	exit.Listen(func(os.Signal) { close(quit) })
}

// startListeners runs one or more listeners for the handler
func startListeners(listen []config.Listen, wait time.Duration, h http.Handler) {
	for _, l := range listen {
		go listenAndServe(l, h)
	}

	// wait for shutdown signal
	<-quit

	// disable routing for all requests
	proxy.Shutdown()

	// trigger graceful shutdown
	log.Printf("[INFO] Graceful shutdown over %s", wait)
	time.Sleep(wait)
	log.Print("[INFO] Down")
}

func listenAndServe(l config.Listen, h http.Handler) {
	srv, err := newServer(l, h)
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}

	if srv.TLSConfig != nil {
		log.Printf("[INFO] HTTPS proxy listening on %s with certificate %s", l.Addr, l.CertFile)
		if srv.TLSConfig.ClientAuth == tls.RequireAndVerifyClientCert {
			log.Printf("[INFO] Client certificate authentication enabled on %s with certificates from %s", l.Addr, l.ClientAuthFile)
		}
	} else {
		log.Printf("[INFO] HTTP proxy listening on %s", l.Addr)
	}

	if err := serve(srv); err != nil {
		log.Fatal("[FATAL] ", err)
	}
}

var tlsLoadX509KeyPair = tls.LoadX509KeyPair

func newServer(l config.Listen, h http.Handler) (*http.Server, error) {
	srv := &http.Server{
		Addr:         l.Addr,
		Handler:      h,
		ReadTimeout:  l.ReadTimeout,
		WriteTimeout: l.WriteTimeout,
	}

	if l.CertFile != "" {
		cert, err := tlsLoadX509KeyPair(l.CertFile, l.KeyFile)
		if err != nil {
			return nil, err
		}

		srv.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		if l.ClientAuthFile != "" {
			pemBlock, err := ioutil.ReadFile(l.ClientAuthFile)
			if err != nil {
				return nil, err
			}
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM(pemBlock) {
				return nil, errors.New("failed to add client auth certs")
			}
			srv.TLSConfig.ClientCAs = pool
			srv.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
	}

	return srv, nil
}

func serve(srv *http.Server) error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}

	ln = tcpKeepAliveListener{ln.(*net.TCPListener)}

	if srv.TLSConfig != nil {
		ln = tls.NewListener(ln, srv.TLSConfig)
	}

	return srv.Serve(&proxyproto.Listener{ln})
}

// copied from http://golang.org/src/net/http/server.go?s=54604:54695#L1967
// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
