package repositories

import (
	`time`

	base `github.com/Permify/permify/pkg/pb/base/v1`
)

// RelationTuple -
type RelationTuple struct {
	Entity          string
	EntityID        string
	Relation        string
	SubjectEntity   string
	SubjectID       string
	SubjectRelation string
	CommitTime      time.Time
}

// ToTuple -
func (r RelationTuple) ToTuple() *base.Tuple {
	return &base.Tuple{
		Entity: &base.Entity{
			Type: r.Entity,
			Id:   r.EntityID,
		},
		Relation: r.Relation,
		Subject: &base.Subject{
			Type:     r.SubjectEntity,
			Id:       r.SubjectID,
			Relation: r.SubjectRelation,
		},
	}
}

// EntityConfig -
type EntityConfig struct {
	Entity           string
	SerializedConfig []byte
	Version          string
	CommitTime       time.Time
}

// Serialized -
func (e EntityConfig) Serialized() string {
	return string(e.SerializedConfig)
}
