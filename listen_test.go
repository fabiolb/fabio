package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/proxy"
	"github.com/eBay/fabio/route"
)

func TestGracefulShutdown(t *testing.T) {
	req := func(url string) int {
		resp, err := http.Get(url)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}

	// start a server which responds after the shutdown has been triggered.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-quit // wait for shutdown signal
		return
	}))
	defer srv.Close()

	// load the routing table
	tbl, err := route.ParseString("route add svc / " + srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	route.SetTable(tbl)

	// start proxy with graceful shutdown period long enough
	// to complete one more request.
	var wg sync.WaitGroup
	l := config.Listen{Addr: "127.0.0.1:57777"}
	wg.Add(1)
	go func() {
		defer wg.Done()
		startListeners([]config.Listen{l}, 250*time.Millisecond, proxy.New(http.DefaultTransport, config.Proxy{}))
	}()

	// trigger shutdown after some time
	shutdownDelay := 100 * time.Millisecond
	go func() {
		time.Sleep(shutdownDelay)
		close(quit)
	}()

	// give proxy some time to start up
	// needs to be done before shutdown is triggered
	time.Sleep(shutdownDelay / 2)

	// make 200 OK request
	// start before and complete after shutdown was triggered
	if got, want := req("http://"+l.Addr+"/"), 200; got != want {
		t.Fatalf("request 1: got %v want %v", got, want)
	}

	// make 503 request
	// start and complete after shutdown was triggered
	if got, want := req("http://"+l.Addr+"/"), 503; got != want {
		t.Fatalf("got %v want %v", got, want)
	}

	// wait for listen() to return
	// note that the actual listeners have not returned yet
	wg.Wait()
}
