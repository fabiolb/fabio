package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/metrics"
	"github.com/fabiolb/fabio/route"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func leastConnRouteTable(t *testing.T, lines string) route.Table {
	t.Helper()
	tbl, err := route.NewTable(bytes.NewBufferString(lines))
	if err != nil {
		t.Fatal(err)
	}
	return tbl
}

func leastConnHTTPProxy(tbl route.Table) *HTTPProxy {
	cache := route.NewGlobCache(16)
	return &HTTPProxy{
		Transport: http.DefaultTransport,
		Lookup: func(req *http.Request) *route.Target {
			return tbl.Lookup(req, route.Picker["lconn"], route.Matcher["prefix"], cache, true)
		},
	}
}

func waitLeastConnActive(t *testing.T, target *route.Target, want int64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if target.Active() == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("target %s: got %d active operations, want %d", target.URL, target.Active(), want)
}

func TestLeastConnHTTP(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(entered)
		<-release
		fmt.Fprint(w, "first")
	}))
	defer first.Close()
	second := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "second")
	}))
	defer second.Close()
	tbl := leastConnRouteTable(t, fmt.Sprintf("route add svc / %s\nroute add svc / %s", first.URL, second.URL))
	proxy := httptest.NewServer(leastConnHTTPProxy(tbl))
	defer proxy.Close()

	finished := make(chan error, 1)
	go func() {
		resp, err := http.Get(proxy.URL)
		if err != nil {
			finished <- err
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err == nil && string(body) != "first" {
			err = fmt.Errorf("first request received %q", body)
		}
		finished <- err
	}()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		close(release)
		t.Fatal("first backend did not receive request")
	}

	busy := tbl[""][0].Targets[0]
	idle := tbl[""][0].Targets[1]
	waitLeastConnActive(t, busy, 1)
	resp, err := http.Get(proxy.URL)
	if err != nil {
		close(release)
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil || string(body) != "second" {
		close(release)
		t.Fatalf("second request: body %q, error %v", body, err)
	}
	waitLeastConnActive(t, idle, 0)
	close(release)
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
	waitLeastConnActive(t, busy, 0)
}

func TestLeastConnHTTPShortCircuits(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	defer backend.Close()
	target := &route.Target{URL: mustParse(backend.URL)}
	proxy := &HTTPProxy{Transport: http.DefaultTransport, Lookup: func(*http.Request) *route.Target { return target }}
	do := func() int {
		t.Helper()
		w := httptest.NewRecorder()
		proxy.ServeHTTP(w, httptest.NewRequest("GET", "http://example.com/", nil))
		if got := target.Active(); got != 0 {
			t.Errorf("short-circuited request left %d active operations", got)
		}
		return w.Code
	}

	target.AuthScheme = "missing"
	if got := do(); got != http.StatusUnauthorized {
		t.Fatalf("missing auth scheme: status %d", got)
	}
	target.AuthScheme = ""
	target.RedirectCode = http.StatusFound
	target.RedirectURL = &url.URL{Scheme: "https", Host: "example.net", Path: "/"}
	if got := do(); got != http.StatusFound {
		t.Fatalf("redirect: status %d", got)
	}
	target.RedirectCode = 0
	target.RedirectURL = nil
	target.Opts = map[string]string{"allow": "ip:10.0.0.0/8"}
	if err := target.ProcessAccessRules(); err != nil {
		t.Fatal(err)
	}
	if got := do(); got != http.StatusForbidden {
		t.Fatalf("access denied: status %d", got)
	}

	target = &route.Target{URL: mustParse(backend.URL)}
	backend.Close()
	if got := do(); got != http.StatusBadGateway {
		t.Fatalf("failed upstream: status %d", got)
	}
}

func TestLeastConnSSE(t *testing.T) {
	release := make(chan struct{})
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: hello\n\n")
		w.(http.Flusher).Flush()
		<-release
	}))
	defer backend.Close()
	tbl := leastConnRouteTable(t, "route add svc / "+backend.URL)
	proxy := httptest.NewServer(leastConnHTTPProxy(tbl))
	defer proxy.Close()

	req, err := http.NewRequest("GET", proxy.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		close(release)
		t.Fatal(err)
	}
	target := tbl[""][0].Targets[0]
	waitLeastConnActive(t, target, 1)
	close(release)
	_, err = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	waitLeastConnActive(t, target, 0)
}

func TestLeastConnWebSocket(t *testing.T) {
	backend := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		io.Copy(io.Discard, ws)
	}))
	defer backend.Close()
	tbl := leastConnRouteTable(t, "route add svc / "+backend.URL)
	proxy := httptest.NewServer(leastConnHTTPProxy(tbl))
	defer proxy.Close()

	ws, err := websocket.Dial("ws"+proxy.URL[len("http"):], "", "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}
	target := tbl[""][0].Targets[0]
	waitLeastConnActive(t, target, 1)
	ws.Close()
	waitLeastConnActive(t, target, 0)
}

type leastConnServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s leastConnServerStream) Context() context.Context { return s.ctx }

func TestLeastConnGRPC(t *testing.T) {
	previous := route.GetTable()
	defer route.SetTable(previous)
	tbl := leastConnRouteTable(t, `route add svc / grpc://first:50051/
route add svc / grpc://second:50051/`)
	route.SetTable(tbl)
	g := GrpcProxyInterceptor{
		Config:       &config.Config{Proxy: config.Proxy{Strategy: "lconn", Matcher: "prefix"}},
		GlobCache:    route.NewGlobCache(16),
		StatsHandler: &GrpcStatsHandler{NoRoute: (metrics.DiscardProvider{}).NewCounter("noroute")},
	}
	stream := leastConnServerStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.MD{})}
	info := &grpc.StreamServerInfo{FullMethod: "/example.Service/Stream"}
	entered := make(chan *route.Target, 1)
	release := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		finished <- g.Stream(nil, stream, info, func(_ any, srvStream grpc.ServerStream) error {
			entered <- srvStream.Context().Value(targetKey{}).(*route.Target)
			<-release
			return nil
		})
	}()
	var first *route.Target
	select {
	case first = <-entered:
	case <-time.After(2 * time.Second):
		close(release)
		t.Fatal("first RPC did not start")
	}
	waitLeastConnActive(t, first, 1)

	wantErr := errors.New("upstream failed")
	var second *route.Target
	err := g.Stream(nil, stream, info, func(_ any, srvStream grpc.ServerStream) error {
		second = srvStream.Context().Value(targetKey{}).(*route.Target)
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("handler error: got %v, want %v", err, wantErr)
	}
	if second == first {
		t.Errorf("second RPC used busy target %s", second.URL)
	}
	if second != nil {
		waitLeastConnActive(t, second, 0)
	}
	close(release)
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
	waitLeastConnActive(t, first, 0)
}
