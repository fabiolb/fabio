package bgp

import (
	"context"
	"log"
	"log/slog"
	"strings"

	"github.com/fabiolb/fabio/logger"
)

// bgpLogHandler wraps fabio's logger to work with slog
type bgpLogHandler struct {
	attrs  []slog.Attr
	groups []string
}

func newBGPLogHandler() *bgpLogHandler {
	return &bgpLogHandler{
		attrs:  make([]slog.Attr, 0),
		groups: make([]string, 0),
	}
}

func (h *bgpLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	lw, ok := log.Writer().(*logger.LevelWriter)
	if !ok {
		return level >= slog.LevelInfo
	}

	switch lw.Level() {
	case "TRACE", "DEBUG":
		return level >= slog.LevelDebug
	case "INFO":
		return level >= slog.LevelInfo
	case "WARN":
		return level >= slog.LevelWarn
	case "ERROR", "FATAL":
		return level >= slog.LevelError
	default:
		return level >= slog.LevelInfo
	}
}

func (h *bgpLogHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String()
	msg := r.Message

	// Build the log message with attributes
	var b strings.Builder
	appendAttr := func(a slog.Attr) {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(a.Key)
		b.WriteString("=>")
		b.WriteString(a.Value.String())
	}

	r.Attrs(func(a slog.Attr) bool {
		appendAttr(a)
		return true
	})
	for _, a := range h.attrs {
		appendAttr(a)
	}

	if b.Len() > 0 {
		log.Printf("[%s] gobgpd %s %s", level, msg, b.String())
	} else {
		log.Printf("[%s] gobgpd %s", level, msg)
	}

	return nil
}

func (h *bgpLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &bgpLogHandler{
		attrs:  make([]slog.Attr, len(h.attrs)+len(attrs)),
		groups: make([]string, len(h.groups)),
	}
	copy(newHandler.attrs, h.attrs)
	copy(newHandler.attrs[len(h.attrs):], attrs)
	copy(newHandler.groups, h.groups)
	return newHandler
}

func (h *bgpLogHandler) WithGroup(name string) slog.Handler {
	newHandler := &bgpLogHandler{
		attrs:  make([]slog.Attr, len(h.attrs)),
		groups: make([]string, len(h.groups)+1),
	}
	copy(newHandler.attrs, h.attrs)
	copy(newHandler.groups, h.groups)
	newHandler.groups[len(h.groups)] = name
	return newHandler
}

// newBGPLogger creates a slog.Logger that integrates with fabio's logging system
func newBGPLogger() *slog.Logger {
	return slog.New(newBGPLogHandler())
}
