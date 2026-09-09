package auth

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/fabiolb/fabio/config"
)

type external struct {
	client            *http.Client
	endpoint          string
	setAuthHeaders    []string
	appendAuthHeaders []string
}

func newExternalAuth(cfg config.ExternalAuth, client *http.Client) (*external, error) {
	if client == nil {
		return nil, fmt.Errorf("external auth requires an HTTP client")
	}

	return &external{
		client:            client,
		endpoint:          cfg.Endpoint,
		setAuthHeaders:    cfg.SetAuthHeaders,
		appendAuthHeaders: cfg.AppendAuthHeaders,
	}, nil
}

func (e *external) Authorized(request *http.Request, response http.ResponseWriter) AuthDecision {
	authRequest, err := http.NewRequestWithContext(request.Context(), http.MethodGet, e.endpoint, nil)
	if err != nil {
		return externalAuthError(response)
	}

	authRequest.Header = request.Header.Clone()
	authRequest.Header.Del("Content-Length")
	authRequest.Header.Del("Transfer-Encoding")
	addExternalForwardHeaders(authRequest.Header, request)

	authResponse, err := e.client.Do(authRequest)
	if err != nil {
		return externalAuthError(response)
	}
	defer authResponse.Body.Close()

	if authResponse.StatusCode >= http.StatusOK && authResponse.StatusCode < http.StatusMultipleChoices {
		for _, header := range e.setAuthHeaders {
			request.Header.Del(header)
			for _, value := range authResponse.Header.Values(header) {
				request.Header.Add(header, value)
			}
		}

		for _, header := range e.appendAuthHeaders {
			for _, value := range authResponse.Header.Values(header) {
				request.Header.Add(header, value)
			}
		}

		_, _ = io.Copy(io.Discard, authResponse.Body)
		return AuthDecision{Authorized: true}
	}

	copyExternalAuthResponseHeaders(response.Header(), authResponse.Header)
	response.WriteHeader(authResponse.StatusCode)
	_, _ = io.Copy(response, authResponse.Body)
	return AuthDecision{Done: true}
}

func externalAuthError(response http.ResponseWriter) AuthDecision {
	http.Error(response, "external authorization failed", http.StatusBadGateway)
	return AuthDecision{Done: true}
}

func addExternalForwardHeaders(header http.Header, request *http.Request) {
	header.Set("X-Forwarded-Method", request.Method)
	requestURI := request.RequestURI
	if requestURI == "" {
		requestURI = request.URL.RequestURI()
	}
	header.Set("X-Forwarded-Uri", requestURI)

	if header.Get("X-Forwarded-Host") == "" && request.Host != "" {
		header.Set("X-Forwarded-Host", request.Host)
	}

	if header.Get("X-Forwarded-Proto") == "" {
		proto := externalRequestScheme(request)
		switch proto {
		case "ws":
			proto = "http"
		case "wss":
			proto = "https"
		}
		header.Set("X-Forwarded-Proto", proto)
	}

	if header.Get("X-Forwarded-Port") == "" {
		header.Set("X-Forwarded-Port", externalRequestPort(request))
	}

	remoteIP, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return
	}

	forwardedFor := remoteIP
	if prior := header.Values("X-Forwarded-For"); len(prior) > 0 {
		forwardedFor = strings.Join(prior, ", ") + ", " + remoteIP
	}
	header.Set("X-Forwarded-For", forwardedFor)

	if header.Get("X-Real-Ip") == "" {
		header.Set("X-Real-Ip", remoteIP)
	}

	if header.Get("Forwarded") == "" {
		header.Set("Forwarded", "for="+remoteIP+"; proto="+externalRequestScheme(request))
	}
}

func externalRequestScheme(request *http.Request) string {
	if proto := request.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}

	if forwarded := request.Header.Get("Forwarded"); forwarded != "" {
		parts := strings.SplitAfterN(forwarded, "proto=", 2)
		if len(parts) == 2 {
			if n := strings.IndexRune(parts[1], ';'); n >= 0 {
				return parts[1][:n]
			}
			return parts[1]
		}
	}

	websocket := request.Header.Get("Upgrade") == "websocket"
	switch {
	case websocket && request.TLS != nil:
		return "wss"
	case websocket:
		return "ws"
	case request.TLS != nil:
		return "https"
	default:
		return "http"
	}
}

func externalRequestPort(request *http.Request) string {
	if _, port, err := net.SplitHostPort(request.Host); err == nil && port != "" {
		return port
	}
	if request.TLS != nil {
		return "443"
	}
	return "80"
}

var externalHopHeaders = map[string]bool{
	"Connection":          true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"Proxy-Connection":    true,
	"Te":                  true,
	"Trailer":             true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
}

func copyExternalAuthResponseHeaders(dst, src http.Header) {
	hopHeaders := make(map[string]bool, len(externalHopHeaders))
	for header := range externalHopHeaders {
		hopHeaders[header] = true
	}
	for _, value := range src.Values("Connection") {
		for header := range strings.SplitSeq(value, ",") {
			hopHeaders[http.CanonicalHeaderKey(strings.TrimSpace(header))] = true
		}
	}

	for header, values := range src {
		if hopHeaders[http.CanonicalHeaderKey(header)] {
			continue
		}
		for _, value := range values {
			dst.Add(header, value)
		}
	}
}
