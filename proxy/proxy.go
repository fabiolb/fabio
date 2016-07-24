package proxy

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/eBay/fabio/config"

	gometrics "github.com/rcrowley/go-metrics"
)

// Proxy is a dynamic reverse proxy.
type Proxy struct {
	tr       http.RoundTripper
	cfg      config.Proxy
	requests gometrics.Timer
	log      *log.Logger
}

func New(tr http.RoundTripper, cfg config.Proxy) *Proxy {
	var l *log.Logger
	if cfg.Log.Enable {
		file, err := os.OpenFile(cfg.Log.File.Path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("[FATAL] ", err)
		}
		l = log.New(file, "", 0)
	}

	return &Proxy{
		tr:       tr,
		cfg:      cfg,
		requests: gometrics.GetOrRegisterTimer("requests", gometrics.DefaultRegistry),
		log:      l,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ShuttingDown() {
		http.Error(w, "shutting down", http.StatusServiceUnavailable)
		return
	}

	t := target(r)
	if t == nil {
		w.WriteHeader(p.cfg.NoRouteStatus)
		return
	}

	if err := addHeaders(r, p.cfg); err != nil {
		http.Error(w, "cannot parse "+r.RemoteAddr, http.StatusInternalServerError)
		return
	}

	var h http.Handler
	switch {
	case r.Header.Get("Upgrade") == "websocket":
		h = newRawProxy(t.URL)

		// To use the filtered proxy use
		// h = newWSProxy(t.URL)
	default:
		h = newHTTPProxy(t.URL, p.tr)
	}

	start := time.Now()
	h.ServeHTTP(w, r)
	p.requests.UpdateSince(start)
	t.Timer.UpdateSince(start)

	if p.cfg.Log.Enable {
		logger(p.log, p.cfg.Log.Format, start, r)
	}
}
