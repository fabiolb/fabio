package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	fastCGIVersion      = 1
	fastCGIBeginRequest = 1
	fastCGIEndRequest   = 3
	fastCGIParams       = 4
	fastCGIStdin        = 5
	fastCGIStdout       = 6
	fastCGIStderr       = 7
	fastCGIRequestID    = 1
	fastCGIMaxContent   = 65535
	fastCGIMaxHeaders   = 64 << 10
)

// fastCGITransport uses one FastCGI responder request per TCP connection.
// ReverseProxy retains ownership of the HTTP response and closes its body.
type fastCGITransport struct {
	root                  string
	originalURI           string
	dialTimeout           time.Duration
	responseHeaderTimeout time.Duration
}

func (t fastCGITransport) RoundTrip(req *http.Request) (*http.Response, error) {
	params, err := t.params(req)
	if err != nil {
		return nil, err
	}
	requestBody := req.Body
	spooled := false
	if req.ContentLength < 0 && requestBody != nil && requestBody != http.NoBody {
		// CGI responders need CONTENT_LENGTH before STDIN. Spool chunked
		// uploads to disk so their length is known without holding them in RAM.
		var length int64
		requestBody, length, err = spoolFastCGIBody(requestBody)
		if err != nil {
			return nil, err
		}
		spooled = true
		params["CONTENT_LENGTH"] = strconv.FormatInt(length, 10)
	}

	conn, err := (&net.Dialer{Timeout: t.dialTimeout}).DialContext(req.Context(), "tcp", req.URL.Host)
	if err != nil {
		if spooled {
			requestBody.Close()
		}
		return nil, err
	}
	stopCancel := context.AfterFunc(req.Context(), func() { conn.Close() })
	body := &fastCGIBody{conn: conn, requestBody: requestBody, stopCancel: stopCancel}
	var deadlineMu sync.Mutex
	headerReceived := false

	// Read the response while sending the request: FastCGI responders may write
	// output before consuming all of STDIN, so sequential I/O can deadlock.
	go func() {
		if err := sendFastCGIRequest(conn, params, requestBody); err != nil {
			conn.Close()
		} else if t.responseHeaderTimeout > 0 {
			// Match net/http's response-header timeout: start after request
			// upload completes, and never time out a streaming response body.
			deadlineMu.Lock()
			if !headerReceived {
				conn.SetReadDeadline(time.Now().Add(t.responseHeaderTimeout))
			}
			deadlineMu.Unlock()
		}
	}()

	reader := bufio.NewReader(&fastCGIStdoutReader{conn: conn})
	status, header, err := readFastCGIHeaders(reader)
	deadlineMu.Lock()
	headerReceived = true
	conn.SetReadDeadline(time.Time{})
	deadlineMu.Unlock()
	if err != nil {
		body.Close()
		if ctxErr := req.Context().Err(); ctxErr != nil {
			return nil, ctxErr
		}
		if netErr, ok := errors.AsType[net.Error](err); ok {
			return nil, netErr
		}
		return nil, err
	}
	body.reader = reader
	contentLength := int64(-1)
	if s := header.Get("Content-Length"); s != "" {
		if n, parseErr := strconv.ParseInt(s, 10, 64); parseErr == nil && n >= 0 {
			contentLength = n
		}
	}
	return &http.Response{
		StatusCode:    status,
		Status:        strconv.Itoa(status) + " " + http.StatusText(status),
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        header,
		Body:          body,
		ContentLength: contentLength,
		Request:       req,
		Close:         true,
	}, nil
}

