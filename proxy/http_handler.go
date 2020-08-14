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
		// this is a simplified director function based on the
		// httputil.NewSingleHostReverseProxy() which does not
		// mangle the request and target URL since the target
		// URL is already in the correct format.
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = target.Path
			req.URL.RawQuery = target.RawQuery
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
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

	return
}
