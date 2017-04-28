package logger

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"
	"text/template"
	"time"
)

func TestParse(t *testing.T) {
	fields := map[string]field{
		"$a": func(b *bytes.Buffer, e *Event) {
			b.WriteString("aa")
		},
		"$b": func(b *bytes.Buffer, e *Event) {
			b.WriteString("bb")
		},
	}
	req := &http.Request{
		Header: http.Header{
			"User-Agent":      {"Mozilla Firefox"},
			"X-Forwarded-For": {"3.3.3.3"},
		},
	}
	tests := []struct {
		format string
		out    string
	}{
		{"", ""},
		{"$a", "aa\n"},
		{"$a $b", "aa bb\n"},
		{"$a \"$header.User-Agent\"", "aa \"Mozilla Firefox\"\n"},
	}

	for i, tt := range tests {
		p, err := parse(tt.format, fields)
		if err != nil {
			t.Errorf("%d: got %v want nil", i, err)
			continue
		}
		var b bytes.Buffer
		p.write(&b, &Event{Start: time.Time{}, End: time.Time{}, Request: req})
		if got, want := string(b.Bytes()), tt.out; got != want {
			t.Errorf("%d: got %q want %q", i, got, want)
		}
	}
}

func TestLog(t *testing.T) {
	rurl := mustParse("http://foo.com/?q=x")
	uurl := mustParse("http://7.8.9.0:5678/foo?q=x")
	start := time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)
	e := &Event{
		Start: start,
		End:   start.Add(123456789 * time.Nanosecond),
		Request: &http.Request{
			RequestURI: rurl.RequestURI(),
			Header: http.Header{
				"User-Agent":      {"Mozilla Firefox"},
				"Referer":         {"http://foo.com/"},
				"X-Forwarded-For": {"3.3.3.3"},
			},
			RemoteAddr: "2.2.2.2:666",
			Host:       rurl.Host,
			URL:        rurl,
			Method:     "GET",
			Proto:      "HTTP/1.1",
		},
		Response: &http.Response{
			StatusCode:    200,
			ContentLength: 1234,
			Header:        http.Header{"foo": []string{"bar"}},
			Request: &http.Request{
				RemoteAddr: "5.6.7.8:1234",
			},
		},
		RequestURL:      rurl,
		UpstreamAddr:    uurl.Host,
		UpstreamService: "svc-a",
		UpstreamURL:     uurl,
	}

	tests := []struct {
		format string
		out    string
	}{
		{"$header.Referer", "http://foo.com/\n"},
		{"$header.X-Forwarded-For", "3.3.3.3\n"},
		{"$header.user-agent", "Mozilla Firefox\n"},
		{"$remote_addr", "2.2.2.2:666\n"},
		{"$remote_host", "2.2.2.2\n"},
		{"$remote_port", "666\n"},
		{"$request", "GET /?q=x HTTP/1.1\n"},
		{"$request_args", "q=x\n"},
		{"$request_host", "foo.com\n"}, // TODO(fs): is this correct?
		{"$request_method", "GET\n"},
		{"$request_proto", "HTTP/1.1\n"},
		{"$request_scheme", "http\n"},
		{"$request_uri", "/?q=x\n"},
		{"$request_url", "http://foo.com/?q=x\n"},
		{"$response_body_size", "1234\n"},
		{"$response_status", "200\n"},
		{"$response_time_ms", "0.123\n"},       // TODO(fs): is this correct?
		{"$response_time_ns", "0.123456789\n"}, // TODO(fs): is this correct?
		{"$response_time_us", "0.123456\n"},    // TODO(fs): is this correct?
		{"$time_common", "01/Jan/2016:00:00:00 +0000\n"},
		{"$time_rfc3339", "2016-01-01T00:00:00Z\n"},
		{"$time_rfc3339_ms", "2016-01-01T00:00:00.123Z\n"},
		{"$time_rfc3339_ns", "2016-01-01T00:00:00.123456789Z\n"},
		{"$time_rfc3339_us", "2016-01-01T00:00:00.123456Z\n"},
		{"$time_unix_ms", "1451606400123\n"},
		{"$time_unix_ns", "1451606400123456789\n"},
		{"$time_unix_us", "1451606400123456\n"},
		{"$upstream_addr", "7.8.9.0:5678\n"},
		{"$upstream_host", "7.8.9.0\n"},
		{"$upstream_port", "5678\n"},
		{"$upstream_request_scheme", "http\n"},
		{"$upstream_request_uri", "/foo?q=x\n"},
		{"$upstream_request_url", "http://7.8.9.0:5678/foo?q=x\n"},
		{"$upstream_service", "svc-a\n"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			b := new(bytes.Buffer)

			l, err := New(b, tt.format)
			if err != nil {
				t.Fatalf("got %v want nil", err)
			}

			l.Log(e)
			if got, want := string(b.Bytes()), tt.out; got != want {
				t.Errorf("got %q want %q", got, want)
			}
		})
	}
}

