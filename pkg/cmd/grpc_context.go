package cmd

import (
	"context"
	"time"
)

const defaultGRPCTimeout = 30 * time.Second

// newGRPCCallContext returns a context canceled after defaultGRPCTimeout or when parent is done.
func newGRPCCallContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, defaultGRPCTimeout)
}
