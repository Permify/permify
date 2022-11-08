package commands

import (
	"golang.org/x/net/context"

	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

func getSubjects(ctx context.Context, command ICommand, ear *base.EntityAndRelation, token string) (iterator database.ISubjectIterator, err error) {
	r := tuple.SplitRelation(ear.GetRelation())
	var tupleCollection database.ITupleCollection
	tupleCollection, err = command.RelationshipReader().QueryRelationships(ctx, &base.TupleFilter{
		Entity: &base.EntityFilter{
			Type: ear.GetEntity().GetType(),
			Ids:  []string{ear.GetEntity().GetId()},
		},
		Relation: r[0],
	}, token)
	if err != nil {
		return nil, err
	}

	tupleIterator := tupleCollection.CreateTupleIterator()

	collection := database.NewSubjectCollection()
	for tupleIterator.HasNext() {
		tup := tupleIterator.GetNext()
		if !tuple.IsSubjectUser(tup.Subject) {
			subject := tup.Subject
			if tup.Subject.Relation == tuple.ELLIPSIS {
				subject.Relation = r[1]
			} else {
				subject.Relation = tup.Subject.Relation
			}
			collection.Add(subject)
		} else {
			collection.Add(tup.Subject)
		}
	}

	return collection.CreateSubjectIterator(), err
}
