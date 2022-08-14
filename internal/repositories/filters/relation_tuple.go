package filters

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// RelationTupleFilter -
type RelationTupleFilter struct {
	Entity          string `json:"entity"`
	ID              string `json:"id"`
	Relation        string `json:"relation"`
	SubjectType     string `json:"subject_type"`
	SubjectID       string `json:"subject_id"`
	SubjectRelation string `json:"subject_relation"`
}

// Validate -
func (r RelationTupleFilter) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r,
		// object
		validation.Field(&r.Entity, validation.Required),
	)
	return
}
