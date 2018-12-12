package tcp

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/fabiolb/fabio/metrics4"
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

	// Conn counts the number of connections.
	Conn metrics4.Counter

	// ConnFail counts the failed upstream connection attempts.
	ConnFail metrics4.Counter

	// Noroute counts the failed Lookup() calls.
	Noroute metrics4.Counter

	// Metrics is the configured metrics backend provider.
	Metrics metrics4.Provider
}

func (p *Proxy) ServeTCP(in net.Conn) error {
	defer in.Close()

	metrics := p.Metrics
	if metrics == nil {
		metrics = &metrics4.MultiProvider{}
	}

	if p.Conn != nil {
		p.Conn.Add(1)
	}

	_, port, _ := net.SplitHostPort(in.LocalAddr().String())
	port = ":" + port
	t := p.Lookup(port)
	if t == nil {
		if p.Noroute != nil {
			p.Noroute.Add(1)
		}
		return nil
	}
	addr := t.URL.Host

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
	cp := func(dst io.Writer, src io.Reader, c metrics4.Counter) {
		errc <- copyBuffer(dst, src, c)
	}

	// rx measures the traffic to the upstream server (in <- out)
	// tx measures the traffic from the upstream server (out <- in)
	rx := metrics.NewCounter(t.TimerName.String() + ".rx")
	tx := metrics.NewCounter(t.TimerName.String() + ".tx")

	go cp(in, out, rx)
	go cp(out, in, tx)
	err = <-errc
	if err != nil && err != io.EOF {
		log.Print("[WARN]: tcp:  ", err)
		return err
	}
	return nil
}
