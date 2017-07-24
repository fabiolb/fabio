package tcp

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"
)

// Handler responds to a TCP request.
//
// ServeTCP should write responses to the in connection and close
// it on return.
type Handler interface {
	ServeTCP(in net.Conn) error
}

type HandlerFunc func(in net.Conn) error

func (f HandlerFunc) ServeTCP(in net.Conn) error {
	return f(in)
}

// Server implements a generic TCP server.
type Server struct {
	Addr         string
	Handler      Handler
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	mu        sync.Mutex
	listeners []net.Listener
	conns     map[net.Conn]bool
}

func (s *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer l.Close()
	return s.Serve(l)
}

func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	l, err := tls.Listen("tcp", s.Addr, cfg)
	if err != nil {
		return err
	}
	defer l.Close()
	return s.Serve(l)
}

func (s *Server) Serve(l net.Listener) error {
	defer l.Close()

	s.mu.Lock()
	s.listeners = append(s.listeners, l)
	s.mu.Unlock()

	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		c = &conn{
			c:            c,
			ReadTimeout:  s.ReadTimeout,
			WriteTimeout: s.WriteTimeout,
		}
		s.mu.Lock()
		if s.conns == nil {
			s.conns = map[net.Conn]bool{}
		}
		s.conns[c] = true
		s.mu.Unlock()

		go func() {
			defer func() {
				c.Close()
				s.mu.Lock()
				delete(s.conns, c)
				s.mu.Unlock()
			}()

			s.Handler.ServeTCP(c)
		}()
	}
}

func (s *Server) closeListeners() error {
	s.mu.Lock()
	for _, l := range s.listeners {
		l.Close()
	}
	s.listeners = nil
	s.mu.Unlock()
	return nil
}

func (s *Server) closeConns() error {
	s.mu.Lock()
	for c := range s.conns {
		c.Close()
	}
	s.conns = nil
	s.mu.Unlock()
	return nil
}

func (s *Server) Close() error {
	s.closeListeners()
	return s.closeConns()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.closeListeners()
	if ctx != nil {
		<-ctx.Done()
	}
	return s.closeConns()
}

// conn implements a connection which honors read and write timeouts.
type conn struct {
	c            net.Conn
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (c *conn) Read(b []byte) (int, error) {
	if c.ReadTimeout > 0 {
		c.c.SetReadDeadline(time.Now().Add(c.ReadTimeout))
	}
	return c.c.Read(b)
}

func (c *conn) Write(b []byte) (int, error) {
	if c.WriteTimeout > 0 {
		c.c.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
	}
	return c.c.Write(b)
}

func (c *conn) Close() error {
	return c.c.Close()
}

func (c *conn) LocalAddr() net.Addr {
	return c.c.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	return c.c.RemoteAddr()
}

func (c *conn) SetDeadline(t time.Time) error {
	return c.c.SetDeadline(t)
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return c.c.SetReadDeadline(t)
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	return c.c.SetWriteDeadline(t)
}
