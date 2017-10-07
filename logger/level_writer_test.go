package logger

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestLevelWriter(t *testing.T) {
	input := []string{
		"2017/01/01 00:00:00 [TRACE] a",
		"2017/01/01 00:00:00 [DEBUG] a",
		"2017/01/01 00:00:00 [INFO] a",
		"2017/01/01 00:00:00 [WARN] a",
		"2017/01/01 00:00:00 [ERROR] a",
		"2017/01/01 00:00:00 [FATAL] a",
	}
	tests := []struct {
		level string
		out   []string
	}{
		{"TRACE", input},
		{"DEBUG", input[1:]},
		{"INFO", input[2:]},
		{"WARN", input[3:]},
		{"ERROR", input[4:]},
		{"FATAL", input[5:]},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			var b bytes.Buffer
			w := NewLevelWriter(&b, tt.level, "2017/01/01 00:00:00 ")
			for _, s := range input {
				if _, err := w.Write([]byte(s + "\n")); err != nil {
					t.Fatal("w.Write:", err)
				}
			}
			out := strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
			if got, want := out, tt.out; !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v want %#v", got, want)
			}
		})
	}
}
