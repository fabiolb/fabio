package tcp

import (
	"io"
	"log"
	"net"
	"time"
)

// Proxy implements a generic TCP proxying handler.
type Proxy struct {
	// DialTimeout sets the timeout for establishing the outbound
	// connection.
	DialTimeout time.Duration

	// Lookup returns a target host for the given server name.
	// The proxy will panic if this value is nil.
	Lookup func(host string) string
}

func (p *Proxy) ServeTCP(in net.Conn) error {
	defer in.Close()

	_, port, _ := net.SplitHostPort(in.LocalAddr().String())
	port = ":" + port
	addr := p.Lookup(port)
	if addr == "" {
		return nil
	}

	out, err := net.DialTimeout("tcp", addr, p.DialTimeout)
	if err != nil {
		log.Print("[WARN] tcp: cannot connect to upstream ", addr)
		return err
	}
	defer out.Close()

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}

	go cp(out, in)
	go cp(in, out)
	err = <-errc
	if err != nil && err != io.EOF {
		log.Print("[WARN]: tcp:  ", err)
		return err
	}
	return nil
}
