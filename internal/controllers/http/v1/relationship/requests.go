package relationship

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/Permify/permify/internal/controllers/utils"
	"github.com/Permify/permify/internal/repositories/filters"
	"github.com/Permify/permify/pkg/tuple"
)

// WriteRequest -
type WriteRequest struct {
	SchemaVersion utils.Version `json:"schema_version"`
	Entity        tuple.Entity  `json:"entity"`
	Relation      string        `json:"relation"`
	Subject       tuple.Subject `json:"subject"`
}

// Validate -
func (r WriteRequest) Validate() (err error) {
	err = validation.ValidateStruct(&r,
		validation.Field(&r.Entity, validation.Required),
		validation.Field(&r.Relation, validation.Required),
		validation.Field(&r.Subject, validation.Required),
	)
	return
}

// ReadRequest -
type ReadRequest struct {
	Filter filters.RelationTupleFilter `json:"filter"`
}

// Validate -
func (r ReadRequest) Validate() (err error) {
	err = validation.ValidateStruct(&r,
		validation.Field(&r.Filter, validation.Required),
	)
	return
}

// DeleteRequest -
type DeleteRequest struct {
	Entity   tuple.Entity  `json:"entity"`
	Relation string        `json:"relation"`
	Subject  tuple.Subject `json:"subject"`
}

// Validate -
func (r DeleteRequest) Validate() (err error) {
	err = validation.ValidateStruct(&r,
		validation.Field(&r.Entity, validation.Required),
		validation.Field(&r.Relation, validation.Required),
		validation.Field(&r.Subject, validation.Required),
	)
	return
}
