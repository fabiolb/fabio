package tcp

import (
	"bufio"
	"io"
	"log"
	"net"
	"time"

	"github.com/fabiolb/fabio/metrics4"
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
	Conn metrics4.Counter

	// ConnFail counts the failed upstream connection attempts.
	ConnFail metrics4.Counter

	// Noroute counts the failed Lookup() calls.
	Noroute metrics4.Counter

	// Metrics is the configured metrics backend provider.
	Metrics metrics4.Provider
}

func (p *SNIProxy) ServeTCP(in net.Conn) error {
	defer in.Close()

	metrics := p.Metrics
	if metrics == nil {
		metrics = &metrics4.MultiProvider{}
	}

	if p.Conn != nil {
		p.Conn.Add(1)
	}

	tlsReader := bufio.NewReader(in)
	tlsHeaders, err := tlsReader.Peek(9)
	if err != nil {
		log.Print("[DEBUG] tcp+sni: TLS handshake failed (failed to peek data)")
		if p.ConnFail != nil {
			p.ConnFail.Add(1)
		}
		return err
	}

	bufferSize, err := clientHelloBufferSize(tlsHeaders)
	if err != nil {
		log.Printf("[DEBUG] tcp+sni: TLS handshake failed (%s)", err)
		if p.ConnFail != nil {
			p.ConnFail.Add(1)
		}
		return err
	}

	data := make([]byte, bufferSize)
	_, err = io.ReadFull(tlsReader, data)
	if err != nil {
		log.Printf("[DEBUG] tcp+sni: TLS handshake failed (%s)", err)
		if p.ConnFail != nil {
			p.ConnFail.Add(1)
		}
		return err
	}

	// readServerName wants only the handshake message so ignore the first
	// 5 bytes which is the TLS record header
	host, ok := readServerName(data[5:])
	if !ok {
		log.Print("[DEBUG] tcp+sni: TLS handshake failed (unable to parse client hello)")
		if p.ConnFail != nil {
			p.ConnFail.Add(1)
		}
		return nil
	}

	if host == "" {
		log.Print("[DEBUG] tcp+sni: server_name missing")
		if p.ConnFail != nil {
			p.ConnFail.Add(1)
		}
		return nil
	}

	t := p.Lookup(host)
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
		log.Print("[WARN] tcp+sni: cannot connect to upstream ", addr)
		if p.ConnFail != nil {
			p.ConnFail.Add(1)
		}
		return err
	}
	defer out.Close()

	// write the data already read from the connection
	n, err := out.Write(data)
	if err != nil {
		log.Print("[WARN] tcp+sni: copy client hello failed. ", err)
		if p.ConnFail != nil {
			p.ConnFail.Add(1)
		}
		return err
	}

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader, c metrics4.Counter) {
		errc <- copyBuffer(dst, src, c)
	}

	// rx measures the traffic to the upstream server (in <- out)
	// tx measures the traffic from the upstream server (out <- in)
	rx := metrics.NewCounter(t.TimerName.String() + ".rx")
	tx := metrics.NewCounter(t.TimerName.String() + ".tx")

	// we've received the ClientHello already
	rx.Add(float64(n))

	go cp(in, out, rx)
	go cp(out, in, tx)
	err = <-errc
	if err != nil && err != io.EOF {
		log.Print("[WARN]: tcp+sni:  ", err)
		return err
	}
	return nil
}
