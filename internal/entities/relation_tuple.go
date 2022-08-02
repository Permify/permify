package entities

import (
	"time"

	"github.com/Permify/permify/pkg/tuple"
)

// RelationTuple -
type RelationTuple struct {
	Entity          string    `json:"entity" bson:"entity"`
	ObjectID        string    `json:"object_id" bson:"object_id"`
	Relation        string    `json:"relation" bson:"relation"`
	UsersetEntity   string    `json:"userset_entity" bson:"userset_entity"`
	UsersetObjectID string    `json:"userset_object_id" bson:"userset_object_id"`
	UsersetRelation string    `json:"userset_relation" bson:"userset_relation"`
	Type            string    `json:"type" bson:"type"`
	CommitTime      time.Time `json:"commit_time" bson:"commit_time"`
}

// Table -
func (RelationTuple) Table() string {
	return "relation_tuple"
}

// Collection -
func (RelationTuple) Collection() string {
	return "relation_tuple"
}

// String -
func (r RelationTuple) String() string {
	tup := tuple.Tuple{
		Object: tuple.Object{
			Namespace: r.Entity,
			ID:        r.ObjectID,
		},
		Relation: r.Relation,
		User: tuple.User{
			ID: r.UsersetObjectID,
			UserSet: tuple.UserSet{
				Object: tuple.Object{
					Namespace: r.UsersetEntity,
					ID:        r.UsersetObjectID,
				},
				Relation: tuple.Relation(r.UsersetRelation),
			},
		},
	}
	return tup.String()
}
