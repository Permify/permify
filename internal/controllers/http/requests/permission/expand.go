package permission

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/Permify/permify/pkg/tuple"
)

// Expand -
type Expand struct {
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
		Entity tuple.Entity `json:"entity" form:"entity" xml:"entity"`
		Action string       `json:"action" form:"action" xml:"action"`
		Depth  int          `json:"depth" form:"depth" xml:"depth"`
	}
}

// Validate -
func (r Expand) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r.Body,
		validation.Field(&r.Body.Entity, validation.Required),
		validation.Field(&r.Body.Action, validation.Required),
		validation.Field(&r.Body.Depth, validation.Min(3)),
	)
	return
}
