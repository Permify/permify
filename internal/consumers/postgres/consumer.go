package postgres

import (
	"context"
	`encoding/json`
	`github.com/lib/pq`

	`github.com/Permify/permify/internal/consumers`
	e `github.com/Permify/permify/internal/entities`
)

// PQConsumer -
type PQConsumer struct {
	Consumer consumers.Consumer
}

// New -
func New(consumer consumers.Consumer) PQConsumer {
	return PQConsumer{
		Consumer: consumer,
	}
}

// Consume -
func (c *PQConsumer) Consume(ctx context.Context, event chan *pq.Notification) {
	go func() {
		for {
			select {
			case n := <-event:

				if n == nil {
					continue
				}

				var err error
				var notification consumers.Notification
				err = json.Unmarshal([]byte(n.Extra), &notification)
				if err != nil {
					continue
				}

				// Parse
				writes, deletes := c.Consumer.Parse(notification)

				var deleteRelationTuples []e.RelationTuple

				for _, t := range deletes {
					relationTuple := e.RelationTuple{
						Entity:   t.Object.Namespace,
						ObjectID: t.Object.ID,
						Relation: t.Relation,
					}

					if t.User.ID != "" {
						relationTuple.UsersetObjectID = t.User.ID
					} else {
						relationTuple.UsersetEntity = t.User.UserSet.Object.Namespace
						relationTuple.UsersetObjectID = t.User.UserSet.Object.ID
						relationTuple.UsersetRelation = t.User.UserSet.Relation
					}

					deleteRelationTuples = append(deleteRelationTuples, relationTuple)
				}

				err = c.Consumer.GetService().DeleteRelationship(ctx, deleteRelationTuples)
				if err != nil {
					continue
				}

				var writeRelationTuples []e.RelationTuple

				for _, w := range writes {
					relationTuple := e.RelationTuple{
						Entity:   w.Object.Namespace,
						ObjectID: w.Object.ID,
						Relation: w.Relation,
						Type:     "auto",
					}

					if w.User.ID != "" {
						relationTuple.UsersetObjectID = w.User.ID
					} else {
						relationTuple.UsersetEntity = w.User.UserSet.Object.Namespace
						relationTuple.UsersetObjectID = w.User.UserSet.Object.ID
						relationTuple.UsersetRelation = w.User.UserSet.Relation
					}
					writeRelationTuples = append(writeRelationTuples, relationTuple)
				}

				err = c.Consumer.GetService().WriteRelationship(ctx, writeRelationTuples)
				if err != nil {
					continue
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}
