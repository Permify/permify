package commands

import (
	"golang.org/x/net/context"

	"github.com/Permify/permify/pkg/errors"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

func getSubjects(ctx context.Context, command ICommand, ear *base.EntityAndRelation) (iterator tuple.ISubjectIterator, err errors.Error) {
	r := tuple.SplitRelation(ear.GetRelation())

	var tupleIterator tuple.ITupleIterator
	tupleIterator, err = command.GetRelationTupleRepository().QueryTuples(ctx, ear.GetEntity().GetType(), ear.GetEntity().GetId(), r[0])
	if err != nil {
		return nil, err
	}

	collection := tuple.NewSubjectCollection()
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
