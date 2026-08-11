package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// BenchmarkNewHTTPProxy benchmarks the httputil.ReverseProxy created by newHTTPProxy
// with focus on the Rewrite function and header manipulation including SetXForwarded
func BenchmarkNewHTTPProxy(b *testing.B) {
	// Create a simple backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)

	tests := []struct {
		name          string
		flushInterval time.Duration
		headers       map[string]string
	}{
		{
			name:          "no flush with basic headers",
			flushInterval: 0,
			headers: map[string]string{
				"User-Agent": "test-agent",
			},
		},
		{
			name:          "no flush with forwarded headers",
			flushInterval: 0,
			headers: map[string]string{
				"X-Forwarded-For":    "1.2.3.4",
				"X-Forwarded-Port":   "8080",
				"X-Forwarded-Prefix": "/api",
				"Forwarded":          "for=1.2.3.4",
			},
		},
		{
			name:          "with flush interval",
			flushInterval: 10 * time.Millisecond,
			headers: map[string]string{
				"X-Forwarded-For": "1.2.3.4",
			},
		},
		{
			name:          "no user agent header",
			flushInterval: 0,
			headers:       map[string]string{},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create the proxy handler using newHTTPProxy
			proxy := newHTTPProxy(backendURL, http.DefaultTransport, tt.flushInterval)

			// Create a test request
			req := httptest.NewRequest("GET", "http://example.com/test?foo=bar", nil)
			req.RemoteAddr = "127.0.0.1:12345"

			// Add headers
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				w := httptest.NewRecorder()
				proxy.ServeHTTP(w, req)
			}
		})
	}
}

// BenchmarkNewHTTPProxyRewrite specifically benchmarks the Rewrite function overhead
func BenchmarkNewHTTPProxyRewrite(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Minimal backend to isolate proxy overhead
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)
	proxy := newHTTPProxy(backendURL, http.DefaultTransport, 0)

	b.Run("with SetXForwarded", func(b *testing.B) {
		req := httptest.NewRequest("GET", "http://example.com/api/v1/users", nil)
		req.RemoteAddr = "192.168.1.100:54321"
		req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
		req.Header.Set("X-Forwarded-Port", "443")
		req.Header.Set("X-Forwarded-Prefix", "/api")
		req.Header.Set("Forwarded", "for=10.0.0.1;proto=https")

		b.ResetTimer()
		b.ReportAllocs()

		for range b.N {
			w := httptest.NewRecorder()
			proxy.ServeHTTP(w, req)
		}
	})

	b.Run("without existing forwarded headers", func(b *testing.B) {
		req := httptest.NewRequest("GET", "http://example.com/api/v1/users", nil)
		req.RemoteAddr = "192.168.1.100:54321"

		b.ResetTimer()
		b.ReportAllocs()

		for range b.N {
			w := httptest.NewRecorder()
			proxy.ServeHTTP(w, req)
		}
	})
}

// BenchmarkNewHTTPProxyParallel benchmarks concurrent requests through the proxy
func BenchmarkNewHTTPProxyParallel(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)
	proxy := newHTTPProxy(backendURL, http.DefaultTransport, 0)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		req := httptest.NewRequest("GET", "http://example.com/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "1.2.3.4")

		for pb.Next() {
			w := httptest.NewRecorder()
			proxy.ServeHTTP(w, req)
		}
	})
}
