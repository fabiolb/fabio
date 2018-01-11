package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/fabiolb/fabio/registry"
)

type ManualPathsHandler struct {
	Prefix string
}

func (h *ManualPathsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// we need this for testing.
	// under normal circumstances this is never nil
	if registry.Default == nil {
		return
	}

	switch r.Method {
	case "GET":
		paths, err := registry.Default.ManualPaths()
		if err != nil {
			log.Print("[ERROR] ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for i, p := range paths {
			paths[i] = strings.TrimPrefix(p, h.Prefix)
		}
		writeJSON(w, r, paths)
		return

	default:
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
	}
}
