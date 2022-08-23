package relationship

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	`github.com/Permify/permify/internal/utils`
	"github.com/Permify/permify/pkg/tuple"
)

// WriteRequest -
type WriteRequest struct {
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
		Relation      string        `json:"relation" form:"relation" xml:"relation"`
		Subject       tuple.Subject `json:"subject" form:"subject" xml:"subject"`
	}
}

// Validate -
func (r WriteRequest) Validate() (err error) {
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
