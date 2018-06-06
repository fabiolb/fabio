package proxy

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/fabiolb/fabio/config"
)

// addResponseHeaders adds/updates headers in the response
func addResponseHeaders(w http.ResponseWriter, r *http.Request, cfg config.Proxy) error {
	if r.TLS != nil && cfg.STSHeader.MaxAge > 0 {
		sts := "max-age=" + i32toa(int32(cfg.STSHeader.MaxAge))
		if cfg.STSHeader.Subdomains {
			sts += "; includeSubdomains"
		}
		if cfg.STSHeader.Preload {
			sts += "; preload"
		}
		w.Header().Set("Strict-Transport-Security", sts)
	}

	return nil
}

// addHeaders adds/updates headers in request
//
// * add/update `Forwarded` header
// * add X-Forwarded-Proto header, if not present
// * add X-Real-Ip, if not present
// * ClientIPHeader != "": Set header with that name to <remote ip>
// * TLS connection: Set header with name from `cfg.TLSHeader` to `cfg.TLSHeaderValue`
//
func addHeaders(r *http.Request, cfg config.Proxy, stripPath string) error {
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

	if r.Header.Get("X-Forwarded-Host") == "" && r.Host != "" {
		r.Header.Set("X-Forwarded-Host", r.Host)
	}

	if stripPath != "" {
		r.Header.Set("X-Forwarded-Prefix", stripPath)
	}

	fwd := r.Header.Get("Forwarded")
	if fwd == "" {
		fwd = "for=" + remoteIP + "; proto=" + proto
	}
	if cfg.LocalIP != "" {
		fwd += "; by=" + cfg.LocalIP
	}
	if r.Proto != "" {
		fwd += "; httpproto=" + strings.ToLower(r.Proto)
	}
	if r.TLS != nil && r.TLS.Version > 0 {
		v := tlsver[r.TLS.Version]
		if v == "" {
			v = uint16base16(r.TLS.Version)
		}
		fwd += "; tlsver=" + v
	}
	if r.TLS != nil && r.TLS.CipherSuite != 0 {
		fwd += "; tlscipher=" + uint16base16(r.TLS.CipherSuite)
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

var tlsver = map[uint16]string{
	tls.VersionSSL30: "ssl30",
	tls.VersionTLS10: "tls10",
	tls.VersionTLS11: "tls11",
	tls.VersionTLS12: "tls12",
}

var digit16 = []byte("0123456789abcdef")

// uint16base64 is a faster version of fmt.Sprintf("0x%04x", n)
//
// BenchmarkUint16Base16/fmt.Sprintf-8         	10000000	       154 ns/op	       8 B/op	       2 allocs/op
// BenchmarkUint16Base16/uint16base16-8        	50000000	        35.0 ns/op	       8 B/op	       1 allocs/op
func uint16base16(n uint16) string {
	b := []byte("0x0000")
	b[5] = digit16[n&0x000f]
	b[4] = digit16[n&0x00f0>>4]
	b[3] = digit16[n&0x0f00>>8]
	b[2] = digit16[n&0xf000>>12]
	return string(b)
}

// i32toa is a faster implentation of strconv.Itoa() without importing another library
// https://stackoverflow.com/a/39444005
func i32toa(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
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
