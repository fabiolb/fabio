// Package assert provides a simple assert framework.
package assert

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// Equal provides an assertEqual function
func Equal(t *testing.T) func(got, want interface{}) {
	return EqualDepth(t, 1, "")
}

func EqualDepth(t *testing.T, calldepth int, desc string) func(got, want interface{}) {
	return func(got, want interface{}) {
		_, file, line, _ := runtime.Caller(calldepth)
		if !reflect.DeepEqual(got, want) {
			fmt.Printf("\t%s:%d: %s: got %v want %v\n", filepath.Base(file), line, desc, got, want)
			t.Fail()
		}
	}
}
