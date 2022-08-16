package responses

import (
	"github.com/Permify/permify/internal/commands"
)

// Expand -
type Expand struct {
	Tree commands.Node `json:"tree"`
}
