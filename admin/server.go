package admin

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/fabiolb/fabio/admin/api"
	"github.com/fabiolb/fabio/admin/ui"
	_ "github.com/fabiolb/fabio/admin/ui/statik"
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/proxy"
	"github.com/rakyll/statik/fs"
)

// Server provides the HTTP server for the admin UI and API.
type Server struct {
	Access   string
	Color    string
	Title    string
	Version  string
	Commands string
	Cfg      *config.Config
}

// ListenAndServe starts the admin server.
func (s *Server) ListenAndServe(l config.Listen, tlscfg *tls.Config) error {
	return proxy.ListenAndServeHTTP(l, s.handler(), tlscfg)
}

func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()

	switch s.Access {
	case "ro":
		mux.HandleFunc("/api/paths", forbidden)
		mux.HandleFunc("/api/manual", forbidden)
		mux.HandleFunc("/api/manual/", forbidden)
		mux.HandleFunc("/manual", forbidden)
		mux.HandleFunc("/manual/", forbidden)
	case "rw":
		// for historical reasons the configured config path starts with a '/'
		// but Consul treats all KV paths without a leading slash.
		pathsPrefix := strings.TrimPrefix(s.Cfg.Registry.Consul.KVPath, "/")
		mux.Handle("/api/paths", &api.ManualPathsHandler{Prefix: pathsPrefix})
		mux.Handle("/api/manual", &api.ManualHandler{BasePath: "/api/manual"})
		mux.Handle("/api/manual/", &api.ManualHandler{BasePath: "/api/manual"})
		mux.Handle("/manual", &ui.ManualHandler{
			BasePath: "/manual",
			Color:    s.Color,
			Title:    s.Title,
			Version:  s.Version,
			Commands: s.Commands,
		})
		mux.Handle("/manual/", &ui.ManualHandler{
			BasePath: "/manual",
			Color:    s.Color,
			Title:    s.Title,
			Version:  s.Version,
			Commands: s.Commands,
		})
	}

	mux.Handle("/api/config", &api.ConfigHandler{Config: s.Cfg})
	mux.Handle("/api/routes", &api.RoutesHandler{})
	mux.Handle("/api/version", &api.VersionHandler{Version: s.Version})
	mux.Handle("/routes", &ui.RoutesHandler{Color: s.Color, Title: s.Title, Version: s.Version})
	mux.HandleFunc("/health", handleHealth)

	statikFS, err := fs.New()
	if err != nil {
		// This error is always a result of invalid, generated zip data. That
		// should never happen and would let TestAdminServerAccess fail before
		// fabio is released.
		panic(err)
	}
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(statikFS)))
	mux.HandleFunc("/favicon.ico", http.NotFound)

	mux.Handle("/", http.RedirectHandler("/routes", http.StatusSeeOther))
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func forbidden(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Forbidden", http.StatusForbidden)
}
