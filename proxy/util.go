package proxy

import (
	"errors"
	"log"
	"net"
	"net/http"

	"github.com/eBay/fabio/config"
	"github.com/eBay/fabio/route"
	"strings"
)

// addHeaders adds/updates headers in request
func addHeaders(r *http.Request, cfg config.Proxy) error {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return errors.New("cannot parse " + r.RemoteAddr)
	}

	addRemoteIpHeaders(r, cfg, remoteIP)
	addXForwardedHeaders(r, cfg, remoteIP)
	addForwardedHeader(r, cfg, remoteIP)
	addTLSHeader(r, cfg, remoteIP)

	return nil
}

// Set Configurable ClientIPHeader here, but
// don't set X-Forwarded-For here, because it will be set by the Golang Reverse Proxy Handler, later
func addRemoteIpHeaders(r *http.Request, cfg config.Proxy, remoteIP string) {
	if cfg.ClientIPHeader != "" && cfg.ClientIPHeader != "X-Forwarded-For" {
		r.Header.Set(cfg.ClientIPHeader, remoteIP)
	}
	if cfg.ClientIPHeader != "X-Real-Ip" {
		r.Header.Set("X-Real-Ip", remoteIP)
	}
}

// Sets X-Forwarded-Host, X-Forwarded-Proto
func addXForwardedHeaders(r *http.Request, cfg config.Proxy, remoteIP string) {
	r.Header.Set("X-Forwarded-Host", r.Host)
	if r.TLS != nil {
		r.Header.Set("X-Forwarded-Proto", "https")
	} else {
		r.Header.Set("X-Forwarded-Proto", "http")
	}
}

// * add/update `Forwarded` header defined by rfc7239
func addForwardedHeader(r *http.Request, cfg config.Proxy, remoteIP string) {
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
	if !strings.Contains(fwd, "host=") {
		fwd += "; host=" + r.Host
	}

	r.Header.Set("Forwarded", fwd)
}

// * TLS connection: Set header with name from `cfg.TLSHeader` to `cfg.TLSHeaderValue`
func addTLSHeader(r *http.Request, cfg config.Proxy, remoteIP string) {
	if cfg.TLSHeader != "" && r.TLS != nil {
		r.Header.Set(cfg.TLSHeader, cfg.TLSHeaderValue)
	}
}

// target looks up a target URL for the request from the current routing table.
func target(r *http.Request) *route.Target {
	t := route.GetTable().Lookup(r, r.Header.Get("trace"))
	if t == nil {
		log.Print("[WARN] No route for ", r.Host, r.URL)
	}
	return t
}
