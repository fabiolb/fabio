// Package api provides the HTTP api.
package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func writeJSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	_, pretty := r.URL.Query()["pretty"]

	var buf []byte
	var err error
	if pretty {
		buf, err = json.MarshalIndent(v, "", "    ")
	} else {
		buf, err = json.Marshal(v)
	}

	if err != nil {
		log.Print("[ERROR] ", err)
		http.Error(w, "internal error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(buf)
}
