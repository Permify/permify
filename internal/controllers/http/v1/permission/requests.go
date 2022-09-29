package permission

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/Permify/permify/internal/controllers/utils"
	"github.com/Permify/permify/pkg/tuple"
)

// CheckRequest -
type CheckRequest struct {
	SchemaVersion utils.Version `json:"schema_version"`
	Entity        tuple.Entity  `json:"entity"`
	Action        string        `json:"action"`
	Subject       tuple.Subject `json:"subject"`
	Depth         int32         `json:"depth"`
}

// Validate -
func (r CheckRequest) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r,
		validation.Field(&r.Entity, validation.Required),
		validation.Field(&r.Action, validation.Required),
		validation.Field(&r.Subject, validation.Required),
		validation.Field(&r.Depth, validation.Min(3)),
	)
	return
}

// ExpandRequest -
type ExpandRequest struct {
	SchemaVersion utils.Version `json:"schema_version"`
	Entity        tuple.Entity  `json:"entity"`
	Action        string        `json:"action"`
}

// Validate -
func (r ExpandRequest) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r,
		validation.Field(&r.Entity, validation.Required),
		validation.Field(&r.Action, validation.Required),
	)
	return
}

// LookupQueryRequest -
type LookupQueryRequest struct {
	SchemaVersion utils.Version `json:"schema_version"`
	EntityType    string        `json:"entity_type"`
	Action        string        `json:"action"`
	Subject       tuple.Subject `json:"subject"`
}

// Validate -
func (r LookupQueryRequest) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r,
		validation.Field(&r.EntityType, validation.Required),
		validation.Field(&r.Action, validation.Required),
		validation.Field(&r.Subject, validation.Required),
	)
	return
}
