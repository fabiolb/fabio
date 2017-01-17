package api

import "net/http"

type ConfigHandler struct {
	Config interface{}
}

func (h *ConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, h.Config)
}
