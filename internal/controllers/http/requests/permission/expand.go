package permission

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/Permify/permify/internal/utils"
	"github.com/Permify/permify/pkg/tuple"
)

// ExpandRequest -
type ExpandRequest struct {
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
		Entity        tuple.Entity  `json:"entity" form:"entity" xml:"entity"`
		Action        string        `json:"action" form:"action" xml:"action"`
	}
}

// Validate -
func (r ExpandRequest) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r.Body,
		validation.Field(&r.Body.Entity, validation.Required),
		validation.Field(&r.Body.Action, validation.Required),
	)
	return
}
