package postgres

import (
	"context"

	"github.com/Permify/permify/internal/consumers"
	e "github.com/Permify/permify/internal/entities"
	publisher "github.com/Permify/permify/pkg/publisher/postgres"
	"github.com/Permify/permify/pkg/tuple"
)

// PQConsumer -
type PQConsumer struct {
	parser consumers.Parser
}

// New -
func New(parser consumers.Parser) PQConsumer {
	return PQConsumer{
		parser: parser,
	}
}

// Consume -
func (c *PQConsumer) Consume(ctx context.Context, event chan *publisher.Notification) {
	for {
		select {
		case n := <-event:

			if n == nil {
				continue
			}

			var err error

			var w []tuple.Tuple
			var d []tuple.Tuple

			w, d, err = c.parser.Parse(*n)

			if err != nil {
				continue
			}

			var deleteRelationTuples []e.RelationTuple

			for _, t := range d {
				relationTuple := e.RelationTuple{
					Entity:   t.Entity.Type,
					ObjectID: t.Entity.ID,
					Relation: t.Relation,
				}

				if t.Subject.IsUser() {
					relationTuple.UsersetEntity = tuple.USER
					relationTuple.UsersetObjectID = t.Subject.ID
				} else {
					relationTuple.UsersetEntity = t.Subject.Type
					relationTuple.UsersetObjectID = t.Subject.ID
					relationTuple.UsersetRelation = t.Subject.Relation.String()
				}

				deleteRelationTuples = append(deleteRelationTuples, relationTuple)
			}

			err = c.parser.GetService().DeleteRelationship(ctx, deleteRelationTuples)
			if err != nil {
				continue
			}

			var writeRelationTuples []e.RelationTuple

			for _, t := range w {
				relationTuple := e.RelationTuple{
					Entity:   t.Entity.Type,
					ObjectID: t.Entity.ID,
					Relation: t.Relation,
					Type:     "auto",
				}

				if t.Subject.IsUser() {
					relationTuple.UsersetEntity = tuple.USER
					relationTuple.UsersetObjectID = t.Subject.ID
				} else {
					relationTuple.UsersetEntity = t.Subject.Type
					relationTuple.UsersetObjectID = t.Subject.ID
					relationTuple.UsersetRelation = t.Subject.Relation.String()
				}
				writeRelationTuples = append(writeRelationTuples, relationTuple)
			}

			err = c.parser.GetService().WriteRelationship(ctx, writeRelationTuples)
			if err != nil {
				continue
			}

		case <-ctx.Done():
			return
		}
	}
}
