package relationship

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Write -
type Write struct {
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
		Entity          string `json:"entity" form:"entity" xml:"entity"`
		ObjectID        string `json:"object_id" form:"object_id" xml:"object_id"`
		Relation        string `json:"relation" form:"relation" xml:"relation"`
		UsersetEntity   string `json:"userset_entity" form:"userset_entity" xml:"userset_entity"`
		UsersetObjectID string `json:"userset_object_id" form:"userset_object_id" xml:"userset_object_id"`
		UsersetRelation string `json:"userset_relation" form:"userset_relation" xml:"userset_relation"`
	}
}

// Validate -
func (r Write) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r.Body,
		validation.Field(&r.Body.Entity, validation.Required),
		validation.Field(&r.Body.ObjectID, validation.Required),
		validation.Field(&r.Body.Relation, validation.Required),
		validation.Field(&r.Body.UsersetEntity, validation.When(r.Body.UsersetRelation != "", validation.Required).Else(validation.Empty)),
		validation.Field(&r.Body.UsersetObjectID, validation.Required),
		validation.Field(&r.Body.UsersetRelation, validation.When(r.Body.UsersetEntity != "", validation.Required).Else(validation.Empty)),
	)
	return
}
