package bgp

import (
	"fmt"
	"log"
	"strings"

	"github.com/fabiolb/fabio/exit"
	"github.com/fabiolb/fabio/logger"

	bgplog "github.com/osrg/gobgp/v3/pkg/log"
)

type bgpLogger struct{}

func (l bgpLogger) Panic(msg string, fields bgplog.Fields) {
	exit.Fatal(convertMsgFields("FATAL", msg, fields))
}

func (l bgpLogger) Fatal(msg string, fields bgplog.Fields) {
	exit.Fatal(convertMsgFields("FATAL", msg, fields))
}

func (l bgpLogger) Error(msg string, fields bgplog.Fields) {
	log.Printf("%s", convertMsgFields("ERROR", msg, fields))
}

func (l bgpLogger) Warn(msg string, fields bgplog.Fields) {
	log.Printf("%s", convertMsgFields("WARN", msg, fields))
}

func (l bgpLogger) Info(msg string, fields bgplog.Fields) {
	log.Printf("%s", convertMsgFields("INFO", msg, fields))
}

func (l bgpLogger) Debug(msg string, fields bgplog.Fields) {
	log.Printf("%s", convertMsgFields("DEBUG", msg, fields))
}

func (l bgpLogger) SetLevel(level bgplog.LogLevel) {
	// noop
}

func (l bgpLogger) GetLevel() bgplog.LogLevel {
	lw, ok := log.Writer().(*logger.LevelWriter)
	if !ok {
		return bgplog.InfoLevel
	}

	switch lw.Level() {
	case "TRACE":
		return bgplog.TraceLevel
	case "DEBUG":
		return bgplog.DebugLevel
	case "INFO":
		return bgplog.InfoLevel
	case "WARN":
		return bgplog.WarnLevel
	case "ERROR":
		return bgplog.ErrorLevel
	case "FATAL":
		return bgplog.FatalLevel
	default:
		return bgplog.InfoLevel
	}
}

func convertMsgFields(level, msg string, fields bgplog.Fields) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s] gobgpd %s", level, msg)
	for k, v := range fields {
		fmt.Fprintf(&b, " %s=>%v", k, v)
	}
	return b.String()
}
