package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const BufferSize = 1024

type Pattern func(w io.Writer, t time.Time, r *http.Request)

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
		"http_x_forwarded_for": func(w io.Writer, t time.Time, r *http.Request) {
			io.WriteString(w, r.Header.Get("X-Forwarded-For"))
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

type Logger struct {
	p []Pattern

	// w is the log destination
	w io.Writer

	// mu guards w
	mu sync.Mutex
}

func New(w io.Writer, format string) (*Logger, error) {
	p, err := parse(format)

	if err != nil {
		return nil, err
	}

	//format can't empty
	if p == nil || len(p) == 0 {
		return nil, fmt.Errorf("Invalid Logger format %s", format)
	}

	return &Logger{w: w, p: p}, nil
}

func parse(format string) ([]Pattern, error) {
	var pp []Pattern

	for _, f := range strings.Fields(format) {
		p := patterns[f]
		if p == nil {
			return nil, fmt.Errorf("Invalid log field \"%s\"", f)
		}
		pp = append(pp, p)

	}

	return pp, nil
}

func (l *Logger) Log(t time.Time, r *http.Request) {
	b := pool.Get().(*bytes.Buffer)
	b.Reset()

	for _, p := range l.p {
		p(b, t, r)
		b.WriteRune(' ')
	}
	b.Truncate(b.Len() - 1) //drop last space
	b.WriteRune('\n')

	l.mu.Lock()
	l.w.Write(b.Bytes())
	l.mu.Unlock()
	pool.Put(b)
}
