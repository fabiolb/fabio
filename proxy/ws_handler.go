package proxy

import (
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/fabiolb/fabio/metrics"
)

// conn measures the number of open web socket connections
var conn = metrics.DefaultRegistry.GetCounter("ws.conn")

type dialFunc func(network, address string) (net.Conn, error)

// newWSHandler returns an HTTP handler which forwards data between
// an incoming and outgoing Websocket connection. It checks whether
// the handshake was completed successfully before forwarding data
// between the client and server.
func newWSHandler(host string, dial dialFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn.Inc(1)
		defer func() { conn.Inc(-1) }()

		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "not a hijacker", http.StatusInternalServerError)
			return
		}

		in, _, err := hj.Hijack()
		if err != nil {
			log.Printf("[ERROR] Hijack error for %s. %s", r.URL, err)
			http.Error(w, "hijack error", http.StatusInternalServerError)
			return
		}
		defer in.Close()

		out, err := dial("tcp", host)
		if err != nil {
			log.Printf("[ERROR] WS error for %s. %s", r.URL, err)
			http.Error(w, "error contacting backend server", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		err = r.Write(out)
		if err != nil {
			log.Printf("[ERROR] Error copying request for %s. %s", r.URL, err)
			http.Error(w, "error copying request", http.StatusInternalServerError)
			return
		}

		// read the initial response to check whether we get an HTTP/1.1 101 ... response
		// to determine whether the handshake worked.
		b := make([]byte, 1024)
		if err := out.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			log.Printf("[ERROR] Error setting read timeout for %s: %s", r.URL, err)
			http.Error(w, "error setting read timeout", http.StatusInternalServerError)
			return
		}

		n, err := out.Read(b)
		if err != nil {
			log.Printf("[ERROR] Error reading response for %s: %s", r.URL, err)
			http.Error(w, "error reading response", http.StatusInternalServerError)
			return
		}

		b = b[:n]
		if m, err := in.Write(b); err != nil || n != m {
			log.Printf("[ERROR] Error sending header for %s: %s", r.URL, err)
			http.Error(w, "error sending response", http.StatusInternalServerError)
			return
		}

		if !bytes.HasPrefix(b, []byte("HTTP/1.1 101")) {
			log.Printf("[INFO] WS Upgrade failed for %s", r.URL)
			http.Error(w, "error handling ws upgrade", http.StatusInternalServerError)
			return
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
			log.Printf("[INFO] WS error for %s. %s", r.URL, err)
		}
	})
}
