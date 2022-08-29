package schema

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/Permify/permify/internal/utils"
)

// LookupRequest -
type LookupRequest struct {
	/**
	 * PathParams
	 */
	PathParams struct{}

	/**
	 * QueryParams
	 */
	QueryParams struct{}

	/**
	 * Body
	 */
	Body struct {
		SchemaVersion utils.Version `json:"schema_version" form:"schema_version" xml:"schema_version"`
		EntityType    string        `json:"entity_type" form:"entity_type" xml:"entity_type"`
		RelationNames []string      `json:"relation_names" form:"relation_names" xml:"relation_names"`
	}
}

// Validate -
func (r LookupRequest) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r.Body,
		validation.Field(&r.Body.EntityType, validation.Required),
		validation.Field(&r.Body.RelationNames, validation.Required),
	)
	return
}
