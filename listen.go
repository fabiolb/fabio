package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/armon/go-proxyproto"
	"github.com/eBay/fabio/cert"
	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/exit"
	"github.com/eBay/fabio/proxy"
)

var quit = make(chan bool)

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
	srv := &http.Server{
		Handler:      h,
		Addr:         l.Addr,
		ReadTimeout:  l.ReadTimeout,
		WriteTimeout: l.WriteTimeout,
	}

	if l.Scheme == "https" {
		src, err := cert.NewSource(l.CertSource)
		if err != nil {
			exit.Fatal("[FATAL] ", err)
		}

		srv.TLSConfig, err = cert.TLSConfig(src, l.StrictMatch)
		if err != nil {
			exit.Fatal("[FATAL] ", err)
		}
	}

	if srv.TLSConfig != nil {
		log.Printf("[INFO] HTTPS proxy listening on %s", l.Addr)
		if srv.TLSConfig.ClientAuth == tls.RequireAndVerifyClientCert {
			log.Printf("[INFO] Client certificate authentication enabled on %s", l.Addr)
		}
	} else {
		log.Printf("[INFO] HTTP proxy listening on %s", l.Addr)
	}

	if err := serve(srv); err != nil {
		exit.Fatal("[FATAL] ", err)
	}
}

func serve(srv *http.Server) error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		exit.Fatal("[FATAL] ", err)
	}

	ln = &proxyproto.Listener{Listener: tcpKeepAliveListener{ln.(*net.TCPListener)}}

	if srv.TLSConfig != nil {
		ln = tls.NewListener(ln, srv.TLSConfig)
	}

	return srv.Serve(ln)
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
	if err = tc.SetKeepAlive(true); err != nil {
		return
	}
	if err = tc.SetKeepAlivePeriod(3 * time.Minute); err != nil {
		return
	}
	return tc, nil
}
