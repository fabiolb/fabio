// Package exit allows to register callbacks which are called on program exit.
package exit

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var wg sync.WaitGroup

// quit channel is closed to cleanup exit listeners.
var quit = make(chan bool)

// Listen registers an exit handler which is called on
// SIGINT/SIGKILL/SIGTERM or when Exit/Fatal/Fatalf is called.
func Listen(fn func(os.Signal)) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		// we use buffered to mitigate losing the signal
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt, os.Kill, syscall.SIGTERM)

		var sig os.Signal
		select {
		case sig = <-sigchan:
		case <-quit:
		}
		if fn != nil {
			fn(sig)
		}
	}()
}

// stubbed out for testing
var osExit = os.Exit

// Exit terminates the application via os.Exit but
// waits for all exit handlers to complete before
// calling os.Exit.
func Exit(code int) {
	defer func() { recover() }() // don't panic if close(quit) is called concurrently
	close(quit)
	wg.Wait()
	osExit(code)
}

// Fatal is a replacement for log.Fatal which will trigger
// the closure of all registered exit handlers and waits
// for their completion and then call os.Exit(1).
func Fatal(v ...interface{}) {
	log.Print(v...)
	Exit(1)
}

// Fatalf is a replacement for log.Fatalf and behaves like Fatal.
func Fatalf(format string, v ...interface{}) {
	log.Printf(format, v...)
	Exit(1)
}

// Wait waits for all exit handlers to complete.
func Wait() {
	wg.Wait()
}
