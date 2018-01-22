package tcp

import (
	"bufio"
	"errors"
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

// Create a buffer large enough to hold the client hello message including
// the tls record header and the handshake message header.
// The function requires at least the first 9 bytes of the tls conversation
// in "data".
// nil, error is returned if the data does not follow the
// specification (https://tools.ietf.org/html/rfc5246) or if the client hello
// is fragmented over multiple records.
func createClientHelloBuffer(data []byte) ([]byte, error) {
	// TLS record header
	// -----------------
	// byte   0: rec type (should be 0x16 == Handshake)
	// byte 1-2: version (should be 0x3000 < v < 0x3003)
	// byte 3-4: rec len
	if len(data) < 9 {
		return nil, errors.New("At least 9 bytes required to determine client hello length")
	}

	if data[0] != 0x16 {
		return nil, errors.New("Not a TLS handshake")
	}

	recordLength := int(data[3])<<8 | int(data[4])
	if recordLength <= 0 || recordLength > 16384 {
		return nil, errors.New("Invalid TLS record length")
	}

	// Handshake record header
	// -----------------------
	// byte   5: hs msg type (should be 0x01 == client_hello)
	// byte 6-8: hs msg len
	if data[5] != 0x01 {
		return nil, errors.New("Not a client hello")
	}

	handshakeLength := int(data[6])<<16 | int(data[7])<<8 | int(data[8])
	if handshakeLength <= 0 || handshakeLength > recordLength-4 {
		return nil, errors.New("Invalid client hello length (fragmentation not implemented)")
	}

	return make([]byte, handshakeLength+9), nil //9 for the header bytes
}

func (p *SNIProxy) ServeTCP(in net.Conn) error {
	defer in.Close()

	if p.Conn != nil {
		p.Conn.Inc(1)
	}

	tlsReader := bufio.NewReader(in)
	data, err := tlsReader.Peek(9)
	if err != nil {
		log.Print("[DEBUG] tcp+sni: TLS handshake failed (failed to peek data)")
		if p.ConnFail != nil {
			p.ConnFail.Inc(1)
		}
		return err
	}

	tlsData, err := createClientHelloBuffer(data)
	if err != nil {
		log.Printf("[DEBUG] tcp+sni: TLS handshake failed (%s)", err)
		if p.ConnFail != nil {
			p.ConnFail.Inc(1)
		}
		return err
	}

	_, err = io.ReadFull(tlsReader, tlsData)
	if err != nil {
		log.Printf("[DEBUG] tcp+sni: TLS handshake failed (%s)", err)
		if p.ConnFail != nil {
			p.ConnFail.Inc(1)
		}
		return err
	}

	host, ok := readServerName(tlsData[5:])
	if !ok {
		log.Print("[DEBUG] tcp+sni: TLS handshake failed (unable to parse client hello)")
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

	out, err := net.DialTimeout("tcp", addr, p.DialTimeout)
	if err != nil {
		log.Print("[WARN] tcp+sni: cannot connect to upstream ", addr)
		if p.ConnFail != nil {
			p.ConnFail.Inc(1)
		}
		return err
	}
	defer out.Close()

	// write the data already read from the connection
	n, err := out.Write(tlsData)
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
