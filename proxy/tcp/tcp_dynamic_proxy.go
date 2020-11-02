package tcp

import (
	gkm "github.com/go-kit/kit/metrics"
	"io"
	"log"
	"net"
	"time"

	"github.com/fabiolb/fabio/route"
)

// Proxy implements a generic TCP proxying handler.
type DynamicProxy struct {
	// DialTimeout sets the timeout for establishing the outbound
	// connection.
	DialTimeout time.Duration

	// Lookup returns a target host for the given request.
	// The proxy will panic if this value is nil.
	Lookup func(host string) *route.Target

	// Conn counts the number of connections.
	Conn gkm.Counter

	// ConnFail counts the failed upstream connection attempts.
	ConnFail gkm.Counter

	// Noroute counts the failed Lookup() calls.
	Noroute gkm.Counter
}

func (p *DynamicProxy) ServeTCP(in net.Conn) error {
	defer in.Close()

	if p.Conn != nil {
		p.Conn.Add(1)
	}
	target := in.LocalAddr().String()
	t := p.Lookup(target)
	if t == nil {
		if p.Noroute != nil {
			p.Noroute.Add(1)
		}
		return nil
	}
	addr := t.URL.Host
	log.Printf("[DEBUG]  Connection: %s incoming %s to %s: ", in.RemoteAddr(), target, addr)

	if t.AccessDeniedTCP(in) {
		return nil
	}

	out, err := net.DialTimeout("tcp", addr, p.DialTimeout)
	if err != nil {
		log.Print("[WARN] tcp: cannot connect to upstream ", addr)
		if p.ConnFail != nil {
			p.ConnFail.Add(1)
		}
		return err
	}
	defer out.Close()

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader, c gkm.Counter) {
		errc <- copyBuffer(dst, src, c)
	}

	go cp(in, out, t.RxCounter)
	go cp(out, in, t.TxCounter)
	err = <-errc
	if err != nil && err != io.EOF {
		log.Print("[WARN]: tcp:  ", err)
		return err
	}
	return nil
}
