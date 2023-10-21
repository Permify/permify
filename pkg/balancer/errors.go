package balancer

import (
	"fmt"
)

var (
	// ErrSubConnMissing indicates that a SubConn (sub-connection) was expected but not found.
	ErrSubConnMissing = fmt.Errorf("sub-connection is missing or not found")
	// ErrSubConnResetFailure indicates an error occurred while trying to reset the SubConn.
	ErrSubConnResetFailure = fmt.Errorf("failed to reset the sub-connection")
)
