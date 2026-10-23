package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/fcgi"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/route"
)

func startFastCGIServer(t *testing.T, handler http.Handler) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	go fcgi.Serve(listener, handler)
	return "http://" + listener.Addr().String()
}

func startFastCGIProxy(t *testing.T, lines string, cfg config.Proxy) *httptest.Server {
	t.Helper()
	tbl, err := route.NewTable(bytes.NewBufferString(lines))
	if err != nil {
		t.Fatal(err)
	}
	p := &HTTPProxy{
		Config:         cfg,
		ProtectHeaders: testProtectHeaders,
		Transport:      http.DefaultTransport,
		Lookup: func(r *http.Request) *route.Target {
			return tbl.Lookup(r, route.Picker["rr"], route.Matcher["prefix"], globCache, globEnabled)
		},
	}
	proxy := httptest.NewServer(p)
	t.Cleanup(proxy.Close)
	return proxy
}

type fastCGIObserved struct {
	Method         string `json:"method"`
	RequestURI     string `json:"request_uri"`
	ScriptName     string `json:"script_name"`
	ScriptFilename string `json:"script_filename"`
	DocumentRoot   string `json:"document_root"`
	Query          string `json:"query"`
	Host           string `json:"host"`
	ServerName     string `json:"server_name"`
	ServerPort     string `json:"server_port"`
	RemoteAddr     string `json:"remote_addr"`
	ContentType    string `json:"content_type"`
	ContentLength  int64  `json:"content_length"`
	Header         string `json:"header"`
	Body           string `json:"body"`
}

func TestProxyFastCGIRequestAndResponse(t *testing.T) {
	backend := startFastCGIServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		env := fcgi.ProcessEnv(r)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Backend", "fastcgi")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(fastCGIObserved{
			Method:         r.Method,
			RequestURI:     r.URL.RequestURI(),
			ScriptName:     r.URL.Path,
			ScriptFilename: env["SCRIPT_FILENAME"],
			DocumentRoot:   env["DOCUMENT_ROOT"],
			Query:          r.URL.RawQuery,
			Host:           r.Host,
			ServerName:     env["SERVER_NAME"],
			ServerPort:     env["SERVER_PORT"],
			RemoteAddr:     r.RemoteAddr,
			ContentType:    r.Header.Get("Content-Type"),
			ContentLength:  r.ContentLength,
			Header:         r.Header.Get("X-Test"),
			Body:           string(body),
		})
	}))
	proxy := startFastCGIProxy(t, fmt.Sprintf(
		`route add php /app %s opts "proto=fastcgi root=/srv/www strip=/app prepend=/scripts"`, backend,
	), config.Proxy{})

	for _, tt := range []struct {
		method, body string
	}{
		{http.MethodGet, ""},
		{http.MethodPost, "a=1&b=hello"},
	} {
		t.Run(tt.method, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, proxy.URL+"/app/index.php?city=S%C3%A9oul", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			req.Host = "example.com:8080"
			req.Header.Set("X-Test", "sent-through-fabio")
			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusCreated || resp.Header.Get("X-Backend") != "fastcgi" ||
				resp.Header.Get("Content-Type") != "application/json" {
				t.Fatalf("response status/headers: %s %v", resp.Status, resp.Header)
			}
			var got fastCGIObserved
			if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if got.Method != tt.method || got.Body != tt.body || got.ContentLength != int64(len(tt.body)) {
				t.Errorf("method/body/length: %+v", got)
			}
			if got.RequestURI != "/app/index.php?city=S%C3%A9oul" ||
				got.ScriptName != "/app/index.php" ||
				got.ScriptFilename != "/srv/www/scripts/index.php" ||
				got.DocumentRoot != "/srv/www" ||
				got.Query != "city=S%C3%A9oul" {
				t.Errorf("FastCGI path/URI variables: %+v", got)
			}
			if got.Host != req.Host || got.ServerName != "example.com" ||
				got.ServerPort != "8080" || got.Header != "sent-through-fabio" ||
				!strings.HasPrefix(got.RemoteAddr, "127.0.0.1:") {
				t.Errorf("FastCGI HTTP/host variables: %+v", got)
			}
			if tt.method == http.MethodPost && got.ContentType != "application/x-www-form-urlencoded" {
				t.Errorf("Content-Type: %q", got.ContentType)
			}
		})
	}
}

