package proxy

import (
	"bytes"
	"compress/gzip"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/route"
)

func TestProxyProducesCorrectXffHeader(t *testing.T) {
	got := "not called"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Forwarded-For")
	}))
	defer server.Close()

	route.SetTable(mustParseTable("route add mock / " + server.URL))

	tr := &http.Transport{Dial: (&net.Dialer{}).Dial}
	proxy := NewHTTPProxy(tr, config.Proxy{LocalIP: "1.1.1.1", ClientIPHeader: "X-Forwarded-For"})

	req := &http.Request{
		RequestURI: "/",
		Header:     http.Header{"X-Forwarded-For": {"3.3.3.3"}},
		RemoteAddr: "2.2.2.2:666",
		URL:        &url.URL{},
	}

	proxy.ServeHTTP(httptest.NewRecorder(), req)

	if want := "3.3.3.3, 2.2.2.2"; got != want {
		t.Errorf("got %v, but want %v", got, want)
	}
}

func TestProxyNoRouteStaus(t *testing.T) {
	route.SetTable(make(route.Table))
	tr := &http.Transport{Dial: (&net.Dialer{}).Dial}
	cfg := config.Proxy{NoRouteStatus: 999}
	proxy := NewHTTPProxy(tr, cfg)
	req := &http.Request{
		RequestURI: "/",
		URL:        &url.URL{},
	}

	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)
	if got, want := rec.Code, cfg.NoRouteStatus; got != want {
		t.Fatalf("got %d want %d", got, want)
	}
}

func TestProxyGzipHandler(t *testing.T) {
	tests := []struct {
		desc            string
		content         http.HandlerFunc
		acceptEncoding  string
		contentEncoding string
		wantResponse    []byte
	}{
		{
			desc:            "plain body - compressed response",
			content:         plainHandler("text/plain"),
			acceptEncoding:  "gzip",
			contentEncoding: "gzip",
			wantResponse:    gzipContent,
		},
		{
			desc:            "plain body - compressed response (with charset)",
			content:         plainHandler("text/plain; charset=UTF-8"),
			acceptEncoding:  "gzip",
			contentEncoding: "gzip",
			wantResponse:    gzipContent,
		},
		{
			desc:            "compressed body - compressed response",
			content:         gzipHandler("text/plain; charset=UTF-8"),
			acceptEncoding:  "gzip",
			contentEncoding: "gzip",
			wantResponse:    gzipContent,
		},
		{
			desc:            "plain body - plain response",
			content:         plainHandler("text/plain"),
			acceptEncoding:  "",
			contentEncoding: "",
			wantResponse:    plainContent,
		},
		{
			desc:            "compressed body - plain response",
			content:         gzipHandler("text/plain"),
			acceptEncoding:  "",
			contentEncoding: "",
			wantResponse:    plainContent,
		},
		{
			desc:            "plain body - plain response (no match)",
			content:         plainHandler("text/javascript"),
			acceptEncoding:  "gzip",
			contentEncoding: "",
			wantResponse:    plainContent,
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop var
		t.Run(tt.desc, func(t *testing.T) {
			server := httptest.NewServer(tt.content)
			defer server.Close()

			route.SetTable(mustParseTable("route add mock / " + server.URL))

			tr := &http.Transport{Dial: (&net.Dialer{}).Dial}
			proxy := NewHTTPProxy(tr, config.Proxy{GZIPContentTypes: regexp.MustCompile("^text/plain(;.*)?$")})
			req := &http.Request{RequestURI: "/", RemoteAddr: "2.2.2.2:2222", Header: http.Header{"Accept-Encoding": []string{tt.acceptEncoding}}, URL: &url.URL{}}
			rec := httptest.NewRecorder()
			proxy.ServeHTTP(rec, req)

			if got, want := rec.Code, 200; got != want {
				t.Fatalf("got code %d want %d", got, want)
			}
			if got, want := rec.Header().Get("Content-Encoding"), tt.contentEncoding; got != want {
				t.Errorf("got content-encoding %q want %q", got, want)
			}
			if got, want := rec.Body.Bytes(), tt.wantResponse; !bytes.Equal(got, want) {
				t.Errorf("got body %q want %q", got, want)
			}
		})
	}
}

var plainContent = []byte("Hello World")
var gzipContent = compress(plainContent)

func plainHandler(contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write(plainContent)
	}
}

func gzipHandler(contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Encoding", "gzip")
		w.Write(gzipContent)
	}
}

// compress returns the gzip compressed content of b.
func compress(b []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(b); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func mustParseTable(s string) route.Table {
	t, err := route.ParseTable(s)
	if err != nil {
		panic(err)
	}
	return t
}
