package ui

import "net/http"

// Addr contains the host:port of the UI endpoint
var configURL string

func Start(addr, cfgurl string) error {
	configURL = cfgurl
	http.HandleFunc("/", handleRoute)
	return http.ListenAndServe(addr, nil)
}
