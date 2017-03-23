package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	proxyproto "github.com/armon/go-proxyproto"
)

func ListenTCP(laddr string, cfg *tls.Config) (net.Listener, error) {
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
