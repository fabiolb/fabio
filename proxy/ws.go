package proxy

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/eBay/fabio/_third_party/golang.org/x/net/websocket"
)

type wsProxy struct{}

func newWSProxy() http.Handler {
	return &wsProxy{}
}

func (p *wsProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	websocket.Handler(wsHandler).ServeHTTP(w, r)
}

func wsHandler(client *websocket.Conn) {
	r := client.Request()
	t := target(r)
	if t == nil {
		client.WriteClose(http.StatusNotFound)
		return
	}

	logerr := func(err error) {
		if err != nil && err != io.EOF {
			log.Printf("[INFO] WS error for %s. %s", r.URL, err)
		}
	}

	// dial the server
	origin := r.Header.Get("Origin")
	targetURL := "ws://" + t.URL.Host + r.RequestURI
	server, err := websocket.Dial(targetURL, "", origin)
	if err != nil {
		logerr(err)
		return
	}

	// send data from client to server
	cerr := make(chan error)
	go func() {
		_, err := io.Copy(server, client)
		cerr <- err
	}()

	// send data from server to client
	serr := make(chan error)
	go func() {
		_, err = io.Copy(client, server)
		serr <- err
	}()

	// wait for either server or client to exit
	// and then close the other side
	start := time.Now()
	select {
	case err := <-cerr:
		logerr(err)
		server.Close()
	case err := <-serr:
		logerr(err)
		client.Close()
	}
	t.Timer.UpdateSince(start)
}
