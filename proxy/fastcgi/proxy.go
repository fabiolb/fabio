package fastcgi

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fabiolb/fabio/config"
)

var (
	headerNameReplacer = strings.NewReplacer(" ", "_", "-", "_")
)

type Proxy struct {
	root        string
	index       string
	stripPrefix string
	upstream    string
	config      *config.Config
	dialFunc    func(string) (FCGIBackend, error)
}

// FCGIBackend describes the capabilities offered by
// FastCGI bakcned server
type FCGIBackend interface {
	// Options proxies HTTP OPTION request to FCGI backend
	Options(parameters map[string]string) (resp *http.Response, err error)

	// Head proxies HTTP HEAD request to FCGI backend
	Head(parameters map[string]string) (resp *http.Response, err error)

	// Get proxies HTTP GET request to FCGI backend
	Get(parameters map[string]string) (resp *http.Response, err error)

	// Post proxies HTTP Post request to FCGI backend
	Post(parameters map[string]string, method string, bodyType string, body io.Reader, contentLength int64) (resp *http.Response, err error)

	// SetReadTimeout sets the maximum time duration the
	// connection will wait to read full response. If the
	// deadline is reached before the response is read in
	// full, the client receives a gateway timeout error.
	SetReadTimeout(time.Duration) error

	// SetSendTimeout sets the maximum time duration the
	// connection will wait to send full request to FCGI backend.
	// If the deadline is reached before the request is sent
	// completely, the client receives a gateway timeout error.
	SetSendTimeout(time.Duration) error

	// Stderr returns any error produced by FCGI backend
	// while processing the request.
	Stderr() string

	// Close closes the connection.
	Close()
}

func NewProxy(cfg *config.Config, upstream string) *Proxy {
	return &Proxy{
		root:     cfg.FastCGI.Root,
		index:    cfg.FastCGI.Index,
		upstream: upstream,
		config:   cfg,
		dialFunc: Connect,
	}
}

// Connect to FCGI backend
func Connect(upstream string) (FCGIBackend, error) {
	return Dial("tcp", upstream)
}

func (p *Proxy) SetRoot(root string) {
	p.root = root
}

func (p *Proxy) SetStripPathPrefix(prefix string) {
	p.stripPrefix = prefix
}

func (p *Proxy) SetIndex(index string) {
	p.index = index
}

func (p *Proxy) stripPathPrefix(path string) string {
	return strings.TrimPrefix(path, p.stripPrefix)
}

