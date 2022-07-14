package permission

import (
	validation "github.com/go-ozzo/ozzo-validation"
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
		User   string `json:"user" form:"user" xml:"user"`
		Action string `json:"action" form:"action" xml:"action"`
		Object string `json:"object" form:"object" xml:"object"`
		Depth  int    `json:"depth" form:"depth" xml:"depth"`
	}
}

// Validate -
func (r Check) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r.Body,
		validation.Field(&r.Body.User, validation.Required),
		validation.Field(&r.Body.Action, validation.Required),
		validation.Field(&r.Body.Object, validation.Required),
		validation.Field(&r.Body.Depth, validation.Min(3)),
	)
	return
}
