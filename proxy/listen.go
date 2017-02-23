package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/armon/go-proxyproto"
	"github.com/eBay/fabio/cert"
	"github.com/eBay/fabio/config"
)

//func listenAndServeHTTP(l config.Listen, h http.Handler) {
//	ln, err := ListenTCP(l.Addr, l.CertSource, l.StrictMatch)
//	if err != nil {
//		exit.Fatal("[FATAL] ", err)
//	}
//
//	srv := &http.Server{
//		Handler:      h,
//		Addr:         l.Addr,
//		ReadTimeout:  l.ReadTimeout,
//		WriteTimeout: l.WriteTimeout,
//		TLSConfig:    ln.(*tcpListener).cfg,
//	}
//
//	if srv.TLSConfig != nil {
//		log.Printf("[INFO] HTTPS proxy listening on %s", l.Addr)
//		if srv.TLSConfig.ClientAuth == tls.RequireAndVerifyClientCert {
//			log.Printf("[INFO] Client certificate authentication enabled on %s", l.Addr)
//		}
//	} else {
//		log.Printf("[INFO] HTTP proxy listening on %s", l.Addr)
//	}
//
//	if err := srv.Serve(ln); err != nil {
//		exit.Fatal("[FATAL] ", err)
//	}
//}

func ListenTCP(laddr string, cs config.CertSource, strictMatch bool) (net.Listener, error) {
	var cfg *tls.Config
	if cs.Name != "" {
		src, err := cert.NewSource(cs)
		if err != nil {
			return nil, fmt.Errorf("listen: Fail to create cert source. %s", err)
		}
		cfg, err = cert.TLSConfig(src, strictMatch)
		if err != nil {
			return nil, fmt.Errorf("listen: Fail to create TLS config. %s", err)
		}
	}

	addr, err := net.ResolveTCPAddr("tcp", laddr)
	if err != nil {
		return nil, fmt.Errorf("listen: Fail to resolve tcp addr. %s", laddr)
	}

	var ln net.Listener
	ln, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen: Fail to listen. %s", err)
	}

	// enable TCPKeepAlive support
	ln = tcpKeepAliveListener{ln.(*net.TCPListener)}

	// enable PROXY protocol support
	ln = &proxyproto.Listener{Listener: ln}

	// enable TLS
	if cfg != nil {
		ln = tls.NewListener(ln, cfg)
	}

	return &tcpListener{ln, addr, cfg}, nil
}

type tcpListener struct {
	l         net.Listener
	addr      net.Addr
	tlsConfig *tls.Config
}

func (ln *tcpListener) Addr() net.Addr {
	return ln.addr
}

func (ln *tcpListener) Accept() (net.Conn, error) {
	return ln.l.Accept()
}

func (ln *tcpListener) Close() error {
	return ln.l.Close()
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
