package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/route"
	grpc_proxy "github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type GRPCServer struct {
	server *grpc.Server
}

func (s *GRPCServer) Close() error {
	s.server.Stop()
	return nil
}

func (s *GRPCServer) Shutdown(ctx context.Context) error {
	s.server.GracefulStop()
	return nil
}

func (s *GRPCServer) Serve(lis net.Listener) error {
	return s.server.Serve(lis)
}

func GetGRPCDirector(cfg *config.Config, tlscfg *tls.Config) func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {

	pick := route.Picker[cfg.Proxy.Strategy]
	match := route.Matcher[cfg.Proxy.Matcher]

	return func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		md, ok := metadata.FromIncomingContext(ctx)

		if !ok {
			return ctx, nil, fmt.Errorf("error extracting metadata from request")
		}

		reqUrl, err := url.ParseRequestURI(fullMethodName)

		if err != nil {
			return ctx, nil, fmt.Errorf("error parsing request url")
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

		target := route.GetTable().Lookup(req, req.Header.Get("trace"), pick, match, cfg.GlobMatchingDisabled)

		if target == nil {
			return nil, nil, fmt.Errorf("no route found")
		}

		opts := []grpc.DialOption{
			grpc.WithDefaultCallOptions(grpc.CallCustomCodec(grpc_proxy.Codec())),
		}

		if target.URL.Scheme == "grpcs" && tlscfg != nil {
			opts = append(opts, grpc.WithTransportCredentials(
				credentials.NewTLS(&tls.Config{
					ClientCAs:          tlscfg.ClientCAs,
					InsecureSkipVerify: target.TLSSkipVerify,
					ServerName:         target.Opts["grpcservername"],
				})))
		}

		newCtx := context.Background()
		newCtx = metadata.NewOutgoingContext(newCtx, md)
		conn, err := grpc.DialContext(newCtx, target.URL.Host, opts...)

		return newCtx, conn, err
	}
}
