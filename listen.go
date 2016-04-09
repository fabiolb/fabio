package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/armon/go-proxyproto"
	"github.com/eBay/fabio/cert"
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
	srv := &http.Server{
		Handler:      h,
		Addr:         l.Addr,
		ReadTimeout:  l.ReadTimeout,
		WriteTimeout: l.WriteTimeout,
	}

	if l.Scheme == "https" {
		src, err := makeCertSource(l.CertSource)
		if err != nil {
			log.Fatal("[FATAL] ", err)
		}

		srv.TLSConfig, err = cert.TLSConfig(src)
		if err != nil {
			log.Fatal("[FATAL] ", err)
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
		log.Fatal("[FATAL] ", err)
	}
}

func makeCertSource(cfg config.CertSource) (cert.Source, error) {
	switch cfg.Type {
	case "file":
		return cert.FileSource{
			CertFile:       cfg.CertPath,
			KeyFile:        cfg.KeyPath,
			ClientAuthFile: cfg.ClientCAPath,
		}, nil

	case "path":
		return cert.PathSource{
			CertPath:     cfg.CertPath,
			ClientCAPath: cfg.ClientCAPath,
			Refresh:      cfg.Refresh,
		}, nil

	case "http":
		return cert.HTTPSource{
			CertURL:     cfg.CertPath,
			ClientCAURL: cfg.ClientCAPath,
			Refresh:     cfg.Refresh,
		}, nil

	case "consul":
		return cert.ConsulSource{
			CertURL:     cfg.CertPath,
			ClientCAURL: cfg.ClientCAPath,
		}, nil

	case "vault":
		return cert.VaultSource{
			// TODO(fs): configure Addr but not token
			CertPath:     cfg.CertPath,
			ClientCAPath: cfg.ClientCAPath,
			Refresh:      cfg.Refresh,
		}, nil

	default:
		return nil, fmt.Errorf("invalid certificate source %q", cfg.Type)
	}
}

func serve(srv *http.Server) error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}

	ln = &proxyproto.Listener{tcpKeepAliveListener{ln.(*net.TCPListener)}}

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
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
