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

	// Conn is the metric name which counts the number of connections.
	Conn string

	// ConnFail is the metric name which counts failed upstream connection attempts.
	ConnFail string

	// Noroute is the metric name which counts failed Lookup() calls.
	Noroute string
}

func (p *SNIProxy) ServeTCP(in net.Conn) error {
	defer in.Close()

	metrics.IncDefault(p.Conn, 1)

	// capture client hello
	data := make([]byte, 1024)
	n, err := in.Read(data)
	if err != nil {
		metrics.IncDefault(p.ConnFail, 1)
		return err
	}
	data = data[:n]

	host, ok := readServerName(data)
	if !ok {
		log.Print("[DEBUG] tcp+sni: TLS handshake failed")
		metrics.IncDefault(p.ConnFail, 1)
		return nil
	}

	if host == "" {
		log.Print("[DEBUG] tcp+sni: server_name missing")
		metrics.IncDefault(p.ConnFail, 1)
		return nil
	}

	t := p.Lookup(host)
	if t == nil {
		metrics.IncDefault(p.Noroute, 1)
		return nil
	}
	addr := t.URL.Host

	out, err := net.DialTimeout("tcp", addr, p.DialTimeout)
	if err != nil {
		log.Print("[WARN] tcp+sni: cannot connect to upstream ", addr)
		metrics.IncDefault(p.ConnFail, 1)
		return err
	}
	defer out.Close()

	// copy client hello
	n, err = out.Write(data)
	if err != nil {
		log.Print("[WARN] tcp+sni: copy client hello failed. ", err)
		metrics.IncDefault(p.ConnFail, 1)
		return err
	}

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader, metric string) {
		errc <- copyBuffer(dst, src, metric)
	}

	// rx measures the traffic to the upstream server (in <- out)
	// tx measures the traffic from the upstream server (out <- in)
	rx := t.TimerName + ".rx"
	tx := t.TimerName + ".tx"

	// we've received the ClientHello already
	metrics.IncDefault(rx, int64(n))

	go cp(in, out, rx)
	go cp(out, in, tx)
	err = <-errc
	if err != nil && err != io.EOF {
		log.Print("[WARN]: tcp+sni:  ", err)
		return err
	}
	return nil
}
