package ui

import "net/http"

// Addr contains the host:port of the UI endpoint
var configURL string
var version string

func Start(addr, cfgurl, ver string) error {
	configURL = cfgurl
	version = ver
	http.HandleFunc("/", handleRoute)
	return http.ListenAndServe(addr, nil)
}
