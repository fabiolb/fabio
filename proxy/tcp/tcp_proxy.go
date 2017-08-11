package tcp

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/fabiolb/fabio/metrics"
	"github.com/fabiolb/fabio/route"
)

// Proxy implements a generic TCP proxying handler.
type Proxy struct {
	// DialTimeout sets the timeout for establishing the outbound
	// connection.
	DialTimeout time.Duration

	// Lookup returns a target host for the given request.
	// The proxy will panic if this value is nil.
	Lookup func(host string) *route.Target

	// Conn is the metric name which counts the number of connections.
	Conn string

	// ConnFail is the metric name which counts failed upstream connection attempts.
	ConnFail string

	// Noroute is the metric name which counts failed Lookup() calls.
	Noroute string
}

func (p *Proxy) ServeTCP(in net.Conn) error {
	defer in.Close()

	metrics.IncDefault(p.Conn, 1)

	_, port, _ := net.SplitHostPort(in.LocalAddr().String())
	port = ":" + port
	t := p.Lookup(port)
	if t == nil {
		metrics.IncDefault(p.Noroute, 1)
		return nil
	}
	addr := t.URL.Host

	out, err := net.DialTimeout("tcp", addr, p.DialTimeout)
	if err != nil {
		log.Print("[WARN] tcp: cannot connect to upstream ", addr)
		metrics.IncDefault(p.ConnFail, 1)
		return err
	}
	defer out.Close()

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader, name string) {
		errc <- copyBuffer(dst, src, name)
	}

	// rx measures the traffic to the upstream server (in <- out)
	// tx measures the traffic from the upstream server (out <- in)
	rx := t.TimerName + ".rx"
	tx := t.TimerName + ".tx"

	go cp(in, out, rx)
	go cp(out, in, tx)
	err = <-errc
	if err != nil && err != io.EOF {
		log.Print("[WARN]: tcp:  ", err)
		return err
	}
	return nil
}
