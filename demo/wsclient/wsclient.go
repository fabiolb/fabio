// Package wsclient implements a simple web socket client
// which reads lines from stdin and sends them to the
// websocket url.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/net/websocket"
)

func main() {
	var url, origin string
	flag.StringVar(&url, "url", "ws://127.0.0.1:9999/echo", "websocket URL")
	flag.StringVar(&origin, "origin", "http://localhost/", "origin header")
	flag.Parse()

	if url == "" {
		flag.Usage()
		os.Exit(1)
	}

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		var msg = make([]byte, 512)
		for {
			n, err := ws.Read(msg)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("R: %s\nS: ", msg[:n])
		}
	}()

	fmt.Print("S: ")
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		if _, err := ws.Write(sc.Bytes()); err != nil {
			log.Fatal(err)
		}
	}
}
