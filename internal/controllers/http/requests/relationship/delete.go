package relationship

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/Permify/permify/pkg/tuple"
)

// DeleteRequest -
type DeleteRequest struct {
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
		Entity   tuple.Entity  `json:"entity"`
		Relation string        `json:"relation"`
		Subject  tuple.Subject `json:"subject"`
	}
}

// Validate -
func (r DeleteRequest) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r.Body,
		// object
		validation.Field(&r.Body.Entity, validation.Required),

		// relation
		validation.Field(&r.Body.Relation, validation.Required),

		// subject
		validation.Field(&r.Body.Subject, validation.Required),
	)
	return
}
