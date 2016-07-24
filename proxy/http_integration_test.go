package proxy

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/logger"
	"github.com/eBay/fabio/route"
	"github.com/pascaldekloe/goe/verify"
)

func TestProxyProducesCorrectXffHeader(t *testing.T) {
	got := "not called"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Forwarded-For")
	}))
	defer server.Close()

	proxy := &HTTPProxy{
		Config:    config.Proxy{LocalIP: "1.1.1.1", ClientIPHeader: "X-Forwarded-For"},
		Transport: http.DefaultTransport,
		Lookup: func(r *http.Request) *route.Target {
			return &route.Target{URL: mustParse(server.URL)}
		},
	}
	req := makeReq("/")
	req.Header.Set("X-Forwarded-For", "3.3.3.3")
	proxy.ServeHTTP(httptest.NewRecorder(), req)

	if want := "3.3.3.3, 2.2.2.2"; got != want {
		t.Errorf("got %v, but want %v", got, want)
	}
}

func TestProxyNoRouteStaus(t *testing.T) {
	proxy := &HTTPProxy{
		Config:    config.Proxy{NoRouteStatus: 999},
		Transport: http.DefaultTransport,
		Lookup:    func(*http.Request) *route.Target { return nil },
	}
	req := makeReq("/")
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	if got, want := rec.Code, proxy.Config.NoRouteStatus; got != want {
		t.Fatalf("got %d want %d", got, want)
	}
}

func TestProxyStripsPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/bar":
			w.Write([]byte("OK"))
		default:
			w.WriteHeader(404)
		}
	}))

	proxy := &HTTPProxy{
		Transport: http.DefaultTransport,
		Lookup: func(r *http.Request) *route.Target {
			tbl, _ := route.NewTable("route add mock /foo/bar " + server.URL + ` opts "strip=/foo"`)
			return tbl.Lookup(r, "", route.Picker["rr"], route.Matcher["prefix"])
		},
	}

	req := makeReq("/foo/bar")
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Fatalf("got status %d want %d", got, want)
	}
	if got, want := rec.Body.String(), "OK"; got != want {
		t.Fatalf("got body %q want %q", got, want)
	}
}

func TestProxyLogOutput(t *testing.T) {
	// build a format string from all log fields and one header field
	fields := []string{"header.X-Foo:$header.X-Foo"}
	for _, k := range logger.Fields {
		fields = append(fields, k[1:]+":"+k)
	}
	format := strings.Join(fields, ";")

	// create a logger
	var b bytes.Buffer
	l, err := logger.New(&b, format)
	if err != nil {
		t.Fatal("logger.New: ", err)
	}

	// create an upstream server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "foo")
	}))
	defer server.Close()

	// create a proxy handler with mocked time
	tm := time.Date(2016, 1, 1, 0, 0, 0, 12345678, time.UTC)
	proxy := &HTTPProxy{
		Time: func() time.Time {
			defer func() { tm = tm.Add(1111111111 * time.Nanosecond) }()
			return tm
		},
		Transport: http.DefaultTransport,
		Lookup: func(r *http.Request) *route.Target {
			return &route.Target{URL: mustParse(server.URL)}
		},
		Logger: l,
	}

	// start an http server with the proxy handler
	// which captures some parameters from the request
	var remoteAddr string
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteAddr = r.RemoteAddr
		proxy.ServeHTTP(w, r)
	}))
	defer proxyServer.Close()

	// create the request
	rawurl := proxyServer.URL + "/foo?x=y"
	req, _ := http.NewRequest("GET", rawurl, nil)
	req.Host = "example.com"
	req.Header.Set("X-Foo", "bar")

	// execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal("http.Get: ", err)
	}
	if resp.StatusCode != 200 {
		t.Fatal("http.Get: want 200 got ", resp.StatusCode)
	}

	upstreamURL, _ := url.Parse(server.URL)
	upstreamHost, upstreamPort, _ := net.SplitHostPort(upstreamURL.Host)
	remoteHost, remotePort, _ := net.SplitHostPort(remoteAddr)
	want := []string{
		"header.X-Foo:bar",
		"remote_addr:" + remoteAddr,
		"remote_host:" + remoteHost,
		"remote_port:" + remotePort,
		"request:GET /foo?x=y HTTP/1.1",
		"request_args:x=y",
		"request_host:example.com",
		"request_method:GET",
		"request_proto:HTTP/1.1",
		"request_scheme:http",
		"request_uri:/foo?x=y",
		"request_url:http://example.com/foo?x=y",
		"response_body_size:3",
		"response_status:200",
		"response_time_ms:1.111",
		"response_time_ns:1.111111111",
		"response_time_us:1.111111",
		"time_common:01/Jan/2016:00:00:01 +0000",
		"time_rfc3339:2016-01-01T00:00:01Z",
		"time_rfc3339_ms:2016-01-01T00:00:01.123Z",
		"time_rfc3339_ns:2016-01-01T00:00:01.123456789Z",
		"time_rfc3339_us:2016-01-01T00:00:01.123456Z",
		"time_unix_ms:1451606401123",
		"time_unix_ns:1451606401123456789",
		"time_unix_us:1451606401123456",
		"upstream_addr:" + upstreamURL.Host,
		"upstream_host:" + upstreamHost,
		"upstream_port:" + upstreamPort,
		"upstream_request_scheme:" + upstreamURL.Scheme,
		"upstream_request_uri:/foo?x=y",
		"upstream_request_url:" + upstreamURL.String() + "/foo?x=y",
	}

	data := string(b.Bytes())
	data = data[:len(data)-1] // strip \n
	got := strings.Split(data, ";")
	sort.Strings(got)

	verify.Values(t, "", got, want)
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

			proxy := &HTTPProxy{
				Config: config.Proxy{
					GZIPContentTypes: regexp.MustCompile("^text/plain(;.*)?$"),
				},
				Transport: http.DefaultTransport,
				Lookup: func(r *http.Request) *route.Target {
					return &route.Target{URL: mustParse(server.URL)}
				},
			}
			req := makeReq("/")
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)
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

func mustParse(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}

func makeReq(path string) *http.Request {
	return &http.Request{
		Method:     "GET",
		RemoteAddr: "2.2.2.2:666",
		Header:     http.Header{},
		RequestURI: path,
		URL:        &url.URL{Path: path},
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

func BenchmarkProxyLogger(b *testing.B) {
	got := "not called"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-Forwarded-For")
	}))
	defer server.Close()

	format := "remote_addr time request body_bytes_sent http_referer http_user_agent server_name proxy_endpoint response_time request_args "
	l, err := logger.New(os.Stdout, format)
	if err != nil {
		b.Fatal("logger.NewHTTPLogger:", err)
	}

	proxy := &HTTPProxy{
		Config: config.Proxy{
			LocalIP:        "1.1.1.1",
			ClientIPHeader: "X-Forwarded-For",
		},
		Transport: http.DefaultTransport,
		Lookup: func(r *http.Request) *route.Target {
			tbl, _ := route.NewTable("route add mock / " + server.URL)
			return tbl.Lookup(r, "", route.Picker["rr"], route.Matcher["prefix"])
		},
		Logger: l,
	}

	req := &http.Request{
		RequestURI: "/",
		Header:     http.Header{"X-Forwarded-For": {"1.2.3.4"}},
		RemoteAddr: "2.2.2.2:666",
		URL:        &url.URL{},
		Method:     "GET",
		Proto:      "HTTP/1.1",
	}

	for i := 0; i < b.N; i++ {
		proxy.ServeHTTP(httptest.NewRecorder(), req)
	}
}
