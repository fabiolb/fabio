package exit

import (
	"os"
	"os/signal"
	"syscall"
)

// Listen will listen to OS signals (currently SIGINT, SIGKILL, SIGTERM)
// and will trigger the callback when signal are received from OS
func Listen(fn func(os.Signal)) {
	go func() {
		// we use buffered to mitigate losing the signal
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt, os.Kill, syscall.SIGTERM)

		sig := <-sigchan
		if fn != nil {
			fn(sig)
		}
	}()
}
