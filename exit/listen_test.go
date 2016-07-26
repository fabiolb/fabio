package exit

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestExit(t *testing.T) {
	var b bytes.Buffer
	var exitCode int
	var sig1, sig2 bool

	flags := log.Flags()
	log.SetFlags(0)
	log.SetOutput(&b)
	defer func() {
		log.SetOutput(os.Stderr)
		log.SetFlags(flags)
	}()

	osExit = func(code int) { exitCode = code }
	defer func() { osExit = os.Exit }()

	Listen(func(os.Signal) { sig1 = true })
	Listen(func(os.Signal) { sig2 = true })

	// trigger a concurrent exit via fatal/fatalf
	// it is not guaranteed that any log output is written
	// before the application exits. This is only to test
	// that two go routines can call Fatal without causing a
	// panic.
	go Fatal("a")
	go Fatalf("b")

	// wait for listeners to return
	Wait()

	out := string(b.Bytes())
	if !strings.Contains("a\n b\n a\nb\n b\na\n", out) {
		t.Errorf("log.Fatal did not happen: %q", out)
	}
	if exitCode != 1 {
		t.Errorf("os.Exit not called")
	}
	if !sig1 || !sig2 {
		t.Errorf("signal handlers not completed")
	}
}
