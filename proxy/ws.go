package proxy

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/eBay/fabio/_third_party/golang.org/x/net/websocket"
)

// newWSProxy returns a websocket handler which forwards
// messages from an incoming WS connection to an outgoing
// WS connection. It builds upon the golang.org/x/net/websocket
// library for both incoming and outgoing connections.
func newWSProxy(t *url.URL) http.Handler {
	return websocket.Handler(func(in *websocket.Conn) {
		defer in.Close()

		r := in.Request()
		targetURL := "ws://" + t.Host + r.RequestURI
		out, err := websocket.Dial(targetURL, "", r.Header.Get("Origin"))
		if err != nil {
			log.Printf("[INFO] WS error for %s. %s", r.URL, err)
			return
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
			log.Printf("[INFO] WS error for %s. %s", r.URL, err)
		}
	})
}
