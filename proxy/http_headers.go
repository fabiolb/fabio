package proxy

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/fabiolb/fabio/config"
)

// addHeaders adds/updates headers in request
//
// * add/update `Forwarded` header
// * add X-Forwarded-Proto header, if not present
// * add X-Real-Ip, if not present
// * ClientIPHeader != "": Set header with that name to <remote ip>
// * TLS connection: Set header with name from `cfg.TLSHeader` to `cfg.TLSHeaderValue`
//
func addHeaders(r *http.Request, cfg config.Proxy) error {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return errors.New("cannot parse " + r.RemoteAddr)
	}

	// set configurable ClientIPHeader
	// X-Real-Ip is set later and X-Forwarded-For is set
	// by the Go HTTP reverse proxy.
	if cfg.ClientIPHeader != "" &&
		cfg.ClientIPHeader != "X-Forwarded-For" &&
		cfg.ClientIPHeader != "X-Real-Ip" {
		r.Header.Set(cfg.ClientIPHeader, remoteIP)
	}

	if r.Header.Get("X-Real-Ip") == "" {
		r.Header.Set("X-Real-Ip", remoteIP)
	}

	// set the X-Forwarded-For header for websocket
	// connections since they aren't handled by the
	// http proxy which sets it.
	ws := r.Header.Get("Upgrade") == "websocket"
	if ws {
		r.Header.Set("X-Forwarded-For", remoteIP)
	}

	// Issue #133: Setting the X-Forwarded-Proto header to
	// anything other than 'http' or 'https' breaks java
	// websocket clients which use java.net.URL for composing
	// the forwarded URL. Since X-Forwarded-Proto is not
	// specified the common practice is to set it to either
	// 'http' for 'ws' and 'https' for 'wss' connections.
	proto := scheme(r)
	if r.Header.Get("X-Forwarded-Proto") == "" {
		switch proto {
		case "ws":
			r.Header.Set("X-Forwarded-Proto", "http")
		case "wss":
			r.Header.Set("X-Forwarded-Proto", "https")
		default:
			r.Header.Set("X-Forwarded-Proto", proto)
		}
	}

	if r.Header.Get("X-Forwarded-Port") == "" {
		r.Header.Set("X-Forwarded-Port", localPort(r))
	}

	fwd := r.Header.Get("Forwarded")
	if fwd == "" {
		fwd = "for=" + remoteIP + "; proto=" + proto
	}
	if cfg.LocalIP != "" {
		fwd += "; by=" + cfg.LocalIP
	}
	r.Header.Set("Forwarded", fwd)

	if cfg.TLSHeader != "" {
		if r.TLS != nil {
			r.Header.Set(cfg.TLSHeader, cfg.TLSHeaderValue)
		} else {
			r.Header.Del(cfg.TLSHeader)
		}
	}

	return nil
}

// scheme derives the request scheme used on the initial
// request first from headers and then from the connection
// using the following heuristic:
//
// If either X-Forwarded-Proto or Forwarded is set then use
// its value to set the other header. If both headers are
// set do not modify the protocol. If none are set derive
// the protocol from the connection.
func scheme(r *http.Request) string {
	xfp := r.Header.Get("X-Forwarded-Proto")
	fwd := r.Header.Get("Forwarded")
	switch {
	case xfp != "" && fwd == "":
		return xfp

	case fwd != "" && xfp == "":
		p := strings.SplitAfterN(fwd, "proto=", 2)
		if len(p) == 1 {
			break
		}
		n := strings.IndexRune(p[1], ';')
		if n >= 0 {
			return p[1][:n]
		}
		return p[1]
	}

	ws := r.Header.Get("Upgrade") == "websocket"
	switch {
	case ws && r.TLS != nil:
		return "wss"
	case ws && r.TLS == nil:
		return "ws"
	case r.TLS != nil:
		return "https"
	default:
		return "http"
	}
}

func localPort(r *http.Request) string {
	if r == nil {
		return ""
	}
	n := strings.Index(r.Host, ":")
	if n > 0 && n < len(r.Host)-1 {
		return r.Host[n+1:]
	}
	if r.TLS != nil {
		return "443"
	}
	return "80"
}
