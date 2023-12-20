// Package httpserver implements HTTP server.
package httpserver

import (
	"context"
	"net/http"
	"time"
)

const (
	_defaultReadTimeout     = 5 * time.Second
	_defaultWriteTimeout    = 5 * time.Second
	_defaultAddr            = ":3476"
	_defaultShutdownTimeout = 3 * time.Second
)

// Server - Structure for server instance
type Server struct {
	server          *http.Server
	notify          chan error
	shutdownTimeout time.Duration
}

// New - Creates new server instance
func New(handler http.Handler, opts ...Option) *Server {
	httpServer := &http.Server{
		Handler:      handler,
		ReadTimeout:  _defaultReadTimeout,
		WriteTimeout: _defaultWriteTimeout,
		Addr:         _defaultAddr,
	}

	s := &Server{
		server:          httpServer,
		notify:          make(chan error, 1),
		shutdownTimeout: _defaultShutdownTimeout,
	}

	// custom options
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Run() {
	go func() {
		s.notify <- s.server.ListenAndServe()
		close(s.notify)
	}()
}

// Notify - Server notify
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown - shutdown the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}
