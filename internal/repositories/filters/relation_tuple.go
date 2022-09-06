package filters

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// RelationTupleFilter -
type RelationTupleFilter struct {
	Entity   EntityFilter  `json:"entity"`
	Relation string        `json:"relation"`
	Subject  SubjectFilter `json:"subject"`
}

// Validate -
func (r RelationTupleFilter) Validate() (err error) {
	err = validation.ValidateStruct(&r,
		validation.Field(&r.Entity),
	)
	return
}

// EntityFilter -
type EntityFilter struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// Validate -
func (r EntityFilter) Validate() (err error) {
	err = validation.ValidateStruct(&r,
		validation.Field(&r.Type, validation.Required),
	)
	return
}

// SubjectFilter -
type SubjectFilter struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	Relation string `json:"relation"`
}

// Validate -
func (r SubjectFilter) Validate() (err error) {
	return nil
}