func TestAtoi(t *testing.T) {
	tests := []struct {
		i   int64
		pad int
		s   string
	}{
		{i: 0, pad: 0, s: "0"},
		{i: 1, pad: 0, s: "1"},
		{i: -1, pad: 0, s: "-1"},
		{i: 12345, pad: 0, s: "12345"},
		{i: -12345, pad: 0, s: "-12345"},
		{i: 9223372036854775807, pad: 0, s: "9223372036854775807"},
		{i: -9223372036854775807, pad: 0, s: "-9223372036854775807"},

		{i: 0, pad: 5, s: "00000"},
		{i: 1, pad: 5, s: "00001"},
		{i: -1, pad: 5, s: "-00001"},
		{i: 12345, pad: 5, s: "12345"},
		{i: -12345, pad: 5, s: "-12345"},
		{i: 9223372036854775807, pad: 5, s: "9223372036854775807"},
		{i: -9223372036854775807, pad: 5, s: "-9223372036854775807"},
	}

	for i, tt := range tests {
		var b bytes.Buffer
		atoi(&b, tt.i, tt.pad)
		if got, want := string(b.Bytes()), tt.s; got != want {
			t.Errorf("%d: got %q want %q", i, got, want)
		}
	}
}

func BenchmarkLog(b *testing.B) {
	start := time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)
	e := &Event{
		Start: start,
		End:   start.Add(100 * time.Millisecond),
		Request: &http.Request{
			RequestURI: "/?q=x",
			Header: http.Header{
				"User-Agent":      {"Mozilla Firefox"},
				"Referer":         {"http://foo.com/"},
				"X-Forwarded-For": {"3.3.3.3"},
			},
			RemoteAddr: "2.2.2.2:666",
			Host:       "foo.com",
			URL: &url.URL{
				Path:     "/",
				RawQuery: "?q=x",
				Host:     "proxy host",
			},
			Method: "GET",
			Proto:  "HTTP/1.1",
		},
		Response: &http.Response{
			StatusCode:    200,
			ContentLength: 1234,
			Header:        http.Header{"foo": []string{"bar"}},
			Request: &http.Request{
				RemoteAddr: "5.6.7.8:1234",
			},
		},
		UpstreamAddr: mustParse("http://7.8.9.0:5678/foo").Host,
	}

	// benchmark the custom parser and text/template
	// to explain why there is a custom approach.
	// The custom parser is 8x faster and has zero allocs.
	//
	// BenchmarkLog/my_parser-8         	 1000000	      2326 ns/op	       0 B/op	       0 allocs/op
	// BenchmarkLog/go_text/template-8  	  100000	     19026 ns/op	     848 B/op	      76 allocs/op
	b.Run("custom parser", func(b *testing.B) {
		var keys []string
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		format := strings.Join(keys, " ")

		l, err := New(ioutil.Discard, format)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Log(e)
		}
	})
	b.Run("text/template", func(b *testing.B) {
		// simulate the text template approach by using
		// the same number of fields as for the other parser
		// but using the same value.
		tmpl := ""
		for i := 0; i < len(fields); i++ {
			tmpl += "{{.Req.RemoteAddr}}"
		}
		t := template.Must(template.New("log").Parse(tmpl))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			t.Execute(ioutil.Discard, e)
		}
	})
}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
