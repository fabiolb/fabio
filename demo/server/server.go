// Package server provides a sample HTTP/Websocket server which registers
// itself in consul using one or more url prefixes to demonstrate and
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
// If the protocol is set to "ws" the registered endpoints function
// as websocket echo servers.
//
// Example:
//
//   # http server
//   ./server -addr 127.0.0.1:5000 -name svc-a -prefix /foo -prefix /bar
//   ./server -addr 127.0.0.1:5001 -name svc-b -prefix /baz -prefix /bar
//   ./server -addr 127.0.0.1:5002 -name svc-c -prefix "/gogl redirect=301,https://www.google.de/"
//
//   # https server
//   ./server -addr 127.0.0.1:5000 -name svc-a -proto https -certFile ... -keyFile ... -prefix /foo
//   ./server -addr 127.0.0.1:5000 -name svc-a -proto https -certFile ... -keyFile ... -prefix "/foo tlsskipverify=true"
//
//   # websocket server
//   ./server -addr 127.0.0.1:6000 -name ws-a -proto ws -prefix /echo1 -prefix /echo2
//
//   # tcp server
//   ./server -addr 127.0.0.1:7000 -name tcp-a -proto tcp -prefix :1234
//
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fabiolb/fabio/proxy/tcp"

	"github.com/hashicorp/consul/api"
	"golang.org/x/net/websocket"
)

type Args struct {
	addr     string
	consul   string
	name     string
	proto    string
	token    string
	certFile string
	keyFile  string
	status   int
	prefixes []string
	tags     []string
}

type stringsVar []string

func (s *stringsVar) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s stringsVar) String() string {
	return strings.Join(s, " ")
}

func main() {
	var args Args

	flag.StringVar(&args.addr, "addr", "127.0.0.1:5000", "host:port of the service")
	flag.StringVar(&args.consul, "consul", "127.0.0.1:8500", "host:port of the consul agent")
	flag.StringVar(&args.name, "name", filepath.Base(os.Args[0]), "name of the service")
	flag.StringVar(&args.proto, "proto", "http", "protocol for endpoints: http, ws or tcp")
	flag.StringVar(&args.token, "token", "", "consul ACL token")
	flag.StringVar(&args.certFile, "cert", "", "path to cert file")
	flag.StringVar(&args.keyFile, "key", "", "path to key file")
	flag.IntVar(&args.status, "status", http.StatusOK, "http status code")
	flag.Var((*stringsVar)(&args.prefixes), "prefix", "'host/path' or ':port' prefix to register. Can be specified multiple times")
	flag.Var((*stringsVar)(&args.tags), "tags", "additional tags to register in consul. Can be specified multiple times")
	flag.Parse()

	if len(args.prefixes) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if (args.proto == "https" || args.proto == "wss") && args.certFile == "" {
		log.Fatalf("Proto %s requires a certificate. Please provide -cert/-key", args.proto)
	}

	type server interface {
		ListenAndServe() error
		ListenAndServeTLS(certFile, keyFile string) error
	}

	var srv server
	var tags []string
	var check *api.AgentServiceCheck
	switch args.proto {
	case "http", "https", "ws", "wss":
		srv, tags, check = newHTTPServer(args)
	case "tcp":
		srv, tags, check = newTCPServer(args)
	default:
		log.Fatal("Invalid protocol ", args.proto)
	}

	// start server
	go func() {
		var err error
		if args.certFile != "" {
			err = srv.ListenAndServeTLS(args.certFile, args.keyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil {
			log.Fatal(err)
		}
	}()

	// get host and port as string/int
	host, portstr, err := net.SplitHostPort(args.addr)
	if err != nil {
		log.Fatal(err)
	}
	port, err := strconv.Atoi(portstr)
	if err != nil {
		log.Fatal(err)
	}

	// register service with health check
	serviceID := args.name + "-" + args.addr
	service := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    args.name,
		Port:    port,
		Address: host,
		Tags:    tags,
		Check:   check,
	}

	config := &api.Config{Address: args.consul, Scheme: "http", Token: args.token}
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Agent().ServiceRegister(service); err != nil {
		log.Fatal(err)
	}
	log.Printf("Registered %s service %q in consul with tags %q", args.proto, args.name, strings.Join(tags, ","))

	// run until we get a signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	// deregister service
	if err := client.Agent().ServiceDeregister(serviceID); err != nil {
		log.Fatal(err)
	}
	log.Printf("Deregistered service %q in consul", args.name)
}

func newHTTPServer(args Args) (*http.Server, []string, *api.AgentServiceCheck) {
	addr, proto, tags := args.addr, args.proto, args.tags

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", 404)
		log.Printf("%s -> 404", r.URL)
	})

	for _, p := range args.prefixes {
		uri := strings.Fields(p)[0]
		switch proto {
		case "http", "https":
			mux.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
				scheme := "http"
				if r.TLS != nil {
					scheme = "https"
				}
				w.WriteHeader(args.status)
				fmt.Fprintf(w, "Serving %s via %s from %s on %s\n", r.RequestURI, scheme, args.name, addr)
			})
		case "ws", "wss":
			mux.Handle(uri, websocket.Handler(WSEchoServer))
		}

		tag := "urlprefix-" + p
		if proto == "https" || proto == "wss" {
			tag += " proto=https"
		}
		tags = append(tags, tag)
	}

	checkScheme := "http"
	if args.certFile != "" {
		checkScheme = "https"
	}
	check := &api.AgentServiceCheck{
		HTTP:          checkScheme + "://" + addr + "/health",
		Interval:      "1s",
		Timeout:       "1s",
		TLSSkipVerify: true,
	}
	return &http.Server{Addr: addr, Handler: mux}, tags, check
}

func WSEchoServer(ws *websocket.Conn) {
	addr := ws.LocalAddr().String()
	pfx := []byte("[" + addr + "] ")

	log.Printf("ws connect on %s", addr)

	// the following could be done with io.Copy(ws, ws)
	// but I want to add some meta data
	var msg = make([]byte, 1024)
	for {
		n, err := ws.Read(msg)
		if err != nil && err != io.EOF {
			log.Printf("ws error on %s. %s", addr, err)
			break
		}
		_, err = ws.Write(append(pfx, msg[:n]...))
		if err != nil && err != io.EOF {
			log.Printf("ws error on %s. %s", addr, err)
			break
		}
	}
	log.Printf("ws disconnect on %s", addr)
}

func newTCPServer(args Args) (*tcp.Server, []string, *api.AgentServiceCheck) {
	tags := args.tags
	for _, p := range args.prefixes {
		tags = append(tags, "urlprefix-"+p+" proto=tcp")
	}
	check := &api.AgentServiceCheck{TCP: args.addr, Interval: "2s", Timeout: "1s"}
	return &tcp.Server{Addr: args.addr, Handler: tcp.HandlerFunc(TCPEchoHandler)}, tags, check
}

func TCPEchoHandler(c net.Conn) error {
	defer c.Close()

	addr := c.LocalAddr().String()
	_, err := fmt.Fprintf(c, "[%s] Welcome\n", addr)
	if err != nil {
		return err
	}

	for {
		line, _, err := bufio.NewReader(c).ReadLine()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(c, "[%s] %s\n", addr, string(line))
		if err != nil {
			return err
		}
	}
}
