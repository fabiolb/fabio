package proxy

import (
	"github.com/fabiolb/fabio/metrics4"
	"io"
	"log"
	"net"
	"net/http"
	"sync/atomic"
)

// conns keeps track of the number of open ws connections
var conns int64

type dialFunc func(network, address string) (net.Conn, error)

// newRawProxy returns an HTTP handler which forwards data between
// an incoming and outgoing TCP connection including the original request.
// This handler establishes a new outgoing connection per request.
func newRawProxy(host string, dial dialFunc, conn metrics4.Gauge) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if conn != nil {
			conn.Set(float64(atomic.AddInt64(&conns, 1)))
			defer func() {
				conn.Set(float64(atomic.AddInt64(&conns, -1)))
			}()
		}

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