func (t fastCGITransport) params(req *http.Request) (map[string]string, error) {
	script := path.Clean("/" + strings.TrimPrefix(req.URL.Path, "/"))
	params := map[string]string{
		"GATEWAY_INTERFACE": "CGI/1.1",
		"REQUEST_METHOD":    req.Method,
		"QUERY_STRING":      req.URL.RawQuery,
		"REQUEST_URI":       t.originalURI,
		"SCRIPT_NAME":       script,
		"SERVER_PROTOCOL":   req.Proto,
		"SERVER_SOFTWARE":   "fabio",
		"HTTP_HOST":         req.Host,
	}
	if params["REQUEST_URI"] == "" {
		params["REQUEST_URI"] = req.URL.RequestURI()
	}
	if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
		params["HTTPS"] = "on"
		params["REQUEST_SCHEME"] = "https"
	} else {
		params["REQUEST_SCHEME"] = "http"
	}
	serverName, serverPort, err := net.SplitHostPort(req.Host)
	if err != nil {
		serverName = req.Host
		serverPort = "80"
		if params["REQUEST_SCHEME"] == "https" {
			serverPort = "443"
		}
	}
	params["SERVER_NAME"], params["SERVER_PORT"] = serverName, serverPort
	if host, port, splitErr := net.SplitHostPort(req.RemoteAddr); splitErr == nil {
		params["REMOTE_ADDR"], params["REMOTE_PORT"] = host, port
	}
	if t.root != "" {
		if !path.IsAbs(t.root) {
			return nil, errors.New("fastcgi: root must be an absolute path")
		}
		root := path.Clean(t.root)
		params["DOCUMENT_ROOT"] = root
		params["SCRIPT_FILENAME"] = path.Join(root, strings.TrimPrefix(script, "/"))
	}
	if value := req.Header.Get("Content-Type"); value != "" {
		params["CONTENT_TYPE"] = value
	}
	if req.ContentLength >= 0 {
		params["CONTENT_LENGTH"] = strconv.FormatInt(req.ContentLength, 10)
	}
	for k, values := range req.Header {
		if strings.EqualFold(k, "Content-Type") || strings.EqualFold(k, "Content-Length") || strings.EqualFold(k, "Host") || strings.EqualFold(k, "Proxy") {
			continue
		}
		name := "HTTP_" + strings.ToUpper(strings.ReplaceAll(k, "-", "_"))
		separator := ", "
		if strings.EqualFold(k, "Cookie") {
			separator = "; "
		}
		params[name] = strings.Join(values, separator)
	}
	return params, nil
}

func spoolFastCGIBody(body io.ReadCloser) (io.ReadCloser, int64, error) {
	defer body.Close()
	file, err := os.CreateTemp("", "fabio-fastcgi-")
	if err != nil {
		return nil, 0, err
	}
	length, err := io.Copy(file, body)
	if err == nil {
		_, err = file.Seek(0, io.SeekStart)
	}
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, 0, err
	}
	return &fastCGITempBody{File: file}, length, nil
}

type fastCGITempBody struct{ *os.File }

func (b *fastCGITempBody) Close() error {
	err := b.File.Close()
	if removeErr := os.Remove(b.Name()); err == nil {
		err = removeErr
	}
	return err
}

