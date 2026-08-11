package proxy

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// StatusClientClosedRequest non-standard HTTP status code for client disconnection
const StatusClientClosedRequest = 499

func newHTTPProxy(target *url.URL, tr http.RoundTripper, flush time.Duration) http.Handler {
	return &httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			req.Out.URL.Scheme = target.Scheme
			req.Out.URL.Host = target.Host
			req.Out.URL.Path = target.Path
			req.Out.URL.RawQuery = target.RawQuery
			if _, ok := req.Out.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Out.Header.Set("User-Agent", "")
			}

			// Preserve X-Forwarded-For from inbound request before calling SetXForwarded
			// SetXForwarded will append the client IP to it
			if xff := req.In.Header.Get("X-Forwarded-For"); xff != "" {
				req.Out.Header.Set("X-Forwarded-For", xff)
			}

			// SetXForwarded will handle X-Forwarded-For (append), X-Forwarded-Host, and X-Forwarded-Proto
			// Other headers (X-Forwarded-Port, X-Forwarded-Prefix, Forwarded) are already set by addHeaders()
			req.SetXForwarded()
		},
		FlushInterval: flush,
		Transport:     tr,
		ErrorHandler:  httpProxyErrorHandler,
	}
}

func httpProxyErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	// According to https://golang.org/src/net/http/httputil/reverseproxy.go#L74, Go will return a 502 (Bad Gateway) StatusCode by default if no ErrorHandler is provided
	// If a "context canceled" error is returned by the http.Request handler this means the client closed the connection before getting a response
	// So we are changing the StatusCode on these situations to the non-standard 499 (Client Closed Request)

	statusCode := http.StatusInternalServerError

	if e, ok := err.(net.Error); ok {
		if e.Timeout() {
			statusCode = http.StatusGatewayTimeout
		} else {
			statusCode = http.StatusBadGateway
		}
	} else if err == io.EOF {
		statusCode = http.StatusBadGateway
	} else if err == context.Canceled {
		statusCode = StatusClientClosedRequest
	}

	w.WriteHeader(statusCode)
	// Theres nothing we can do if the client closes the connection and logging the "context canceled" errors will just add noise to the error log
	// Note: The access_log will still log the 499 response status codes
	if statusCode != StatusClientClosedRequest {
		log.Print("[ERROR] ", err)
	}
}
