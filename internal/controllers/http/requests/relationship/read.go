package relationship

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/Permify/permify/internal/repositories/filters"
)

// ReadRequest -
type ReadRequest struct {
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
		Filter filters.RelationTupleFilter `json:"filter"`
	}
}

// Validate -
func (r ReadRequest) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r.Body,
		// filter
		validation.Field(&r.Body.Filter, validation.Required),
	)
	return
}
