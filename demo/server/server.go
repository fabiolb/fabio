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
//   ./server -addr 127.0.0.1:5000 -name svc-a -prefix /foo,/bar
//   ./server -addr 127.0.0.1:5001 -name svc-b -prefix /baz,/bar
//
//   # websocket server
//   ./server -addr 127.0.0.1:6000 -name ws-a -prefix /echo1,/echo2 -proto ws
//
//   # tcp server
//   ./server -addr 127.0.0.1:7000 -name tcp-a -proto tcp
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

	"github.com/eBay/fabio/proxy/tcp"
	"github.com/hashicorp/consul/api"
	"golang.org/x/net/websocket"
)

type Server interface {
	// embedded server methods
	ListenAndServe() error
	ListenAndServeTLS(certFile, keyFile string) error

	// consul register helpers
	Tags() []string
	Check() *api.AgentServiceCheck
}

type HTTPServer struct {
	*http.Server
	tags  []string
	check *api.AgentServiceCheck
}

func (s *HTTPServer) Check() *api.AgentServiceCheck {
	return s.check
}

func (s *HTTPServer) Tags() []string {
	return s.tags
}

type TCPServer struct {
	*tcp.Server
	tags  []string
	check *api.AgentServiceCheck
}

func (s *TCPServer) Check() *api.AgentServiceCheck {
	return s.check
}

func (s *TCPServer) Tags() []string {
	return s.tags
}

func main() {
	var addr, consul, name, prefix, proto, token, rawtags string
	var certFile, keyFile string
	var status int
	flag.StringVar(&addr, "addr", "127.0.0.1:5000", "host:port of the service")
	flag.StringVar(&consul, "consul", "127.0.0.1:8500", "host:port of the consul agent")
	flag.StringVar(&name, "name", filepath.Base(os.Args[0]), "name of the service")
	flag.StringVar(&prefix, "prefix", "", "comma-sep list of 'host/path' or ':port' prefixes to register")
	flag.StringVar(&proto, "proto", "http", "protocol for endpoints: http, ws or tcp")
	flag.StringVar(&rawtags, "tags", "", "additional tags to register in consul")
	flag.StringVar(&token, "token", "", "consul ACL token")
	flag.StringVar(&certFile, "cert", "", "path to cert file")
	flag.StringVar(&keyFile, "key", "", "path to key file")
	flag.IntVar(&status, "status", http.StatusOK, "http status code")
	flag.Parse()

	if prefix == "" {
		flag.Usage()
		os.Exit(1)
	}

	var srv Server
	switch proto {
	case "http", "ws":
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "OK")
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not found", 404)
			log.Printf("%s -> 404", r.URL)
		})

		tags := strings.Split(rawtags, ",")
		for _, p := range strings.Split(prefix, ",") {
			tags = append(tags, "urlprefix-"+p)
			switch proto {
			case "http":
				mux.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(status)
					fmt.Fprintf(w, "Serving %s from %s on %s\n", r.RequestURI, name, addr)
				})
			case "ws":
				mux.Handle(p, websocket.Handler(WSEchoServer))
			}
		}

		var check *api.AgentServiceCheck
		if certFile != "" {
			check = &api.AgentServiceCheck{TCP: addr, Interval: "2s", Timeout: "1s"}
		} else {
			check = &api.AgentServiceCheck{HTTP: "http://" + addr + "/health", Interval: "1s", Timeout: "1s"}
		}
		srv = &HTTPServer{&http.Server{Addr: addr, Handler: mux}, tags, check}

	case "tcp":
		tags := strings.Split(rawtags, ",")
		for _, p := range strings.Split(prefix, ",") {
			tags = append(tags, "urlprefix-"+p+" proto=tcp")
		}
		check := &api.AgentServiceCheck{TCP: addr, Interval: "2s", Timeout: "1s"}
		srv = &TCPServer{&tcp.Server{Addr: addr, Handler: tcp.HandlerFunc(TCPEchoHandler)}, tags, check}

	default:
		log.Fatal("Invalid protocol ", proto)
	}

	// start server
	go func() {
		var err error
		if certFile != "" {
			err = srv.ListenAndServeTLS(certFile, keyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil {
			log.Fatal(err)
		}
	}()

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
		Tags:    srv.Tags(),
		Check:   srv.Check(),
	}

	config := &api.Config{Address: consul, Scheme: "http", Token: token}
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Agent().ServiceRegister(service); err != nil {
		log.Fatal(err)
	}
	log.Printf("Registered %s service %q in consul with tags %q", proto, name, strings.Join(srv.Tags(), ","))

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
