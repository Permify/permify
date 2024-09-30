package balancer

import (
	"errors"
)

var (
	// ErrSubConnMissing indicates that a SubConn (sub-connection) was expected but not found.
	ErrSubConnMissing = errors.New("sub-connection is missing or not found")
	// ErrSubConnResetFailure indicates an error occurred while trying to reset the SubConn.
	ErrSubConnResetFailure = errors.New("failed to reset the sub-connection")
)
