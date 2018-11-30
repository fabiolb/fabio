package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/metrics"
	"github.com/fabiolb/fabio/route"
	grpc_proxy "github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

type gRPCServer struct {
	server *grpc.Server
}

func (s *gRPCServer) Close() error {
	s.server.Stop()
	return nil
}

func (s *gRPCServer) Shutdown(ctx context.Context) error {
	s.server.GracefulStop()
	return nil
}

func (s *gRPCServer) Serve(lis net.Listener) error {
	return s.server.Serve(lis)
}

func GetGRPCDirector(tlscfg *tls.Config) func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {

	connectionPool := newGrpcConnectionPool(tlscfg)

	return func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		md, ok := metadata.FromIncomingContext(ctx)

		if !ok {
			return ctx, nil, fmt.Errorf("error extracting metadata from request")
		}

		outCtx, _ := context.WithCancel(ctx)
		outCtx = metadata.NewOutgoingContext(outCtx, md.Copy())

		target, _ := ctx.Value(targetKey{}).(*route.Target)

		if target == nil {
			log.Println("[WARN] grpc: no route for ", fullMethodName)
			return outCtx, nil, fmt.Errorf("no route found")
		}

		conn, err := connectionPool.Get(outCtx, target)

		return outCtx, conn, err
	}

}

type GrpcProxyInterceptor struct {
	Config       *config.Config
	StatsHandler *GrpcStatsHandler
}

type targetKey struct{}

type proxyStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (p proxyStream) Context() context.Context {
	return p.ctx
}

func (g GrpcProxyInterceptor) Stream(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := stream.Context()

	target, err := g.lookup(ctx, info.FullMethod)

	if err != nil {
		log.Println("[ERROR] grpc: error looking up route", err)
		return status.Error(codes.Internal, "internal error")
	}

	if target == nil {
		g.StatsHandler.NoRoute.Inc(1)
		log.Println("[WARN] grpc: no route found for", info.FullMethod)
		return status.Error(codes.NotFound, "no route found")
	}

	ctx = context.WithValue(ctx, targetKey{}, target)

	proxyStream := proxyStream{
		ServerStream: stream,
		ctx:          ctx,
	}

	start := time.Now()

	err = handler(srv, proxyStream)

	end := time.Now()
	dur := end.Sub(start)

	target.Timer.Update(dur)

	return err
}

func (g GrpcProxyInterceptor) lookup(ctx context.Context, fullMethodName string) (*route.Target, error) {
	pick := route.Picker[g.Config.Proxy.Strategy]
	match := route.Matcher[g.Config.Proxy.Matcher]

	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, fmt.Errorf("error extracting metadata from request")
	}

	reqUrl, err := url.ParseRequestURI(fullMethodName)

	if err != nil {
		log.Print("[WARN] Error parsing grpc request url ", fullMethodName)
		return nil, fmt.Errorf("error parsing request url")
	}

	headers := http.Header{}

	for k, v := range md {
		for _, h := range v {
			headers.Add(k, h)
		}
	}

	req := &http.Request{
		Host:   "",
		URL:    reqUrl,
		Header: headers,
	}

	return route.GetTable().Lookup(req, req.Header.Get("trace"), pick, match, g.Config.GlobMatchingDisabled), nil
}

type GrpcStatsHandler struct {
	Connect metrics.Counter
	Request metrics.Timer
	NoRoute metrics.Counter
}

type connCtxKey struct{}
type rpcCtxKey struct{}

func (h *GrpcStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return context.WithValue(ctx, connCtxKey{}, info)
}

func (h *GrpcStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return context.WithValue(ctx, rpcCtxKey{}, info)
}

func (h *GrpcStatsHandler) HandleRPC(ctx context.Context, rpc stats.RPCStats) {
	rpcStats, _ := rpc.(*stats.End)

	if rpcStats == nil {
		return
	}

	dur := rpcStats.EndTime.Sub(rpcStats.BeginTime)

	h.Request.Update(dur)

	s, _ := status.FromError(rpcStats.Error)
	metrics.DefaultRegistry.GetTimer(fmt.Sprintf("grpc.status.%s", strings.ToLower(s.Code().String())))
}

// HandleConn processes the Conn stats.
func (h *GrpcStatsHandler) HandleConn(ctx context.Context, conn stats.ConnStats) {
	connBegin, _ := conn.(*stats.ConnBegin)

	if connBegin != nil {
		h.Connect.Inc(1)
	}
}

type grpcConnectionPool struct {
	connections     map[*route.Target]*grpc.ClientConn
	lock            sync.RWMutex
	cleanupInterval time.Duration
	tlscfg          *tls.Config
}

func newGrpcConnectionPool(tlscfg *tls.Config) *grpcConnectionPool {
	cp := &grpcConnectionPool{
		connections:     make(map[*route.Target]*grpc.ClientConn),
		lock:            sync.RWMutex{},
		cleanupInterval: time.Second * 5,
		tlscfg:          tlscfg,
	}

	go cp.cleanup()

	return cp
}

func (p *grpcConnectionPool) Get(ctx context.Context, target *route.Target) (*grpc.ClientConn, error) {
	p.lock.RLock()
	conn := p.connections[target]
	p.lock.RUnlock()

	if conn != nil && conn.GetState() != connectivity.Shutdown {
		return conn, nil
	}

	return p.newConnection(ctx, target)
}

func (p *grpcConnectionPool) newConnection(ctx context.Context, target *route.Target) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.CallCustomCodec(grpc_proxy.Codec())),
	}

	if target.URL.Scheme == "grpcs" && p.tlscfg != nil {
		opts = append(opts, grpc.WithTransportCredentials(
			credentials.NewTLS(&tls.Config{
				ClientCAs:          p.tlscfg.ClientCAs,
				InsecureSkipVerify: target.TLSSkipVerify,
				// as per the http/2 spec, the host header isn't required, so if your
				// target service doesn't have IP SANs in it's certificate
				// then you will need to override the servername
				ServerName: target.Opts["grpcservername"],
			})))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.DialContext(ctx, target.URL.Host, opts...)

	if err == nil {
		p.Set(target, conn)
	}

	return conn, err
}

func (p *grpcConnectionPool) Set(target *route.Target, conn *grpc.ClientConn) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.connections[target] = conn
}

func (p *grpcConnectionPool) cleanup() {
	for {
		p.lock.Lock()
		table := route.GetTable()
		for target, cs := range p.connections {
			if cs.GetState() == connectivity.Shutdown {
				delete(p.connections, target)
				continue
			}

			if !hasTarget(target, table) {
				log.Println("[DEBUG] grpc: cleaning up connection to", target.URL.Host)
				cs.Close()
				delete(p.connections, target)
			}
		}
		p.lock.Unlock()
		time.Sleep(p.cleanupInterval)
	}
}

func hasTarget(target *route.Target, table route.Table) bool {
	for _, routes := range table {
		for _, r := range routes {
			for _, t := range r.Targets {
				if target == t {
					return true
				}
			}
		}
	}
	return false
}
