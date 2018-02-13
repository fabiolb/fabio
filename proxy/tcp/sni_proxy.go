package tcp

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/fabiolb/fabio/metrics"
	"github.com/fabiolb/fabio/route"
)

// SNIProxy implements an SNI aware transparent TCP proxy which captures the
// TLS client hello, extracts the host name and uses it for finding the
// upstream server. Then it replays the ClientHello message and copies data
// transparently allowing to route a TLS connection based on the SNI header
// without decrypting it.
type SNIProxy struct {
	// DialTimeout sets the timeout for establishing the outbound
	// connection.
	DialTimeout time.Duration

	// Lookup returns a target host for the given server name.
	// The proxy will panic if this value is nil.
	Lookup func(host string) *route.Target

	// Conn counts the number of connections.
	Conn metrics.Counter

	// ConnFail counts the failed upstream connection attempts.
	ConnFail metrics.Counter

	// Noroute counts the failed Lookup() calls.
	Noroute metrics.Counter
}

func (p *SNIProxy) ServeTCP(in net.Conn) error {
	defer in.Close()

	if p.Conn != nil {
		p.Conn.Inc(1)
	}

	// capture client hello
	data := make([]byte, 1024)
	n, err := in.Read(data)
	if err != nil {
		if p.ConnFail != nil {
			p.ConnFail.Inc(1)
		}
		return err
	}
	data = data[:n]

	host, ok := readServerName(data)
	if !ok {
		log.Print("[DEBUG] tcp+sni: TLS handshake failed")
		if p.ConnFail != nil {
			p.ConnFail.Inc(1)
		}
		return nil
	}

	if host == "" {
		log.Print("[DEBUG] tcp+sni: server_name missing")
		if p.ConnFail != nil {
			p.ConnFail.Inc(1)
		}
		return nil
	}

	t := p.Lookup(host)
	if t == nil {
		if p.Noroute != nil {
			p.Noroute.Inc(1)
		}
		return nil
	}
	addr := t.URL.Host

	if t.AccessDeniedTCP(in) {
		log.Print("[INFO] route rules denied access to ", t.URL.String())
		return nil
	}

	out, err := net.DialTimeout("tcp", addr, p.DialTimeout)
	if err != nil {
		log.Print("[WARN] tcp+sni: cannot connect to upstream ", addr)
		if p.ConnFail != nil {
			p.ConnFail.Inc(1)
		}
		return err
	}
	defer out.Close()

	// copy client hello
	n, err = out.Write(data)
	if err != nil {
		log.Print("[WARN] tcp+sni: copy client hello failed. ", err)
		if p.ConnFail != nil {
			p.ConnFail.Inc(1)
		}
		return err
	}

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader, c metrics.Counter) {
		errc <- copyBuffer(dst, src, c)
	}

	// rx measures the traffic to the upstream server (in <- out)
	// tx measures the traffic from the upstream server (out <- in)
	rx := metrics.DefaultRegistry.GetCounter(t.TimerName + ".rx")
	tx := metrics.DefaultRegistry.GetCounter(t.TimerName + ".tx")

	// we've received the ClientHello already
	rx.Inc(int64(n))

	go cp(in, out, rx)
	go cp(out, in, tx)
	err = <-errc
	if err != nil && err != io.EOF {
		log.Print("[WARN]: tcp+sni:  ", err)
		return err
	}
	return nil
}
