package admin

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/fabiolb/fabio/admin/api"
	"github.com/fabiolb/fabio/admin/ui"
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/proxy"
)

// Server provides the HTTP server for the admin UI and API.
type Server struct {
	Color    string
	Title    string
	Version  string
	Commands string
	Cfg      *config.Config
}

// ListenAndServe starts the admin server.
func (s *Server) ListenAndServe(l config.Listen, tlscfg *tls.Config) error {
	http.Handle("/api/config", &api.ConfigHandler{s.Cfg})
	http.Handle("/api/manual", &api.ManualHandler{})
	http.Handle("/api/routes", &api.RoutesHandler{})
	http.Handle("/api/version", &api.VersionHandler{s.Version})
	http.Handle("/manual", &ui.ManualHandler{Color: s.Color, Title: s.Title, Version: s.Version, Commands: s.Commands})
	http.Handle("/routes", &ui.RoutesHandler{Color: s.Color, Title: s.Title, Version: s.Version})
	http.HandleFunc("/logo.svg", ui.HandleLogo)
	http.HandleFunc("/health", handleHealth)
	http.Handle("/", http.RedirectHandler("/routes", http.StatusSeeOther))
	return proxy.ListenAndServeHTTP(l, nil, tlscfg)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}
