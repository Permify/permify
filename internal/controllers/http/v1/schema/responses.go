package schema

import (
	"github.com/Permify/permify/pkg/dsl/schema"
)

// WriteResponse -
type WriteResponse struct {
	Version string `json:"version"`
}

// ReadResponse -
type ReadResponse struct {
	Entities map[string]schema.Entity `json:"entities"`
}

// LookupResponse -
type LookupResponse struct {
	ActionNames []string `json:"action_names"`
}
