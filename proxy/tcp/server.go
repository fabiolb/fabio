package tcp

import (
	"context"
	"net"
	"sync"
	"time"
)

type Handler interface {
	ServeTCP(in net.Conn) error
}

// Server implements a generic TCP server.
type Server struct {
	Addr         string
	Handler      Handler
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	mu        sync.Mutex
	listeners []net.Listener
	conns     map[int64]net.Conn
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
			id:           time.Now().UnixNano(),
			c:            c,
			ReadTimeout:  s.ReadTimeout,
			WriteTimeout: s.WriteTimeout,
		}
		s.mu.Lock()
		if s.conns == nil {
			s.conns = map[int64]net.Conn{}
		}
		s.conns[c.(*conn).id] = c
		s.mu.Unlock()
		go s.Handler.ServeTCP(c)
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
	for _, c := range s.conns {
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
	id           int64
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
