package api

import "net/http"

type ConfigHandler struct {
	Config any
}

func (h *ConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, h.Config)
}
