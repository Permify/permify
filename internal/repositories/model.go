package repositories

import (
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// RelationTuple -
type RelationTuple struct {
	EntityType      string
	EntityID        string
	Relation        string
	SubjectType     string
	SubjectID       string
	SubjectRelation string
}

// ToTuple -
func (r RelationTuple) ToTuple() *base.Tuple {
	return &base.Tuple{
		Entity: &base.Entity{
			Type: r.EntityType,
			Id:   r.EntityID,
		},
		Relation: r.Relation,
		Subject: &base.Subject{
			Type:     r.SubjectType,
			Id:       r.SubjectID,
			Relation: r.SubjectRelation,
		},
	}
}

// SchemaDefinition -
type SchemaDefinition struct {
	EntityType           string
	SerializedDefinition []byte
	Version              string
}

// Serialized -
func (e SchemaDefinition) Serialized() string {
	return string(e.SerializedDefinition)
}
