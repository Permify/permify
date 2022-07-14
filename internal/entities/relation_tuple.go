package entities

import (
	"time"
)

// RelationTuple -
type RelationTuple struct {
	ID              int       `json:"id"`
	Entity          string    `json:"entity"`
	ObjectID        string    `json:"object_id"`
	Relation        string    `json:"relation"`
	UsersetEntity   string    `json:"userset_entity"`
	UsersetObjectID string    `json:"userset_object_id"`
	UsersetRelation string    `json:"userset_relation"`
	Type            string    `json:"type"`
	CommitTime      time.Time `json:"commit_time"`
}

// Table -
func (RelationTuple) Table() string {
	return "relation_tuple"
}
