package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/exit"
	"github.com/eBay/fabio/route"
)

var quit = make(chan bool)
var commas = regexp.MustCompile(`\s*,\s*`)
var semicolons = regexp.MustCompile(`\s*;\s*`)

func init() {
	exit.Listen(func(os.Signal) { close(quit) })
}

// listen starts one or more listeners for the handler. The list
// of addresses are
func listen(cfg []config.Listen, wait time.Duration, h http.Handler) {
	for _, l := range cfg {
		if l.TLS {
			go listenAndServeTLS(l, h)
		} else {
			go listenAndServe(l, h)
		}
	}

	// wait for shutdown signal
	<-quit

	// disable routing for all requests
	route.Shutdown()

	// trigger graceful shutdown
	log.Printf("[INFO] Graceful shutdown over %s", wait)
	time.Sleep(wait)
	log.Print("[INFO] Down")
}

func listenAndServe(l config.Listen, h http.Handler) {
	log.Printf("[INFO] HTTP proxy listening on %s", l.Addr)
	srv := &http.Server{
		Addr:         l.Addr,
		Handler:      h,
		ReadTimeout:  l.ReadTimeout,
		WriteTimeout: l.WriteTimeout,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("[FATAL] ", err)
	}
}

// listenAndServeTLS starts an HTTPS server with the given certificate.
func listenAndServeTLS(l config.Listen, h http.Handler) {
	log.Printf("[INFO] HTTPS proxy listening on %s with certificate %s", l.Addr, l.CertFile)
	cert, err := tls.LoadX509KeyPair(l.CertFile, l.KeyFile)
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}

	srv := &http.Server{
		Addr:         l.Addr,
		Handler:      h,
		ReadTimeout:  l.ReadTimeout,
		WriteTimeout: l.WriteTimeout,
		TLSConfig: &tls.Config{
			NextProtos:   []string{"http/1.1"},
			Certificates: []tls.Certificate{cert},
		},
	}

	ln, err := net.Listen("tcp", l.Addr)
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}

	tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, srv.TLSConfig)
	if err := srv.Serve(tlsListener); err != nil {
		log.Fatal("[FATAL] ", err)
	}
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
