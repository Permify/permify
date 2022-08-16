package permission

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/Permify/permify/pkg/tuple"
)

// Check -
type Check struct {
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
		Entity  tuple.Entity  `json:"entity" form:"entity" xml:"entity"`
		Action  string        `json:"action" form:"action" xml:"action"`
		Subject tuple.Subject `json:"subject" form:"subject" xml:"subject"`
		Depth   int32         `json:"depth" form:"depth" xml:"depth"`
	}
}

// Validate -
func (r Check) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r.Body,
		validation.Field(&r.Body.Entity, validation.Required),
		validation.Field(&r.Body.Action, validation.Required),
		validation.Field(&r.Body.Subject, validation.Required),
		validation.Field(&r.Body.Depth, validation.Min(3)),
	)
	return
}
