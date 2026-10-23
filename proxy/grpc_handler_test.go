package proxy

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/metrics"
	"github.com/fabiolb/fabio/route"
	grpc_proxy "github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func TestGRPCProxyGzipForwarding(t *testing.T) {
	backendListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	backend := grpc.NewServer()
	healthServer := health.NewServer()
	healthServer.SetServingStatus("gzip-test", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(backend, healthServer)
	go backend.Serve(backendListener)
	t.Cleanup(backend.Stop)

	table, err := route.NewTable(bytes.NewBufferString("route add health /grpc.health.v1.Health grpc://" + backendListener.Addr().String()))
	if err != nil {
		t.Fatal(err)
	}
	previousTable := route.GetTable()
	route.SetTable(table)
	t.Cleanup(func() { route.SetTable(previousTable) })

	cfg := &config.Config{Proxy: config.Proxy{
		Strategy:             "rnd",
		Matcher:              "prefix",
		GRPCMaxRxMsgSize:     4 << 20,
		GRPCGShutdownTimeout: time.Second,
	}}
	interceptor := GrpcProxyInterceptor{
		Config:       cfg,
		StatsHandler: &GrpcStatsHandler{NoRoute: metrics.DiscardProvider{}.NewCounter("grpc.noroute")},
		GlobCache:    route.NewGlobCache(64),
	}
	proxyServer := grpc.NewServer(
		grpc.UnknownServiceHandler(grpc_proxy.TransparentHandler(GetGRPCDirector(nil, cfg))),
		grpc.StreamInterceptor(interceptor.Stream),
		grpc.MaxRecvMsgSize(cfg.Proxy.GRPCMaxRxMsgSize),
	)
	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go proxyServer.Serve(proxyListener)
	t.Cleanup(proxyServer.Stop)

	conn, err := grpc.NewClient(proxyListener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	client := healthpb.NewHealthClient(conn)

	for _, tc := range []struct {
		name string
		opts []grpc.CallOption
	}{
		{name: "gzip", opts: []grpc.CallOption{grpc.UseCompressor("gzip")}},
		{name: "uncompressed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			response, err := client.Check(ctx, &healthpb.HealthCheckRequest{Service: "gzip-test"}, tc.opts...)
			if err != nil {
				t.Fatalf("health check through gRPC proxy: %v", err)
			}
			if response.GetStatus() != healthpb.HealthCheckResponse_SERVING {
				t.Fatalf("health check status = %s, want SERVING", response.GetStatus())
			}
		})
	}
}
