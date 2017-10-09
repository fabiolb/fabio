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
	w         io.Writer
	level     atomic.Value // string
	prefixLen int
}

// NewLevelWriter creates a new leveled writer for the given output and a
// default level. Prefix is the string that is expected before the opening
// bracket and usually depends on the chosen log format. For the default log
// format prefix should be set to "2017/01/01 00:00:00 " whereby only the
// format and the spaces are relevant but not the date and time itself.
func NewLevelWriter(w io.Writer, level, prefix string) *LevelWriter {
	lw := &LevelWriter{w: w, prefixLen: len(prefix)}
	if !lw.SetLevel(level) {
		panic(fmt.Sprintf("invalid log level %s", level))
	}
	return lw
}

func (w *LevelWriter) Write(b []byte) (int, error) {
	// check if the log line starts with the prefix
	if len(b) < w.prefixLen+2 || b[w.prefixLen] != '[' {
		return fmt.Fprint(w.w, "invalid log msg: ", string(b))
	}

	// determine the level by looking at the character after the opening
	// bracket.
	level := rune(b[w.prefixLen+1]) // T, D, I, W, E, or F

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
	case "FATAL":
		w.level.Store(levels[5:])
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
