package admin

import (
	"net/http"

	"github.com/eBay/fabio/admin/api"
	"github.com/eBay/fabio/admin/ui"
)

func Start(addr, version string) error {
	ui.Version = version
	http.HandleFunc("/api/manual", api.HandleManual)
	http.HandleFunc("/api/routes", api.HandleRoutes)
	http.HandleFunc("/manual", ui.HandleManual)
	http.HandleFunc("/routes", ui.HandleRoutes)
	http.Handle("/", http.RedirectHandler("/routes", http.StatusSeeOther))
	return http.ListenAndServe(addr, nil)
}
