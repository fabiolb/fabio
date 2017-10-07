package logger

import (
	"fmt"
	"io"
	"strings"
	"sync/atomic"
)

// LevelWriter implements a simplistic levelled log writer which supports
// TRACE, DEBUG, INFO, WARN, ERROR and FATAL. The log level can be changed at
// runtime.
type LevelWriter struct {
	w     io.Writer
	level atomic.Value // string
}

func NewLevelWriter(w io.Writer, level string) *LevelWriter {
	lw := &LevelWriter{w: w}
	if !lw.SetLevel(level) {
		panic(fmt.Sprintf("invalid log level %s", level))
	}
	return lw
}

// prefixLen is a sample prefix of the normal log output to determine the
// position of the opening bracket '['. It might be better to detect this but
// this will do until the format of the log output changes.
var prefixLen = len("2017/10/07 20:50:53 [")

func (w *LevelWriter) Write(b []byte) (int, error) {
	// check if the log line starts with the prefix
	if len(b) < prefixLen || b[prefixLen-1] != '[' {
		return fmt.Fprint(w.w, "invalid log msg: ", string(b))
	}

	// determine the level by looking at the character after the opening
	// bracket.
	level := rune(b[prefixLen]) // T, D, I, W, E, or F

	// w.level contains the characters of all the allowed levels so we can just
	// check whether the level character is in that set.
	if strings.ContainsRune(w.level.Load().(string), level) {
		return w.w.Write(b)
	}
	return 0, nil
}

// SetLevel sets the log level to the new value and returns true
// if that was successful.
func (w *LevelWriter) SetLevel(s string) bool {
	// levels contains the first character of the levels in descending order
	const levels = "TDIWEF"
	switch strings.ToUpper(s) {
	case "TRACE":
		w.level.Store(levels[0:])
		return true
	case "DEBUG":
		w.level.Store(levels[1:])
		return true
	case "INFO":
		w.level.Store(levels[2:])
		return true
	case "WARN":
		w.level.Store(levels[3:])
		return true
	case "ERROR":
		w.level.Store(levels[4:])
		return true
	default:
		return false
	}
}

// Level returns the current log level.
func (w *LevelWriter) Level() string {
	l := w.level.Load().(string)
	switch l[0] {
	case 'T':
		return "TRACE"
	case 'D':
		return "DEBUG"
	case 'I':
		return "INFO"
	case 'W':
		return "WARN"
	case 'E':
		return "ERROR"
	case 'F':
		return "FATAL"
	default:
		return "???" + l + "???"
	}
}
