package admin

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/fabiolb/fabio/admin/api"
	"github.com/fabiolb/fabio/admin/ui"
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/proxy"
)

// Server provides the HTTP server for the admin UI and API.
type Server struct {
	Cfg      *config.Config
	Access   string
	Color    string
	Title    string
	Path     string
	Version  string
	Commands string
}

// ListenAndServe starts the admin server.
func (s *Server) ListenAndServe(l config.Listen, tlscfg *tls.Config) error {
	return proxy.ListenAndServeHTTP(l, s.handler(), tlscfg)
}

func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()
	p := strings.TrimRight(s.Path, "/")

	switch s.Access {
	case "ro":
		mux.HandleFunc(p+"/api/paths", forbidden)
		mux.HandleFunc(p+"/api/manual", forbidden)
		mux.HandleFunc(p+"/api/manual/", forbidden)
		mux.HandleFunc(p+"/manual", forbidden)
		mux.HandleFunc(p+"/manual/", forbidden)
	case "rw":
		// for historical reasons the configured config path starts with a '/'
		// but Consul treats all KV paths without a leading slash.
		pathsPrefix := strings.TrimPrefix(s.Cfg.Registry.Consul.KVPath, "/")
		mux.Handle(p+"/api/paths", &api.ManualPathsHandler{Prefix: pathsPrefix})
		mux.Handle(p+"/api/manual", &api.ManualHandler{BasePath: p + "/api/manual"})
		mux.Handle(p+"/api/manual/", &api.ManualHandler{BasePath: p + "/api/manual"})
		mux.Handle(p+"/manual", &ui.ManualHandler{
			BasePath: p + "/manual",
			Color:    s.Color,
			Title:    s.Title,
			Version:  s.Version,
			Commands: s.Commands,
			Path:     p,
		})
		mux.Handle(p+"/manual/", &ui.ManualHandler{
			BasePath: p + "/manual",
			Color:    s.Color,
			Title:    s.Title,
			Version:  s.Version,
			Commands: s.Commands,
			Path:     p,
		})
	}

	mux.Handle(p+"/api/config", &api.ConfigHandler{Config: s.Cfg})
	mux.Handle(p+"/api/routes", &api.RoutesHandler{})
	mux.Handle(p+"/api/version", &api.VersionHandler{Version: s.Version})
	mux.Handle(p+"/routes", &ui.RoutesHandler{Color: s.Color, Title: s.Title, Version: s.Version, Path: p, RoutingTable: s.Cfg.UI.RoutingTable})
	mux.HandleFunc(p+"/health", handleHealth)

	mux.Handle(p+"/assets/", http.StripPrefix(p, http.FileServer(http.FS(ui.Static))))
	mux.HandleFunc(p+"/favicon.ico", http.NotFound)

	mux.Handle(p+"/", http.RedirectHandler(p+"/routes", http.StatusSeeOther))
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func forbidden(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Forbidden", http.StatusForbidden)
}
