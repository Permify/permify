package schema

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/Permify/permify/internal/controllers/utils"
)

// ReadRequest -
type ReadRequest struct {
	SchemaVersion utils.Version `json:"schema_version"`
}

// Validate -
func (r ReadRequest) Validate() (err error) {
	return nil
}

// LookupRequest -
type LookupRequest struct {
	SchemaVersion utils.Version `json:"schema_version"`
	EntityType    string        `json:"entity_type"`
	RelationNames []string      `json:"relation_names"`
}

// Validate -
func (r LookupRequest) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r,
		validation.Field(&r.EntityType, validation.Required),
		validation.Field(&r.RelationNames, validation.Required),
	)
	return
}
