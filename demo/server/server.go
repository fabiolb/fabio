// Package server provides a sample HTTP web server which registers
// itself in consul using one or more URL prefixes to demonstrate and
// test the automatic fabio routing table update.
//
// During startup the server performs the following steps:
//
// * Add a handler for each prefix which provides a unique
//   response for that instance and endpoint
// * Add a `/health` handler for the consul health check
// * Register the service in consul with the listen address,
//   a health check under the given name and with one `urlprefix-`
//   tag per prefix
// * Install a signal handler to deregister the service on exit
//
// Example:
//
//   ./server -addr 127.0.0.1:5000 -name svc-a -prefix /foo,/bar
//   ./server -addr 127.0.0.1:6000 -name svc-b -prefix /baz,/bar
//
// This used to be hosted under https://github.com/magiconair/fabio-example
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eBay/fabio/_third_party/github.com/hashicorp/consul/api"
)

func main() {
	var addr, consul, name, prefix string
	flag.StringVar(&addr, "addr", "127.0.0.1:5000", "host:port of the service")
	flag.StringVar(&consul, "consul", "127.0.0.1:8500", "host:port of the consul agent")
	flag.StringVar(&name, "name", filepath.Base(os.Args[0]), "name of the service")
	flag.StringVar(&prefix, "prefix", "", "comma-sep list of host/path prefixes to register")
	flag.Parse()

	if prefix == "" {
		flag.Usage()
		os.Exit(1)
	}

	// register prefixes
	prefixes := strings.Split(prefix, ",")
	for _, p := range prefixes {
		http.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Serving %s from %s on %s\n", r.RequestURI, name, addr)
		})
	}

	// start http server
	go func() {
		log.Printf("Listening on %s serving %s", addr, prefix)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	// register consul health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	// build urlprefix-host/path tag list
	// e.g. urlprefix-/foo, urlprefix-/bar, ...
	var tags []string
	for _, p := range prefixes {
		tags = append(tags, "urlprefix-"+p)
	}

	// get host and port as string/int
	host, portstr, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatal(err)
	}
	port, err := strconv.Atoi(portstr)
	if err != nil {
		log.Fatal(err)
	}

	// register service with health check
	serviceID := name + "-" + addr
	service := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    name,
		Port:    port,
		Address: host,
		Tags:    tags,
		Check: &api.AgentServiceCheck{
			HTTP:     "http://" + addr + "/health",
			Interval: "1s",
			Timeout:  "1s",
		},
	}

	config := &api.Config{Address: consul, Scheme: "http"}
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Agent().ServiceRegister(service); err != nil {
		log.Fatal(err)
	}
	log.Printf("Registered service %q in consul with tags %q", name, strings.Join(tags, ","))

	// run until we get a signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	// deregister service
	if err := client.Agent().ServiceDeregister(serviceID); err != nil {
		log.Fatal(err)
	}
	log.Printf("Deregistered service %q in consul", name)
}
