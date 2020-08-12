package proxy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/inetaf/tcpproxy"
)

type childProxy struct {
	l net.Listener
	s Server
}

type InetAfTCPProxyServer struct {
	Proxy    *tcpproxy.Proxy
	children []*childProxy
}

// Close - implements Server - is this even called?
func (tps *InetAfTCPProxyServer) Close() error {
	_ = tps.Proxy.Close()
	firstErr := tps.Proxy.Wait()
	errChan := make(chan error, len(tps.children))
	for _, sl := range tps.children {
		go func(sl *childProxy) {
			errChan <- sl.s.Close()
		}(sl)
	}
	for range tps.children {
		err := <-errChan
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		if firstErr == nil {
			firstErr = err
		}
		if err != nil {
			log.Printf("[ERROR] %s", err)
		}
	}
	return firstErr

}

// Serve - implements server.  The listener is ignored, but it
// calls serve on the children
func (tps *InetAfTCPProxyServer) Serve(_ net.Listener) error {
	if len(tps.children) == 0 {
		return fmt.Errorf("no children defined for listener")
	}
	errChan := make(chan error, len(tps.children))
	for _, sl := range tps.children {
		go func(sl *childProxy) {
			errChan <- sl.s.Serve(sl.l)
		}(sl)
	}
	firstErr := tps.Proxy.Wait()
	for range tps.children {
		err := <-errChan
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		if firstErr == nil {
			firstErr = err
		}
		if err != nil {
			log.Print("[FATAL] ", err)
		}
	}
	return firstErr
}

// ServeLater - l is really only for listeners that are
// tcpproxy.TargetListener or a derivative.  Don't call after
// Serve() is called.
func (tps *InetAfTCPProxyServer) ServeLater(l net.Listener, s Server) {
	tps.children = append(tps.children, &childProxy{l, s})
}

func (tps *InetAfTCPProxyServer) Shutdown(ctx context.Context) error {
	_ = tps.Proxy.Close()        // always returns nil error anyway
	firstErr := tps.Proxy.Wait() // wait for outer listener to close before telling the childProxy
	errChan := make(chan error, len(tps.children))
	for _, sl := range tps.children {
		go func(sl *childProxy) {
			errChan <- sl.s.Shutdown(ctx)
		}(sl)
	}
	for range tps.children {
		err := <-errChan
		if firstErr == nil {
			firstErr = err
		}
		if err != nil {
			log.Print("[ERROR] ", err)
		}
	}
	return firstErr
}
