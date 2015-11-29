package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/eBay/fabio/registry"
)

type manual struct {
	Value   string `json:"value"`
	Version uint64 `json:"version,string"`
}

func HandleManual(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		value, version, err := registry.DefaultBackend.ReadManual()
		if err != nil {
			log.Printf("[ERROR] ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, r, manual{value, version})
		return

	case "PUT":
		var m manual
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			log.Printf("[ERROR] ", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		ok, err := registry.DefaultBackend.WriteManual(m.Value, m.Version)
		if err != nil {
			log.Printf("[ERROR] ", err)
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
