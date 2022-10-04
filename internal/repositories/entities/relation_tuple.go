package entities

import (
	"time"

	"github.com/Permify/permify/pkg/tuple"
)

// RelationTuple -
type RelationTuple struct {
	Entity          string    `json:"entity"`
	ObjectID        string    `json:"object_id"`
	Relation        string    `json:"relation"`
	UsersetEntity   string    `json:"userset_entity"`
	UsersetObjectID string    `json:"userset_object_id"`
	UsersetRelation string    `json:"userset_relation"`
	CommitTime      time.Time `json:"commit_time"`
}

// Table -
func (RelationTuple) Table() string {
	return "relation_tuple"
}

// ToTuple -
func (r RelationTuple) ToTuple() tuple.Tuple {
	return tuple.Tuple{
		Entity: tuple.Entity{
			Type: r.Entity,
			ID:   r.ObjectID,
		},
		Relation: tuple.Relation(r.Relation),
		Subject: tuple.Subject{
			Type:     r.UsersetEntity,
			ID:       r.UsersetObjectID,
			Relation: tuple.Relation(r.UsersetRelation),
		},
	}
}

// RelationTuples -
type RelationTuples []RelationTuple

// ToTuple -
func (r RelationTuples) ToTuple() (tuples []tuple.Tuple) {
	for _, a := range r {
		tuples = append(tuples, a.ToTuple())
	}
	return
}
