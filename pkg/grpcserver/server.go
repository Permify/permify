package grpcserver

import (
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	_defaultAddr            = ":3478"
	_defaultShutdownTimeout = 3 * time.Second
)

// Server - Structure for server instance
type Server struct {
	addr            string
	listener        net.Listener
	Server          *grpc.Server
	notify          chan error
	shutdownTimeout time.Duration
}

// New - Creates new Server
func New(opts ...Option) *Server {
	sr := grpc.NewServer()

	s := &Server{
		Server:          sr,
		notify:          make(chan error, 1),
		shutdownTimeout: _defaultShutdownTimeout,
		addr:            _defaultAddr,
	}

	// custom options
	for _, opt := range opts {
		opt(s)
	}

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		panic(err)
	}

	s.listener = listener
	reflection.Register(s.Server)

	return s
}

// Run -
func (s *Server) Run() {
	go func() {
		s.notify <- s.Server.Serve(s.listener)
		close(s.notify)
	}()
}

// Notify -
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown -.
func (s *Server) Shutdown() error {
	s.Server.GracefulStop()
	return nil
}
