package logger

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	ts := time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		req    *http.Request
		format string
		out    string
	}{
		{
			&http.Request{
				RequestURI: "/",
				Header:     http.Header{"X-Forwarded-For": {"3.3.3.3"}},
				RemoteAddr: "2.2.2.2:666",
				URL:        &url.URL{},
				Method:     "GET",
				Proto:      "HTTP/1.1",
			},
			"remote_addr time request body_bytes_sent http_x_forwarded_for",
			"2.2.2.2 2016-01-01T00:00:00Z \"GET  HTTP/1.1\" 0 3.3.3.3\n",
		},
	}

	for i, tt := range tests {
		b := new(bytes.Buffer)
		l, err := New(b, tt.format)

		if err != nil {
			t.Fatalf("%d: got %v want nil", i, err)
		}

		l.Log(ts, tt.req)

		if got, want := string(b.Bytes()), tt.out; got != want {
			t.Errorf("%d: got %q want %q", i, got, want)
		}
	}

}
