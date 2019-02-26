package tcptest

import (
	"crypto/tls"
	"net"
	"time"
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
	Dialer     net.Dialer
	Timeout    time.Duration
	Sleep      time.Duration
	ProxyProto bool
}

func (d *RetryDialer) Dial(network, addr string) (c net.Conn, err error) {
	dial := func() (net.Conn, error) {
		conn, err := d.Dialer.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		if d.ProxyProto {
			pxy := "PROXY TCP4 1.2.3.4 5.6.7.8 12345 54321\r\n"
			conn.Write([]byte(pxy))
		}
		return conn, err
	}
	return retry(dial, d.Timeout, d.Sleep)
}

func NewTLSRetryDialer(cfg *tls.Config) *TLSRetryDialer {
	return &TLSRetryDialer{TLS: cfg}
}

type TLSRetryDialer struct {
	TLS        *tls.Config
	Dialer     net.Dialer
	Timeout    time.Duration
	Sleep      time.Duration
	ProxyProto bool
}

func (d *TLSRetryDialer) Dial(network, addr string) (c net.Conn, err error) {
	dial := func() (net.Conn, error) {
		conn, err := net.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		if d.ProxyProto {
			pxy := "PROXY TCP4 1.2.3.4 5.6.7.8 12345 54321\r\n"
			conn.Write([]byte(pxy))
		}
		return tls.Client(conn, d.TLS), nil
	}
	return retry(dial, d.Timeout, d.Sleep)
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
