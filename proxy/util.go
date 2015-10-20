package proxy

import (
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/route"
)

// addHeaders adds/updates headers in request
//
// * add/update `Forwarded` header
// * ClientIPHeader == "X-Forwarded-For": add/update `X-Forwarded-For` header
// * ClientIPHeader != "": Set header with that name to <remote ip>
// * TLS connection: Set header with name from `cfg.TLSHeader` to `cfg.TLSHeaderValue`
//
func addHeaders(r *http.Request, cfg config.Proxy) error {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return errors.New("cannot parse " + r.RemoteAddr)
	}

	if cfg.ClientIPHeader != "" {
		r.Header.Set(cfg.ClientIPHeader, remoteIP)
	}

	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" && cfg.LocalIP != "" {
		r.Header.Set("X-Forwarded-For", xff+", "+cfg.LocalIP)
	}

	fwd := r.Header.Get("Forwarded")
	if fwd == "" {
		fwd = "for=" + remoteIP
		if r.TLS != nil {
			fwd += "; proto=https"
		} else {
			fwd += "; proto=http"
		}
	}
	if cfg.LocalIP != "" {
		fwd += "; by=" + cfg.LocalIP
	}
	r.Header.Set("Forwarded", fwd)

	if cfg.TLSHeader != "" && r.TLS != nil {
		r.Header.Set(cfg.TLSHeader, cfg.TLSHeaderValue)
	}

	return nil
}

// target looks up a target URL for the request from the current routing table.
func target(r *http.Request) *route.Target {
	t := route.GetTable().Lookup(r, r.Header.Get("trace"))
	if t == nil {
		log.Print("[WARN] No route for ", r.URL)
	}
	return t
}
