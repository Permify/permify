package permission

import (
	"github.com/Permify/permify/internal/commands"
)

// ExpandResponse -
type ExpandResponse struct {
	Tree commands.Node `json:"tree"`
}
