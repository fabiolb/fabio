package tcptest

import (
	"crypto/tls"
	"encoding/binary"
	"io"
	"net"
	"time"
)

// ProxyProtoVersion identifies the PROXY protocol version written by a test
// dialer. A zero value retains the existing v1 behavior.
type ProxyProtoVersion uint8

const (
	ProxyProtoVersion1 ProxyProtoVersion = 1
	ProxyProtoVersion2 ProxyProtoVersion = 2
)

type Dialer interface {
	Dial(network, addr string) (net.Conn, error)
}

func NewRetryDialer() *RetryDialer {
	return &RetryDialer{}
}

// RetryDialer retries the Dial function until it succeeds or
// the timeout has been reached. The default timeout is one
// second and the default sleep interval is 100ms.
type RetryDialer struct {
	Dialer            net.Dialer
	Timeout           time.Duration
	Sleep             time.Duration
	ProxyProto        bool
	ProxyProtoVersion ProxyProtoVersion
}

func (d *RetryDialer) Dial(network, addr string) (c net.Conn, err error) {
	dial := func() (net.Conn, error) {
		conn, err := d.Dialer.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		if d.ProxyProto {
			err = writeProxyProtoHeader(conn, d.ProxyProtoVersion)
		}
		return conn, err
	}
	return retry(dial, d.Timeout, d.Sleep)
}

func NewTLSRetryDialer(cfg *tls.Config) *TLSRetryDialer {
	return &TLSRetryDialer{TLS: cfg}
}

type TLSRetryDialer struct {
	TLS               *tls.Config
	Dialer            net.Dialer
	Timeout           time.Duration
	Sleep             time.Duration
	ProxyProto        bool
	ProxyProtoVersion ProxyProtoVersion
}

func (d *TLSRetryDialer) Dial(network, addr string) (c net.Conn, err error) {
	dial := func() (net.Conn, error) {
		conn, err := d.Dialer.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		if d.ProxyProto {
			err = writeProxyProtoHeader(conn, d.ProxyProtoVersion)
		}
		return tls.Client(conn, d.TLS), err
	}
	return retry(dial, d.Timeout, d.Sleep)
}

func writeProxyProtoHeader(conn net.Conn, version ProxyProtoVersion) error {
	header := proxyProtoHeader(version)
	n, err := conn.Write(header)
	if err != nil {
		return err
	}
	if n != len(header) {
		return io.ErrShortWrite
	}
	return nil
}

func proxyProtoHeader(version ProxyProtoVersion) []byte {
	if version != ProxyProtoVersion2 {
		return []byte("PROXY TCP4 1.2.3.4 5.6.7.8 12345 54321\r\n")
	}

	// This v2 fixture is intentionally assembled without go-proxyproto so the
	// integration tests exercise the parser against the wire representation.
	header := make([]byte, 28)
	copy(header, []byte{
		0x0d, 0x0a, 0x0d, 0x0a, 0x00, 0x0d, 0x0a, 0x51,
		0x55, 0x49, 0x54, 0x0a,
		0x21, 0x11, // version 2, PROXY command, TCP over IPv4
	})
	binary.BigEndian.PutUint16(header[14:16], 12)
	copy(header[16:20], []byte{1, 2, 3, 4})
	copy(header[20:24], []byte{5, 6, 7, 8})
	binary.BigEndian.PutUint16(header[24:26], 12345)
	binary.BigEndian.PutUint16(header[26:28], 54321)
	return header
}

type dialer func() (net.Conn, error)

func retry(dial dialer, timeout, sleep time.Duration) (c net.Conn, err error) {
	if sleep == 0 {
		sleep = 100 * time.Millisecond
	}
	if timeout == 0 {
		timeout = time.Second
	}
	deadline := time.Now().Add(timeout)

	for {
		c, err = dial()
		if err != nil && time.Now().Before(deadline) {
			time.Sleep(sleep)
			continue
		}
		return
	}
}
