package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/eBay/fabio/route"
)

var quit = make(chan bool)
var commas = regexp.MustCompile(`\s*,\s*`)
var semicolons = regexp.MustCompile(`\s*;\s*`)

func init() {
	go func() {
		// we use buffered to mitigate losing the signal
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-sigchan
		close(quit)
	}()
}

// listen starts one or more listeners for the handler. The list
// of addresses are
func listen(addrs string, wait time.Duration, h http.Handler) {
	for _, addr := range commas.Split(addrs, -1) {
		if addr == "" {
			continue
		}

		p := semicolons.Split(addr, 4)
		switch len(p) {
		case 1:
			go listenAndServe(p[0], h)
		case 2:
			go listenAndServeTLS(p[0], p[1], p[1], h)
		case 3:
			go listenAndServeTLS(p[0], p[1], p[2], h)
		default:
			log.Fatal("[FATAL] Invalid address format ", addr)
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

func listenAndServe(addr string, h http.Handler) {
	log.Printf("[INFO] HTTP proxy listening on %s", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatal("[FATAL] ", err)
	}
}

// listenAndServeTLS starts an HTTPS server with the given certificate.
func listenAndServeTLS(addr, certFile, keyFile string, h http.Handler) {
	log.Printf("[INFO] HTTPS proxy listening on %s with certificate %s", addr, certFile)
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}

	config := &tls.Config{
		NextProtos:   []string{"http/1.1"},
		Certificates: []tls.Certificate{cert},
	}
	srv := &http.Server{Addr: addr, TLSConfig: config, Handler: h}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}

	tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, config)
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
