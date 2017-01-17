package api

import (
	"fmt"
	"net/http"
)

type VersionHandler struct {
	Version string
}

func (h *VersionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, h.Version)
}