func TestProxyFastCGIStatuses(t *testing.T) {
	backend := startFastCGIServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/"))
		w.Header().Set("X-Status-Test", "present")
		w.WriteHeader(status)
		fmt.Fprint(w, "CGI response")
	}))
	proxy := startFastCGIProxy(t, fmt.Sprintf(
		`route add cgi / %s opts "proto=fastcgi"`, backend,
	), config.Proxy{})
	for _, status := range []int{http.StatusNotFound, http.StatusInternalServerError} {
		resp, err := http.Get(fmt.Sprintf("%s/%d", proxy.URL, status))
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil || resp.StatusCode != status || resp.Header.Get("X-Status-Test") != "present" ||
			string(body) != "CGI response" {
			t.Fatalf("status %d: response=%d header=%s body=%q err=%v", status, resp.StatusCode, resp.Header.Get("X-Status-Test"), body, err)
		}
	}
}

func TestProxyFastCGILargeStreamingBody(t *testing.T) {
	const size = 1024 * 1024
	backend := startFastCGIServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n, err := io.Copy(io.Discard, r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("X-Bytes-Received", strconv.FormatInt(n, 10))
		block := bytes.Repeat([]byte("a"), 32<<10)
		for range size / len(block) {
			w.Write(block)
		}
	}))
	proxy := startFastCGIProxy(t, fmt.Sprintf(
		`route add cgi / %s opts "proto=fastcgi"`, backend,
	), config.Proxy{})
	req, err := http.NewRequest(http.MethodPost, proxy.URL+"/upload", bytes.NewReader(bytes.Repeat([]byte("b"), size)))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	n, err := io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if err != nil || resp.StatusCode != http.StatusOK ||
		resp.Header.Get("X-Bytes-Received") != strconv.Itoa(size) || n != size {
		t.Fatalf("large streaming transfer: status %s bytes %d header %s error %v", resp.Status, n, resp.Header.Get("X-Bytes-Received"), err)
	}
}

func TestProxyFastCGISSE(t *testing.T) {
	release := make(chan struct{})
	backend := startFastCGIServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: hello\n\n")
		w.(http.Flusher).Flush()
		<-release
	}))
	proxy := startFastCGIProxy(t, fmt.Sprintf(
		`route add cgi / %s opts "proto=fastcgi"`, backend,
	), config.Proxy{FlushInterval: time.Millisecond})
	req, err := http.NewRequest(http.MethodGet, proxy.URL+"/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "text/event-stream")
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		close(release)
		t.Fatal(err)
	}
	got, err := bufio.NewReader(resp.Body).ReadString('\n')
	close(release)
	resp.Body.Close()
	if err != nil || got != "data: hello\n" {
		t.Fatalf("first FastCGI SSE event: %q, %v", got, err)
	}
}

func TestProxyFastCGIConnectionFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	proxy := startFastCGIProxy(t, fmt.Sprintf(
		`route add php / http://%s/ opts "proto=fastcgi"`, addr,
	), config.Proxy{DialTimeout: 100 * time.Millisecond})
	resp, err := http.Get(proxy.URL + "/missing.php")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("connection refused: status %d, want 502", resp.StatusCode)
	}
}

func TestProxyFastCGIResponseTimeout(t *testing.T) {
	release := make(chan struct{})
	backend := startFastCGIServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	proxy := startFastCGIProxy(t, fmt.Sprintf(
		`route add php / %s opts "proto=fastcgi"`, backend,
	), config.Proxy{ResponseHeaderTimeout: 100 * time.Millisecond})
	resp, err := http.Get(proxy.URL)
	close(release)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Fatalf("slow response header: status %d, want 504", resp.StatusCode)
	}
}

func TestProxyFastCGICancellation(t *testing.T) {
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	backend := startFastCGIServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		<-release
	}))
	defer close(release)
	proxy := startFastCGIProxy(t, fmt.Sprintf(
		`route add php / %s opts "proto=fastcgi"`, backend,
	), config.Proxy{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, proxy.URL+"/slow.php", nil).WithContext(ctx)
	var got *httptest.ResponseRecorder
	finished := make(chan struct{})
	go func() {
		got = httptest.NewRecorder()
		proxy.Config.Handler.ServeHTTP(got, req)
		close(finished)
	}()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("FastCGI handler was not entered")
	}
	cancel()
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled FastCGI request did not finish")
	}
	if got.Code != StatusClientClosedRequest {
		t.Fatalf("cancelled request: status %d, want 499", got.Code)
	}
}

func TestProxyFastCGIMalformedResponse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		readFastCGITestRequest(conn, io.Discard)
		writeFastCGIRecord(conn, fastCGIStdout, []byte("Not a CGI response header\r\n\r\n"))
		writeFastCGIRecord(conn, fastCGIEndRequest, make([]byte, 8))
	}()
	proxy := startFastCGIProxy(t, fmt.Sprintf(
		`route add php / http://%s/ opts "proto=fastcgi"`, listener.Addr().String(),
	), config.Proxy{ResponseHeaderTimeout: time.Second})
	resp, err := http.Get(proxy.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("malformed response: status %d, want 500", resp.StatusCode)
	}
}

