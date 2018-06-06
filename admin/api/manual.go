package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fabiolb/fabio/registry"
)

// ManualHandler provides a fetch and update handler for the manual overrides api.
type ManualHandler struct {
	BasePath string
}

type manual struct {
	Value   string `json:"value"`
	Version uint64 `json:"version,string"`
}

func (h *ManualHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// we need this for testing.
	// under normal circumstances this is never nil
	if registry.Default == nil {
		return
	}

	path := r.RequestURI[len(h.BasePath):]

	switch r.Method {
	case "GET":
		value, version, err := registry.Default.ReadManual(path)
		if err != nil {
			log.Print("[ERROR] ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, r, manual{value, version})
		return

	case "PUT":
		var m manual
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			log.Print("[ERROR] ", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		ok, err := registry.Default.WriteManual(path, m.Value, m.Version)
		if err != nil {
			log.Print("[ERROR] ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !ok {
			http.Error(w, "version mismatch", http.StatusConflict)
			return
		}

	default:
		http.Error(w, "not allowed", http.StatusMethodNotAllowed)
	}
}
