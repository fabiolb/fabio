package api

import (
	"encoding/json"
	"net/http"

	"github.com/eBay/fabio/registry"
	"github.com/eBay/fabio/mdllog"
)

// ManualHandler provides a fetch and update handler for the manual overrides api.
type ManualHandler struct{}

type manual struct {
	Value   string `json:"value"`
	Version uint64 `json:"version,string"`
}

func (h *ManualHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		value, version, err := registry.Default.ReadManual()
		if err != nil {
			mdllog.Error.Print("[ERROR] ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, r, manual{value, version})
		return

	case "PUT":
		var m manual
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			mdllog.Error.Print("[ERROR] ", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		ok, err := registry.Default.WriteManual(m.Value, m.Version)
		if err != nil {
			mdllog.Error.Print("[ERROR] ", err)
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
