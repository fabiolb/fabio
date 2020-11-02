package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/proxy/tcp"

	"github.com/armon/go-proxyproto"
	"github.com/inetaf/tcpproxy"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
		IdleTimeout:  l.IdleTimeout,
		TLSConfig:    cfg,
	}
	return serve(ln, srv)
}

func ListenAndServePrometheus(l config.Listen, pcfg config.Prometheus, cfg *tls.Config) error {
	ln, err := ListenTCP(l, cfg)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	if pcfg.Path != "/" {
		mux.HandleFunc("/", func(rw http.ResponseWriter, _ *http.Request) {
			rw.Header().Set("Location", pcfg.Path)
			rw.WriteHeader(http.StatusPermanentRedirect)
		})
	}
	mux.Handle(pcfg.Path, promhttp.Handler())

	srv := &http.Server{
		Addr:         l.Addr,
		Handler:      mux,
		ReadTimeout:  l.ReadTimeout,
		WriteTimeout: l.WriteTimeout,
		IdleTimeout:  l.IdleTimeout,
		TLSConfig:    cfg,
	}
	return serve(ln, srv)
}

func ListenAndServeHTTPSTCPSNI(l config.Listen, h http.Handler, p tcp.Handler, cfg *tls.Config, m tcpproxy.Matcher) error {
	// we only want proxy proto enabled on tcp proxies
	pxyProto := l.ProxyProto
	l.ProxyProto = false
	tp := &tcpproxy.Proxy{
		ListenFunc: func(net, laddr string) (net.Listener, error) {
			// cfg is nil here so it's not terminating TLS (yet)
			return ListenTCP(l, nil)
		},
	}

	// This inspects SNI for matches.  If this succeeds then we Proxy tcp.
	tcpSNIListener := &tcpproxy.TargetListener{Address: l.Addr}
	tp.AddSNIMatchRoute(l.Addr, m, tcpSNIListener)

	// Fallthrough to https
	httpsListener := &tcpproxy.TargetListener{Address: l.Addr}
	tp.AddRoute(l.Addr, httpsListener)

	// Start the listener
	err := tp.Start()
	if err != nil {
		return err
	}

	tps := &InetAfTCPProxyServer{Proxy: tp}
	var tln net.Listener = tcpSNIListener
	// enable proxy protocol on the tcp side if configured to do so
	if pxyProto {
		tln = &proxyproto.Listener{
			Listener:           tln,
			ProxyHeaderTimeout: l.ProxyHeaderTimeout,
		}
	}
	tps.ServeLater(tln, &tcp.Server{
		Addr:         l.Addr,
		Handler:      p,
		ReadTimeout:  l.ReadTimeout,
		WriteTimeout: l.WriteTimeout,
	})

	// wrap TargetListener in a tls terminating version for HTTPS
	tps.ServeLater(tls.NewListener(httpsListener, cfg), &http.Server{
		Addr:         l.Addr,
		Handler:      h,
		ReadTimeout:  l.ReadTimeout,
		WriteTimeout: l.WriteTimeout,
		IdleTimeout:  l.IdleTimeout,
		TLSConfig:    cfg,
	})

	// tcpproxy creates its own listener from the configuration above so we can
	// safely pass nil here.
	return serve(nil, tps)
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
	err := srv.Serve(ln)
	if err != nil {
		var opErr *net.OpError
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		} else if errors.As(err, &opErr) {
			if opErr.Err != nil && opErr.Err.Error() == "use of closed network connection" {
				err = nil
			}
		}
	}
	return err
}
