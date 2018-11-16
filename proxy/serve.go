package proxy

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/proxy/tcp"
)

type Server interface {
	Close() error
	Serve(l net.Listener) error
	Shutdown(ctx context.Context) error
}

var (
	// mu guards servers which contains the list
	// of running proxy servers.
	mu      sync.Mutex
	servers []Server
)

func Close() {
	mu.Lock()
	for _, srv := range servers {
		srv.Close()
	}
	servers = []Server{}
	mu.Unlock()
}

func Shutdown(timeout time.Duration) {
	mu.Lock()
	srvs := make([]Server, len(servers))
	copy(srvs, servers)
	servers = []Server{}
	mu.Unlock()

	var wg sync.WaitGroup
	for _, srv := range srvs {
		wg.Add(1)
		go func(srv Server) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			srv.Shutdown(ctx)
		}(srv)
	}
	wg.Wait()
}

func ListenAndServeHTTP(l config.Listen, h http.Handler, cfg *tls.Config) error {
	ln, err := ListenTCP(l, cfg)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:         l.Addr,
		Handler:      h,
		ReadTimeout:  l.ReadTimeout,
		WriteTimeout: l.WriteTimeout,
		TLSConfig:    cfg,
	}
	return serve(ln, srv)
}

func ListenAndServeGRPC(l config.Listen, opts []grpc.ServerOption, cfg *tls.Config) error {
	ln, err := ListenTCP(l, cfg)
	if err != nil {
		return err
	}

	srv := &gRPCServer{
		server: grpc.NewServer(opts...),
	}

	return serve(ln, srv)
}

func ListenAndServeTCP(l config.Listen, h tcp.Handler, cfg *tls.Config) error {
	ln, err := ListenTCP(l, cfg)
	if err != nil {
		return err
	}
	srv := &tcp.Server{
		Addr:         l.Addr,
		Handler:      h,
		ReadTimeout:  l.ReadTimeout,
		WriteTimeout: l.WriteTimeout,
	}
	return serve(ln, srv)
}

func serve(ln net.Listener, srv Server) error {
	mu.Lock()
	servers = append(servers, srv)
	mu.Unlock()
	return srv.Serve(ln)
}
