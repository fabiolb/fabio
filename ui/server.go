package ui

import "net/http"

// Addr contains the host:port of the UI endpoint
var configPath string

func Start(addr, cfgpath string) error {
	configPath = cfgpath
	http.HandleFunc("/", handleRoute)
	return http.ListenAndServe(addr, nil)
}
