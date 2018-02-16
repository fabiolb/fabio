package fastcgi

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fabiolb/fabio/config"
)

func getBackendDialer(b *staticFcgiBackend) func(string) (FCGIBackend, error) {
	return func(u string) (FCGIBackend, error) {
		return b, nil
	}
}

type rc struct {
	eof error
	v   []byte
}

func (r *rc) Close() error { return nil }

func (r *rc) Read(p []byte) (n int, err error) {
	if r.eof != nil {
		return 0, r.eof
	}

	copy(p, r.v)
	r.eof = io.EOF
	return len(r.v), nil
}

func TestServeHTTP(t *testing.T) {
	data := struct {
		readTimeout   time.Duration
		sendTimeout   time.Duration
		env           map[string]string
		method        string
		contentType   string
		body          io.Reader
		contentLength int64
	}{}

	req, err := http.NewRequest("post", "https://app.host/user/index.php/profile?key=value", strings.NewReader("test request body"))
	if err != nil {
		t.Error("failed to create new http request", err)
	}
	req.Header.Add("Content-Length", "17")
	req.Header.Add("Content-Type", "text/plain")

	response := http.Response{
		Status:           http.StatusText(http.StatusOK),
		StatusCode:       http.StatusOK,
		Proto:            "HTTP/1.1",
		ProtoMajor:       1,
		ProtoMinor:       1,
		Header:           http.Header{},
		ContentLength:    10,
		TransferEncoding: nil,
		Close:            true,
		Uncompressed:     false,
		Request:          req,
		TLS:              nil,
		Body: &rc{
			v: []byte("successful response"),
		},
	}

	backend := &staticFcgiBackend{
		SetReadTimeoutFunc: func(t time.Duration) error {
			data.readTimeout = t
			return nil
		},
		SetSendTimeoutFunc: func(t time.Duration) error {
			data.sendTimeout = t
			return nil
		},
		PostFunc: func(params map[string]string, m string, ct string, b io.Reader, cl int64) (*http.Response, error) {
			data.env = params
			data.method = m
			data.contentType = ct
			data.body = b
			data.contentLength = cl
			return &response, nil
		},
		StderrFunc: func() string { return "" },
	}

	proxy := Proxy{
		upstream: "app.fpm.internal",
		config: &config.Config{
			FastCGI: config.FastCGI{
				Root:         "/site",
				SplitPath:    ".php",
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
			},
		},
		dialFunc: getBackendDialer(backend),
	}

	resp := httptest.NewRecorder()
	proxy.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("expected response code '200', got '%d'", resp.Code)
	}
	if resp.Body.String() != "successful response" {
		t.Errorf("expected response body 'successful response', got '%s'", resp.Body.String())
	}
}

func TestBuildEnv(t *testing.T) {
	proxy := Proxy{
		upstream: "app.fpm.internal",
		config: &config.Config{
			FastCGI: config.FastCGI{
				Root:      "/site",
				SplitPath: ".php",
			},
		},
		dialFunc: nil,
	}

	req, err := http.NewRequest("post", "https://app.host:443/test/url?key=value", strings.NewReader("test request body"))
	if err != nil {
		t.Error("failed to create new http request", err)
	}

	req.Header.Add("Content-Length", "17")
	req.Header.Add("Content-Type", "text/plain")
	req.Header.Add("X-Custom-Header-One", "One")
	req.TLS = &tls.ConnectionState{}

	env, err := proxy.buildEnv(req, "/docs/index.php/user/profile")
	if err != nil {
		t.Error("failed to build environment", err)
	}

	expected := map[string]string{
		"AUTH_TYPE":                "",
		"CONTENT_LENGTH":           "17",
		"CONTENT_TYPE":             "text/plain",
		"GATEWAY_INTERFACE":        "CGI/1.1",
		"PATH_INFO":                "/user/profile",
		"QUERY_STRING":             "key=value",
		"REMOTE_ADDR":              "",
		"REMOTE_HOST":              "",
		"REMOTE_PORT":              "",
		"REMOTE_IDENT":             "",
		"REMOTE_USER":              "",
		"REQUEST_METHOD":           "post",
		"SERVER_NAME":              "app.host",
		"SERVER_PORT":              "443",
		"SERVER_PROTOCOL":          "HTTP/1.1",
		"SERVER_SOFTWARE":          "fabio",
		"DOCUMENT_ROOT":            "/site",
		"DOCUMENT_URI":             "/docs/index.php",
		"HTTP_HOST":                "app.host:443",
		"REQUEST_URI":              "/test/url?key=value",
		"SCRIPT_FILENAME":          "/site/docs/index.php/user/profile",
		"SCRIPT_NAME":              "/docs/index.php",
		"PATH_TRANSLATED":          "/site/user/profile",
		"HTTPS":                    "on",
		"HTTP_X_CUSTOM_HEADER_ONE": "One",
	}

	for ke, ve := range expected {
		if v, ok := env[ke]; !ok {
			t.Errorf("key '%s' is not present in environment map", ke)
		} else if v != ve {
			t.Errorf("Key '%s': expected value '%s', got '%s'", ke, ve, v)
		}
	}
}
