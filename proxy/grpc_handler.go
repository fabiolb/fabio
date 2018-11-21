package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/metrics"
	"github.com/fabiolb/fabio/route"
	grpc_proxy "github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	return func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		md, ok := metadata.FromIncomingContext(ctx)

		if !ok {
			return ctx, nil, fmt.Errorf("error extracting metadata from request")
		}

		target, _ := ctx.Value(targetKey{}).(*route.Target)

		if target == nil {
			log.Println("[WARN] grpc: no route for ", fullMethodName)
			return ctx, nil, fmt.Errorf("no route found")
		}

		opts := []grpc.DialOption{
			grpc.WithDefaultCallOptions(grpc.CallCustomCodec(grpc_proxy.Codec())),
		}

		if target.URL.Scheme == "grpcs" && tlscfg != nil {
			opts = append(opts, grpc.WithTransportCredentials(
				credentials.NewTLS(&tls.Config{
					ClientCAs:          tlscfg.ClientCAs,
					InsecureSkipVerify: target.TLSSkipVerify,
					// as per the http/2 spec, the host header isn't required, so if your
					// target service doesn't have IP SANs in it's certificate
					// then you will need to override the servername
					ServerName: target.Opts["grpcservername"],
				})))
		}

		newCtx := context.Background()
		newCtx = metadata.NewOutgoingContext(newCtx, md)

		conn, err := grpc.DialContext(newCtx, target.URL.Host, opts...)

		return newCtx, conn, err
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

func (g GrpcProxyInterceptor) Unary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	target, err := g.lookup(ctx, info.FullMethod)

	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, targetKey{}, target)

	start := time.Now()

	res, err := handler(ctx, req)

	end := time.Now()
	dur := end.Sub(start)

	target.Timer.Update(dur)

	return res, err
}

func (g GrpcProxyInterceptor) Stream(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := stream.Context()

	target, err := g.lookup(ctx, info.FullMethod)

	if err != nil {
		return err
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

	if target != nil {
		target.Timer.Update(dur)
	} else {
		g.StatsHandler.NoRoute.Inc(1)
	}

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
