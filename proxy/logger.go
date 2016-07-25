package proxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/eBay/fabio/config"
)

type Logger struct {
	mu      sync.Mutex // ensures atomic writes;
	out     io.Writer  // destination for output
	pattern []Pattern
	enable  bool
}

func newLogger(cfg config.Proxy) *Logger {

	if len(cfg.Log.Target) > 0 && len(cfg.Log.Format) <= 0 {
		log.Fatal("[FATAL] Invalid Logger format ", cfg.Log.Format)
	}

	flags := strings.Split(cfg.Log.Format, " ")
	c := make([]Pattern, len(flags))

	for i, flag := range flags {
		c[i] = patterns[flag]
	}

	switch cfg.Log.Target {
	case "stdout":
		log.Printf("[INFO] Output logger to stdout")
		return &Logger{out: os.Stdout, pattern: c, enable: true}
	case "":
		log.Printf("[INFO] Logger disabled")
	default:
		log.Fatal("[FATAL] Invalid Logger target ", cfg.Log.Target)
	}
	return &Logger{}
}

type Pattern func(w io.Writer, t time.Time, r *http.Request)

const BufferSize = 1024

var (
	pool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, BufferSize))
		},
	}
	patterns = map[string]Pattern{
		"remote_addr": func(w io.Writer, t time.Time, r *http.Request) {
			io.WriteString(w, r.RemoteAddr[:strings.Index(r.RemoteAddr, ":")])
		},
		"time": func(w io.Writer, t time.Time, r *http.Request) {
			io.WriteString(w, t.Format(time.RFC3339))
		},
		"request": func(w io.Writer, t time.Time, r *http.Request) {
			io.WriteString(w, "\""+r.Method+" "+r.URL.Path+" "+r.Proto+"\"")
		},
		"body_bytes_sent": func(w io.Writer, t time.Time, r *http.Request) {
			io.WriteString(w, fmt.Sprintf("%d", uint64(r.ContentLength)))
		},
		"http_referer": func(w io.Writer, t time.Time, r *http.Request) {
			io.WriteString(w, r.Referer())
		},
		"http_user_agent": func(w io.Writer, t time.Time, r *http.Request) {
			io.WriteString(w, r.UserAgent())
		},
		"server_name": func(w io.Writer, t time.Time, r *http.Request) {
			io.WriteString(w, r.Host)
		},
		"proxy_endpoint": func(w io.Writer, t time.Time, r *http.Request) {
			io.WriteString(w, r.URL.Host)
		},
		"response_time": func(w io.Writer, t time.Time, r *http.Request) {
			io.WriteString(w, fmt.Sprintf("%.4f", time.Since(t).Seconds()))
		},
		"request_args": func(w io.Writer, t time.Time, r *http.Request) {
			io.WriteString(w, r.URL.RawQuery)
		},
	}
)

func (l *Logger) logger(t time.Time, r *http.Request) {

	if !l.enable {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	b := pool.Get().(*bytes.Buffer)
	b.Reset()

	for _, p := range l.pattern {
		p(b, t, r)
		b.WriteRune(' ')
	}
	b.Truncate(b.Len() - 1) //drop last space
	b.WriteByte('\n')
	l.out.Write(b.Bytes())
	pool.Put(b)
}
