package api

import (
	"net/http"

	"github.com/fabiolb/fabio/config"
)

type ConfigHandler struct {
	Config *config.Config
}

func (h *ConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, config.Sanitise(h.Config))
}
