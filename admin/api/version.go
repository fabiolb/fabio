package api

import (
	"fmt"
	"net/http"
)

var Version string

func HandleVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, Version)
}