// readFastCGITestRequest receives one complete request and decodes the PARAMS
// stream so wire-level CGI variables are checked before net/http/fcgi converts
// them to a Go http.Request. STDIN is copied to the provided writer.
func readFastCGITestRequest(conn net.Conn, stdin io.Writer) (map[string]string, error) {
	var params bytes.Buffer
	for {
		var header [8]byte
		if _, err := io.ReadFull(conn, header[:]); err != nil {
			return nil, err
		}
		if header[0] != fastCGIVersion || binary.BigEndian.Uint16(header[2:4]) != fastCGIRequestID {
			return nil, fmt.Errorf("invalid request header %v", header)
		}
		content := make([]byte, binary.BigEndian.Uint16(header[4:6]))
		if _, err := io.ReadFull(conn, content); err != nil {
			return nil, err
		}
		if _, err := io.CopyN(io.Discard, conn, int64(header[6])); err != nil {
			return nil, err
		}
		switch header[1] {
		case fastCGIParams:
			params.Write(content)
		case fastCGIStdin:
			if len(content) == 0 {
				return decodeFastCGITestParams(params.Bytes())
			}
			if _, err := stdin.Write(content); err != nil {
				return nil, err
			}
		}
	}
}

func decodeFastCGITestParams(data []byte) (map[string]string, error) {
	readLength := func() (int, error) {
		if len(data) == 0 {
			return 0, io.ErrUnexpectedEOF
		}
		if data[0] < 128 {
			n := int(data[0])
			data = data[1:]
			return n, nil
		}
		if len(data) < 4 {
			return 0, io.ErrUnexpectedEOF
		}
		n := int(binary.BigEndian.Uint32(data[:4]) & 0x7fffffff)
		data = data[4:]
		return n, nil
	}
	result := make(map[string]string)
	for len(data) > 0 {
		nameLen, err := readLength()
		if err != nil {
			return nil, err
		}
		valueLen, err := readLength()
		if err != nil {
			return nil, err
		}
		if nameLen > len(data) || valueLen > len(data)-nameLen {
			return nil, io.ErrUnexpectedEOF
		}
		result[string(data[:nameLen])] = string(data[nameLen : nameLen+valueLen])
		data = data[nameLen+valueLen:]
	}
	return result, nil
}

func TestProxyFastCGIWireParameters(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	type result struct {
		params map[string]string
		err    error
	}
	got := make(chan result, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			got <- result{err: err}
			return
		}
		defer conn.Close()
		params, err := readFastCGITestRequest(conn, io.Discard)
		got <- result{params: params, err: err}
		if err != nil {
			return
		}
		writeFastCGIRecord(conn, fastCGIStdout, []byte("Status: 200 OK\r\nContent-Type: text/plain\r\n\r\nwire ok"))
		writeFastCGIRecord(conn, fastCGIStdout, nil)
		writeFastCGIRecord(conn, fastCGIEndRequest, make([]byte, 8))
	}()
	proxy := startFastCGIProxy(t, fmt.Sprintf(
		`route add php /app http://%s/ opts "proto=fastcgi root=/srv/www strip=/app prepend=/scripts"`,
		listener.Addr().String(),
	), config.Proxy{})
	req, err := http.NewRequest(http.MethodGet, proxy.URL+"/app/%ED%95%9C%EA%B8%80.php?x=one", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Long", strings.Repeat("z", 256))
	req.Header.Set("Proxy", "http://attacker.example:8080")
	req.Header.Add("Cookie", "a=1")
	req.Header.Add("Cookie", "b=2")
	req.Header.Add("X-Repeated", "one")
	req.Header.Add("X-Repeated", "two")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil || resp.StatusCode != 200 || string(body) != "wire ok" {
		t.Fatalf("raw FastCGI response: status %v, body %q, err %v", resp.Status, body, err)
	}
	observed := <-got
	if observed.err != nil {
		t.Fatal(observed.err)
	}
	for k, want := range map[string]string{
		"REQUEST_METHOD":  "GET",
		"REQUEST_URI":     "/app/%ED%95%9C%EA%B8%80.php?x=one",
		"SCRIPT_NAME":     "/scripts/한글.php",
		"SCRIPT_FILENAME": "/srv/www/scripts/한글.php",
		"DOCUMENT_ROOT":   "/srv/www",
		"QUERY_STRING":    "x=one",
		"HTTP_HOST":       strings.TrimPrefix(proxy.URL, "http://"),
		"HTTP_X_LONG":     strings.Repeat("z", 256),
		"HTTP_COOKIE":     "a=1; b=2",
		"HTTP_X_REPEATED": "one, two",
	} {
		if observed.params[k] != want {
			t.Errorf("%s = %q, want %q", k, observed.params[k], want)
		}
	}
	if value, exists := observed.params["HTTP_PROXY"]; exists {
		t.Errorf("HTTP_PROXY must not be passed to CGI, got %q", value)
	}
}