func (p *Proxy) ensureIndexFile(path string) string {
	prefix := ""
	if strings.HasPrefix(path, "/") {
		prefix = "/"
	}

	if strings.HasPrefix(path, prefix+p.index) {
		return path
	}

	return filepath.Join(p.index, path)
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fpath := p.stripPathPrefix(r.URL.Path)
	env, err := p.buildEnv(r, p.ensureIndexFile(fpath))
	if err != nil {
		log.Printf("[WARN] failed to create fastcgi environment. %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for x, y := range env {
		log.Printf("[INFO] >>>>  %s  =>  %s", x, y)
	}

	fcgiBackend, err := p.dialFunc(p.upstream)
	if err != nil {
		log.Printf("[WARN] failed to connect with FastCGI upstream. %s", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer fcgiBackend.Close()

	if err := fcgiBackend.SetReadTimeout(p.config.FastCGI.ReadTimeout); err != nil {
		log.Printf("[ERROR] failed to set connection read timeout. %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if err := fcgiBackend.SetSendTimeout(p.config.FastCGI.WriteTimeout); err != nil {
		log.Printf("[ERROR] failed to set connection write timeout. %s", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var resp *http.Response

	contentLength := r.ContentLength
	if contentLength == 0 {
		contentLength, _ = strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
	}

	switch r.Method {
	case "HEAD":
		resp, err = fcgiBackend.Head(env)
	case "GET":
		resp, err = fcgiBackend.Get(env)
	case "OPTIONS":
		resp, err = fcgiBackend.Options(env)
	default:
		resp, err = fcgiBackend.Post(env, r.Method, r.Header.Get("Content-Type"), r.Body, contentLength)
	}

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			log.Printf("[ERROR] FastCGI upstream timed out during request. %s", err)
			http.Error(w, "Gateway Timeout", http.StatusGatewayTimeout)
			return
		} else if err != io.EOF {
			log.Printf("[ERROR] failed to read response from FastCGI upstream. %s", err)
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}
	}

	writeHeader(w, resp)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("[ERROR] failed to write response body. %s", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	if errOut := fcgiBackend.Stderr(); errOut != "" {
		log.Printf("[WARN] Error from FastCGI upstream: %s", errOut)
		return
	}
}

func writeHeader(w http.ResponseWriter, r *http.Response) {
	for key, vals := range r.Header {
		for _, val := range vals {
			w.Header().Add(key, val)
		}
	}
	w.WriteHeader(r.StatusCode)
}

func (p Proxy) buildEnv(r *http.Request, fpath string) (map[string]string, error) {
	absPath := filepath.Join(p.root, fpath)

	// Separate remote IP and port; more lenient than net.SplitHostPort
	var ip, port string
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx > -1 {
		ip = r.RemoteAddr[:idx]
		port = r.RemoteAddr[idx+1:]
	} else {
		ip = r.RemoteAddr
	}

	username := ""
	if r.URL.User != nil {
		username = r.URL.User.Username()
	}

	// Remove [] from IPv6 addresses
	ip = strings.TrimPrefix(ip, "[")
	ip = strings.TrimSuffix(ip, "]")

	splitPos := p.splitPos(fpath)
	if splitPos == -1 {
		return nil, fmt.Errorf("cannot split path on %s", p.config.FastCGI.SplitPath)
	}

	// Request has the extension; path was split successfully
	docURI := fpath[:splitPos+len(p.config.FastCGI.SplitPath)]
	pathInfo := fpath[splitPos+len(p.config.FastCGI.SplitPath):]
	scriptName := fpath
	scriptFilename := absPath

	// Strip PATH_INFO from SCRIPT_NAME
	scriptName = strings.TrimSuffix(scriptName, pathInfo)

	// Some variables are unused but cleared explicitly to prevent
	// the parent environment from interfering.
	env := map[string]string{
		// Variables defined in CGI 1.1 spec
		"AUTH_TYPE":         "", // Not used
		"CONTENT_LENGTH":    r.Header.Get("Content-Length"),
		"CONTENT_TYPE":      r.Header.Get("Content-Type"),
		"GATEWAY_INTERFACE": "CGI/1.1",
		"PATH_INFO":         pathInfo,
		"QUERY_STRING":      r.URL.RawQuery,
		"REMOTE_ADDR":       ip,
		"REMOTE_HOST":       ip, // For speed, remote host lookups disabled
		"REMOTE_PORT":       port,
		"REMOTE_IDENT":      "", // Not used
		"REMOTE_USER":       username,
		"REQUEST_METHOD":    r.Method,
		"SERVER_NAME":       r.URL.Hostname(),
		"SERVER_PORT":       r.URL.Port(),
		"SERVER_PROTOCOL":   r.Proto,
		"SERVER_SOFTWARE":   "fabio",

		// Other variables
		"DOCUMENT_ROOT":   p.root,
		"DOCUMENT_URI":    docURI,
		"HTTP_HOST":       r.Host, // added here, since not always part of headers
		"REQUEST_URI":     p.stripPathPrefix(r.URL.RequestURI()),
		"SCRIPT_FILENAME": scriptFilename,
		"SCRIPT_NAME":     scriptName,
	}

	// compliance with the CGI specification requires that
	// PATH_TRANSLATED should only exist if PATH_INFO is defined.
	// Info: https://www.ietf.org/rfc/rfc3875 Page 14
	if env["PATH_INFO"] != "" {
		env["PATH_TRANSLATED"] = filepath.Join(p.root, pathInfo) // Info: http://www.oreilly.com/openbook/cgi/ch02_04.html
	}

	// Some web apps rely on knowing HTTPS or not
	if r.TLS != nil {
		env["HTTPS"] = "on"
	}

	// Add all HTTP headers to env variables
	for field, val := range r.Header {
		header := strings.ToUpper(field)
		header = headerNameReplacer.Replace(header)
		env["HTTP_"+header] = strings.Join(val, ", ")
	}
	return env, nil
}

func (p *Proxy) splitPos(path string) int {
	return strings.Index(strings.ToLower(path), strings.ToLower(p.config.FastCGI.SplitPath))
}
