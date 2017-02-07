package mdllog

import (
	"testing"
)

func TestAll(t *testing.T) {
	Trace.Println("Trace logs")
	Debug.Println("Debug logs")
	Info.Println("Info logs")
	Warning.Println("Warning logs")
	Error.Println("Error logs")
}
