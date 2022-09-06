package permission

import (
	"github.com/Permify/permify/internal/commands"
)

// CheckResponse -
type CheckResponse struct {
	Can            bool        `json:"can"`
	RemainingDepth int32       `json:"remaining_depth"`
	Decisions      interface{} `json:"decisions"`
}

// ExpandResponse -
type ExpandResponse struct {
	Tree commands.Node `json:"tree"`
}
