package proxy

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/route"
)

// TCPProxy implements an SNI aware transparent TCP proxy which captures the
// TLS client hello, extracts the host name and uses it for finding the
// upstream server. Then it replays the ClientHello message and copies data
// transparently allowing to route a TLS connection based on the SNI header
// without decrypting it.
//
// This implementation is EXPERIMENTAL in the sense that it has been tested
// to work but is considered incomplete for production use. It needs support
// for read and write timeouts which require replacing the io.Copy() with
// something that can set the deadlines on the underlying connections. One
// possible way could be to use TeeReader/TeeWriter streams which discard
// the data and only set the deadlines. The implementation also needs a
// full integration test.
//
// This implementation exists to gather more real-world data to finalize
// the code at a later stage.
type TCPProxy interface {
	Serve(conn net.Conn)
}

func NewTCPSNIProxy(cfg config.Proxy) TCPProxy {
	return &tcpSNIProxy{cfg: cfg}
}

type tcpSNIProxy struct {
	cfg config.Proxy
}

func (p *tcpSNIProxy) Serve(in net.Conn) {
	defer in.Close()

	if ShuttingDown() {
		return
	}

	// capture client hello
	data := make([]byte, 1024)
	n, err := in.Read(data)
	if err != nil {
		return
	}
	data = data[:n]

	serverName, ok := readServerName(data)
	if !ok {
		fmt.Fprintln(in, "handshake failed")
		log.Print("[DEBUG] tcp+sni: TLS handshake failed")
		return
	}

	if serverName == "" {
		fmt.Fprintln(in, "server_name missing")
		log.Print("[DEBUG] tcp+sni: server_name missing")
		return
	}

	t := route.GetTable().LookupHost(serverName)
	if t == nil {
		log.Print("[WARN] tcp+sni: No route for ", serverName)
		return
	}

	out, err := net.DialTimeout("tcp", t.URL.Host, p.cfg.DialTimeout)
	if err != nil {
		log.Print("[WARN] tcp+sni: cannot connect to upstream ", t.URL.Host)
		return
	}
	defer out.Close()

	// copy client hello
	_, err = out.Write(data)
	if err != nil {
		log.Print("[WARN] tcp+sni: copy client hello failed. ", err)
		return
	}

	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader) {
		// TODO(fs): this implementation does not enforce any timeouts.
		// for this the io.Copy will have to be replaced with something
		// more sophisticated. Idea: use TeeReader/TeeWriter to discard
		// the second data stream and set the deadlines.
		_, err := io.Copy(dst, src)
		errc <- err
	}

	go cp(out, in)
	go cp(in, out)
	err = <-errc
	if err != nil && err != io.EOF {
		log.Print("[WARN]: tcp+sni:  ", err)
	}
}
