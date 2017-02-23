package tcp

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/route"
)

// SNIProxy implements an SNI aware transparent TCP proxy which captures the
// TLS client hello, extracts the host name and uses it for finding the
// upstream server. Then it replays the ClientHello message and copies data
// transparently allowing to route a TLS connection based on the SNI header
// without decrypting it.
type SNIProxy struct {
	// Config is the proxy configuration as provided during startup.
	Config config.Proxy

	// Lookup returns a target host for the given server name.
	// The proxy will panic if this value is nil.
	Lookup func(string) *route.Target
}

func (p *SNIProxy) ServeTCP(in net.Conn) error {
	defer in.Close()

	// capture client hello
	data := make([]byte, 1024)
	n, err := in.Read(data)
	if err != nil {
		return err
	}
	data = data[:n]

	serverName, ok := readServerName(data)
	if !ok {
		fmt.Fprintln(in, "handshake failed")
		log.Print("[DEBUG] tcp+sni: TLS handshake failed")
		return nil
	}

	if serverName == "" {
		fmt.Fprintln(in, "server_name missing")
		log.Print("[DEBUG] tcp+sni: server_name missing")
		return nil
	}

	t := p.Lookup(serverName)
	if t == nil {
		log.Print("[WARN] tcp+sni: No route for ", serverName)
		return nil
	}

	out, err := net.DialTimeout("tcp", t.URL.Host, p.Config.DialTimeout)
	if err != nil {
		log.Print("[WARN] tcp+sni: cannot connect to upstream ", t.URL.Host)
		return err
	}
	defer out.Close()

	// copy client hello
	_, err = out.Write(data)
	if err != nil {
		log.Print("[WARN] tcp+sni: copy client hello failed. ", err)
		return err
	}

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}

	go cp(out, in)
	go cp(in, out)
	err = <-errc
	if err != nil && err != io.EOF {
		log.Print("[WARN]: tcp+sni:  ", err)
		return err
	}
	return nil
}
