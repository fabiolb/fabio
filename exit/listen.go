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
// SIGINT/SIGTERM or when Exit/Fatal/Fatalf is called.
// SIGHUP is ignored since that is usually used for
// triggering a reload of configuration which isn't
// supported but shouldn't kill the process either.
func Listen(fn func(os.Signal)) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			sigchan := make(chan os.Signal, 1)
			signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

			var sig os.Signal
			select {
			case sig = <-sigchan:
				switch sig {
				case syscall.SIGHUP:
					log.Print("[INFO] Caught SIGHUP. Ignoring")
					continue
				case os.Interrupt:
					log.Print("[INFO] Caught SIGINT. Exiting")
				case syscall.SIGTERM:
					log.Print("[INFO] Caught SIGTERM. Exiting")
				default:
					// fallthrough in case we forgot to add a switch clause.
					log.Printf("[INFO] Caught signal %v. Exiting", sig)
				}
			case <-quit:
			}
			if fn != nil {
				fn(sig)
			}
			return
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
