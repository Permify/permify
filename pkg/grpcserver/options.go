package grpcserver

import (
	"net"
)

// Option - Type for server options
type Option func(*Server)

// Port -
func Port(port string) Option {
	return func(s *Server) {
		s.addr = net.JoinHostPort("", port)
	}
}