func TestProxyFastCGIChunkedPOST(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	type result struct {
		params map[string]string
		body   string
		err    error
	}
	got := make(chan result, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			got <- result{err: err}
			return
		}
		defer conn.Close()
		var stdin bytes.Buffer
		params, err := readFastCGITestRequest(conn, &stdin)
		got <- result{params: params, body: stdin.String(), err: err}
		if err != nil {
			return
		}
		writeFastCGIRecord(conn, fastCGIStdout, []byte("Content-Type: text/plain\r\n\r\nreceived"))
		writeFastCGIRecord(conn, fastCGIStdout, nil)
		writeFastCGIRecord(conn, fastCGIEndRequest, make([]byte, 8))
	}()
	proxy := startFastCGIProxy(t, fmt.Sprintf(
		`route add cgi / http://%s/ opts "proto=fastcgi"`, listener.Addr().String(),
	), config.Proxy{})
	payload := strings.Repeat("chunk-data-", 9000)
	req, err := http.NewRequest(http.MethodPost, proxy.URL+"/upload", io.NopCloser(strings.NewReader(payload)))
	if err != nil {
		t.Fatal(err)
	}
	req.ContentLength = -1
	req.TransferEncoding = []string{"chunked"}
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil || resp.StatusCode != http.StatusOK || string(data) != "received" {
		t.Fatalf("chunked FastCGI response: status %s body %q err %v", resp.Status, data, err)
	}
	observed := <-got
	if observed.err != nil {
		t.Fatal(observed.err)
	}
	if observed.params["REQUEST_METHOD"] != http.MethodPost ||
		observed.params["CONTENT_LENGTH"] != strconv.Itoa(len(payload)) ||
		observed.params["CONTENT_TYPE"] != "application/octet-stream" ||
		observed.body != payload {
		t.Errorf("chunked FastCGI request: length=%q type=%q method=%q body bytes=%d, want %d",
			observed.params["CONTENT_LENGTH"], observed.params["CONTENT_TYPE"], observed.params["REQUEST_METHOD"], len(observed.body), len(payload))
	}
}

func TestProxyFastCGIScriptBoundary(t *testing.T) {
	transport := fastCGITransport{root: "/srv/www"}
	for _, tc := range []struct{ script, want string }{
		{"/../../etc/passwd", "/srv/www/etc/passwd"},
		{"/nested/../index.php", "/srv/www/index.php"},
		{"/café.php", "/srv/www/café.php"},
	} {
		req := httptest.NewRequest(http.MethodGet, "http://example.com"+tc.script, nil)
		params, err := transport.params(req)
		if err != nil {
			t.Fatal(err)
		}
		if got := params["SCRIPT_FILENAME"]; got != tc.want {
			t.Errorf("script %q resolved to %q, want %q", tc.script, got, tc.want)
		}
	}
	transport.root = "relative/www"
	if _, err := transport.params(httptest.NewRequest(http.MethodGet, "http://example.com/index.php", nil)); err == nil {
		t.Fatal("relative FastCGI root must be rejected")
	}
}

func TestProxyFastCGIHTTPSListener(t *testing.T) {
	backend := startFastCGIServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fcgi.ProcessEnv(r)["REQUEST_SCHEME"])
	}))
	tbl, err := route.NewTable(bytes.NewBufferString(fmt.Sprintf(
		`route add php / %s opts "proto=fastcgi root=/srv/www"`, backend,
	)))
	if err != nil {
		t.Fatal(err)
	}
	listener := httptest.NewTLSServer(&HTTPProxy{
		Transport: http.DefaultTransport,
		Lookup: func(req *http.Request) *route.Target {
			return tbl.Lookup(req, route.Picker["rr"], route.Matcher["prefix"], globCache, globEnabled)
		},
	})
	defer listener.Close()
	resp, err := listener.Client().Get(listener.URL + "/index.php")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil || resp.StatusCode != http.StatusOK || string(data) != "https" {
		t.Fatalf("HTTPS listener FastCGI response: status %s body %q err %v", resp.Status, data, err)
	}
}
