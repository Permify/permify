package relationship

import (
	validation "github.com/go-ozzo/ozzo-validation"
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
		Namespace        string `json:"namespace" form:"namespace" xml:"namespace"`
		ObjectID         string `json:"object_id" form:"object_id" xml:"object_id"`
		Relation         string `json:"relation" form:"relation" xml:"relation"`
		UsersetNamespace string `json:"userset_namespace" form:"userset_namespace" xml:"userset_namespace"`
		UsersetObjectID  string `json:"userset_object_id" form:"userset_object_id" xml:"userset_object_id"`
		UsersetRelation  string `json:"userset_relation" form:"userset_relation" xml:"userset_relation"`
	}
}

func (r Write) Validate() (err error) {
	// Validate Body
	err = validation.ValidateStruct(&r.Body,
		validation.Field(&r.Body.Namespace, validation.Required),
		validation.Field(&r.Body.ObjectID, validation.Required),
		validation.Field(&r.Body.Relation, validation.Required),
		validation.Field(&r.Body.UsersetNamespace),
		validation.Field(&r.Body.UsersetObjectID, validation.Required),
		validation.Field(&r.Body.UsersetRelation),
	)
	return
}