func sendFastCGIRequest(conn net.Conn, params map[string]string, body io.ReadCloser) error {
	// Role RESPONDER (1), keepConn false. The responder closes each connection.
	if err := writeFastCGIRecord(conn, fastCGIBeginRequest, []byte{0, 1, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	}
	w := fastCGIParamWriter{conn: conn, buffer: make([]byte, 0, 16<<10)}
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := w.add(key, params[key]); err != nil {
			return err
		}
	}
	if err := w.flush(); err != nil {
		return err
	}
	if err := writeFastCGIRecord(conn, fastCGIParams, nil); err != nil {
		return err
	}
	if body != nil {
		buf := make([]byte, 32<<10)
		for {
			n, err := body.Read(buf)
			if n > 0 {
				if writeErr := writeFastCGIRecord(conn, fastCGIStdin, buf[:n]); writeErr != nil {
					return writeErr
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}
	}
	return writeFastCGIRecord(conn, fastCGIStdin, nil)
}

func writeFastCGIRecord(w io.Writer, recordType byte, data []byte) error {
	for {
		n := min(len(data), fastCGIMaxContent)
		var header [8]byte
		header[0], header[1], header[3] = fastCGIVersion, recordType, fastCGIRequestID
		binary.BigEndian.PutUint16(header[4:6], uint16(n))
		if err := fastCGIWriteAll(w, header[:]); err != nil {
			return err
		}
		if n != 0 {
			if err := fastCGIWriteAll(w, data[:n]); err != nil {
				return err
			}
		}
		data = data[n:]
		if len(data) == 0 {
			return nil
		}
	}
}

func fastCGIWriteAll(w io.Writer, b []byte) error {
	for len(b) > 0 {
		n, err := w.Write(b)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		b = b[n:]
	}
	return nil
}

type fastCGIParamWriter struct {
	conn   net.Conn
	buffer []byte
}

func (w *fastCGIParamWriter) flush() error {
	if len(w.buffer) == 0 {
		return nil
	}
	if err := writeFastCGIRecord(w.conn, fastCGIParams, w.buffer); err != nil {
		return err
	}
	w.buffer = w.buffer[:0]
	return nil
}

func (w *fastCGIParamWriter) writeString(s string) error {
	for len(s) > 0 {
		space := cap(w.buffer) - len(w.buffer)
		if space == 0 {
			if err := w.flush(); err != nil {
				return err
			}
			space = cap(w.buffer)
		}
		n := min(len(s), space)
		w.buffer = append(w.buffer, s[:n]...)
		s = s[n:]
	}
	return nil
}

func fastCGILength(n int) string {
	if n < 128 {
		return string([]byte{byte(n)})
	}
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(n)|0x80000000)
	return string(b[:])
}

func (w *fastCGIParamWriter) add(name, value string) error {
	if len(name) > 0x7fffffff || len(value) > 0x7fffffff {
		return errors.New("fastcgi: parameter too large")
	}
	for _, s := range []string{fastCGILength(len(name)), fastCGILength(len(value)), name, value} {
		if err := w.writeString(s); err != nil {
			return err
		}
	}
	return nil
}

// fastCGIStdoutReader strips record framing and discards STDERR without
// buffering the response body. END_REQUEST terminates the stream.
type fastCGIStdoutReader struct {
	conn      net.Conn
	remaining int
	padding   int
	ended     bool
}

func (r *fastCGIStdoutReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	for {
		if r.ended {
			return 0, io.EOF
		}
		if r.remaining > 0 {
			n, err := r.conn.Read(p[:min(len(p), r.remaining)])
			r.remaining -= n
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return n, err
		}
		if r.padding > 0 {
			if _, err := io.CopyN(io.Discard, r.conn, int64(r.padding)); err != nil {
				return 0, err
			}
			r.padding = 0
		}
		var header [8]byte
		if _, err := io.ReadFull(r.conn, header[:]); err != nil {
			return 0, err
		}
		if header[0] != fastCGIVersion || binary.BigEndian.Uint16(header[2:4]) != fastCGIRequestID {
			return 0, errors.New("fastcgi: invalid response record")
		}
		content := int(binary.BigEndian.Uint16(header[4:6]))
		r.padding = int(header[6])
		switch header[1] {
		case fastCGIStdout:
			r.remaining = content
		case fastCGIStderr:
			if _, err := io.CopyN(io.Discard, r.conn, int64(content)); err != nil {
				return 0, err
			}
		case fastCGIEndRequest:
			if content != 8 {
				return 0, errors.New("fastcgi: malformed END_REQUEST")
			}
			var end [8]byte
			if _, err := io.ReadFull(r.conn, end[:]); err != nil {
				return 0, err
			}
			if end[4] != 0 {
				return 0, fmt.Errorf("fastcgi: protocol status %d", end[4])
			}
			r.ended = true
			return 0, io.EOF
		default:
			return 0, fmt.Errorf("fastcgi: unexpected response type %d", header[1])
		}
	}
}

func readFastCGIHeaders(r *bufio.Reader) (int, http.Header, error) {
	var raw bytes.Buffer
	for {
		line, err := r.ReadSlice('\n')
		if raw.Len()+len(line) > fastCGIMaxHeaders {
			return 0, nil, errors.New("fastcgi: response headers too large")
		}
		raw.Write(line)
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		if err != nil {
			return 0, nil, fmt.Errorf("fastcgi: incomplete response headers: %w", err)
		}
		if bytes.Equal(line, []byte("\r\n")) || bytes.Equal(line, []byte("\n")) {
			break
		}
	}
	parsed, err := textproto.NewReader(bufio.NewReader(&raw)).ReadMIMEHeader()
	if err != nil {
		return 0, nil, fmt.Errorf("fastcgi: invalid response headers: %w", err)
	}
	header := http.Header(parsed)
	status := http.StatusOK
	if _, exists := header["Status"]; exists {
		value := header.Get("Status")
		fields := strings.Fields(value)
		if len(fields) == 0 {
			return 0, nil, errors.New("fastcgi: invalid response status")
		}
		code, err := strconv.Atoi(fields[0])
		if err != nil || code < 100 || code > 999 {
			return 0, nil, errors.New("fastcgi: invalid response status")
		}
		status = code
	} else if header.Get("Location") != "" {
		status = http.StatusFound
	}
	header.Del("Status")
	return status, header, nil
}

type fastCGIBody struct {
	reader      io.Reader
	conn        net.Conn
	requestBody io.ReadCloser
	stopCancel  func() bool
	once        sync.Once
}

func (b *fastCGIBody) Read(p []byte) (int, error) {
	n, err := b.reader.Read(p)
	if err != nil {
		b.Close()
	}
	return n, err
}

func (b *fastCGIBody) Close() error {
	b.once.Do(func() {
		b.stopCancel()
		b.conn.Close()
		if b.requestBody != nil {
			b.requestBody.Close()
		}
	})
	return nil
}
