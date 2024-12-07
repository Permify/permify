package context

import (
	"slices"
	"testing"

	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

func TestQueryRelationships1(t *testing.T) {
	tuples := []*base.Tuple{
		{Entity: &base.Entity{Type: "type1", Id: "1"}, Relation: "relation1", Subject: &base.Subject{Type: "type2", Id: "2", Relation: "relation2"}},
		{Entity: &base.Entity{Type: "type3", Id: "3"}, Relation: "relation3", Subject: &base.Subject{Type: "type4", Id: "4", Relation: "relation4"}},
		{Entity: &base.Entity{Type: "type5", Id: "5"}, Relation: "relation5", Subject: &base.Subject{Type: "type6", Id: "6", Relation: "relation6"}},
	}

	contextualTuples := NewContextualTuples(tuples...)
	filter := &base.TupleFilter{Entity: &base.EntityFilter{Type: "type1", Ids: []string{"1"}}, Relation: "relation1", Subject: &base.SubjectFilter{Type: "type2", Ids: []string{"2"}, Relation: "relation2"}}

	iterator, err := contextualTuples.QueryRelationships(filter, database.NewCursorPagination())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !iterator.HasNext() {
		t.Errorf("Expected at least one tuple, got none")
	}

	filteredTuple := iterator.GetNext()
	if filteredTuple.Entity.Type != "type1" || filteredTuple.Entity.Id != "1" || filteredTuple.Relation != "relation1" || filteredTuple.Subject.Type != "type2" || filteredTuple.Subject.Id != "2" || filteredTuple.Subject.Relation != "relation2" {
		t.Errorf("Unexpected tuple: %+v", filteredTuple)
	}

	if iterator.HasNext() {
		t.Errorf("Expected exactly one tuple, got more")
	}
}

func TestQueryRelationships2(t *testing.T) {
	tuples := []*base.Tuple{
		{Entity: &base.Entity{Type: "type", Id: "1"}, Relation: "relation1", Subject: &base.Subject{Type: "type2", Id: "2", Relation: "relation2"}},
		{Entity: &base.Entity{Type: "type", Id: "3"}, Relation: "relation3", Subject: &base.Subject{Type: "type4", Id: "4", Relation: "relation4"}},
		{Entity: &base.Entity{Type: "type", Id: "5"}, Relation: "relation5", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
	}

	contextualTuples := NewContextualTuples(tuples...)
	filter := &base.TupleFilter{Entity: &base.EntityFilter{Type: "type", Ids: []string{"5"}}, Relation: "relation5"}

	iterator, err := contextualTuples.QueryRelationships(filter, database.NewCursorPagination())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !iterator.HasNext() {
		t.Errorf("Expected at least one tuple, got none")
	}

	filteredTuple := iterator.GetNext()
	if filteredTuple.Entity.Type != "type" || filteredTuple.Entity.Id != "5" || filteredTuple.Relation != "relation5" || filteredTuple.Subject.Type != "user" || filteredTuple.Subject.Id != "6" || filteredTuple.Subject.Relation != "" {
		t.Errorf("Unexpected tuple: %+v", filteredTuple)
	}

	if iterator.HasNext() {
		t.Errorf("Expected exactly one tuple, got more")
	}
}

func TestQueryRelationships3(t *testing.T) {
	tuples := []*base.Tuple{
		{Entity: &base.Entity{Type: "type1", Id: "1"}, Relation: "relation1", Subject: &base.Subject{Type: "type1", Id: "1", Relation: "relation1"}},
		{Entity: &base.Entity{Type: "type1", Id: "3"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type2", Id: "3"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type2", Id: "2"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type1", Id: "3"}, Relation: "relation1", Subject: &base.Subject{Type: "type2", Id: "4", Relation: "relation2"}},
		{Entity: &base.Entity{Type: "type3", Id: "1"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type3", Id: "3"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
	}

	contextualTuples := NewContextualTuples(tuples...)
	filter := &base.TupleFilter{Entity: &base.EntityFilter{Type: "type1", Ids: []string{"3"}}, Relation: "relation1"}

	iterator, err := contextualTuples.QueryRelationships(filter, database.NewCursorPagination())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	count := 0
	for iterator.HasNext() {
		filteredTuple := iterator.GetNext()

		count++

		if !slices.Contains([]string{
			"type1:3#relation1@user:6",
			"type1:3#relation1@type2:4#relation2",
		}, tuple.ToString(filteredTuple)) {
			t.Errorf("Unexpected tuple: %+v", filteredTuple)
		}
	}

	if count != 2 {
		t.Errorf("Unexpected count")
	}
}

func TestQueryRelationships4(t *testing.T) {
	tuples := []*base.Tuple{
		{Entity: &base.Entity{Type: "type1", Id: "1"}, Relation: "relation1", Subject: &base.Subject{Type: "type1", Id: "1", Relation: "relation1"}},
		{Entity: &base.Entity{Type: "type1", Id: "3"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type2", Id: "3"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type2", Id: "2"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type1", Id: "3"}, Relation: "relation1", Subject: &base.Subject{Type: "type2", Id: "4", Relation: "relation2"}},
		{Entity: &base.Entity{Type: "type3", Id: "1"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type3", Id: "3"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
	}

	contextualTuples := NewContextualTuples(tuples...)
	filter := &base.TupleFilter{Entity: &base.EntityFilter{Type: "type1", Ids: []string{"3"}}, Relation: "relation1", Subject: &base.SubjectFilter{Type: "type2", Ids: []string{}, Relation: ""}}

	iterator, err := contextualTuples.QueryRelationships(filter, database.NewCursorPagination())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	count := 0
	for iterator.HasNext() {
		filteredTuple := iterator.GetNext()

		count++

		if !slices.Contains([]string{
			"type1:3#relation1@type2:4#relation2",
		}, tuple.ToString(filteredTuple)) {
			t.Errorf("Unexpected tuple: %+v", filteredTuple)
		}
	}

	if count != 1 {
		t.Errorf("Unexpected count")
	}
}

func TestQueryRelationships5(t *testing.T) {
	tuples := []*base.Tuple{
		{Entity: &base.Entity{Type: "type1", Id: "1"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "1", Relation: ""}},
		{Entity: &base.Entity{Type: "type1", Id: "3"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type1", Id: "3"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "4", Relation: ""}},

		{Entity: &base.Entity{Type: "type2", Id: "2"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type2", Id: "3"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "5", Relation: ""}},
		{Entity: &base.Entity{Type: "type3", Id: "1"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type3", Id: "3"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
	}

	contextualTuples := NewContextualTuples(tuples...)
	filter := &base.TupleFilter{Entity: &base.EntityFilter{Type: "type1", Ids: []string{"1", "3"}}, Relation: "relation1", Subject: &base.SubjectFilter{Type: "user", Ids: []string{"1", "6", "4", "7"}, Relation: ""}}

	iterator, err := contextualTuples.QueryRelationships(filter, database.NewCursorPagination())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	count := 0
	for iterator.HasNext() {
		filteredTuple := iterator.GetNext()

		count++

		if !slices.Contains([]string{
			"type1:1#relation1@user:1",
			"type1:3#relation1@user:6",
			"type1:3#relation1@user:4",
		}, tuple.ToString(filteredTuple)) {
			t.Errorf("Unexpected tuple: %+v", filteredTuple)
		}
	}

	if count != 3 {
		t.Errorf("Unexpected count")
	}
}

func TestQueryRelationships6(t *testing.T) {
	tuples := []*base.Tuple{
		{Entity: &base.Entity{Type: "type1", Id: "2"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "1", Relation: ""}},
		{Entity: &base.Entity{Type: "type1", Id: "3"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type1", Id: "3"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "4", Relation: ""}},
		{Entity: &base.Entity{Type: "type2", Id: "12"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type2", Id: "32"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "5", Relation: ""}},
		{Entity: &base.Entity{Type: "type3", Id: "43"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type3", Id: "56"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
	}

	contextualTuples := NewContextualTuples(tuples...)
	filter := &base.TupleFilter{Entity: &base.EntityFilter{Type: "", Ids: []string{"1", "3", "12", "32", "43", "56"}}, Relation: "", Subject: &base.SubjectFilter{Type: "user", Ids: []string{"1", "6", "4", "7"}, Relation: ""}}

	iterator, err := contextualTuples.QueryRelationships(filter, database.NewCursorPagination(database.Cursor("Mw=="), database.Sort("entity_id")))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	count := 0
	for iterator.HasNext() {
		filteredTuple := iterator.GetNext()

		count++

		if !slices.Contains([]string{
			"type1:3#relation1@user:6",
			"type1:3#relation1@user:4",
			"type3:43#relation2@user:6",
			"type3:56#relation2@user:6",
		}, tuple.ToString(filteredTuple)) {
			t.Errorf("Unexpected tuple: %+v", filteredTuple)
		}
	}

	if count != 4 {
		t.Errorf("Unexpected count")
	}
}

func TestQueryRelationships7(t *testing.T) {
	tuples := []*base.Tuple{
		{Entity: &base.Entity{Type: "type1", Id: "2"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "1", Relation: ""}},
		{Entity: &base.Entity{Type: "type1", Id: "3"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type1", Id: "3"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "4", Relation: ""}},
		{Entity: &base.Entity{Type: "type2", Id: "12"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type2", Id: "32"}, Relation: "relation1", Subject: &base.Subject{Type: "user", Id: "5", Relation: ""}},
		{Entity: &base.Entity{Type: "type3", Id: "43"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
		{Entity: &base.Entity{Type: "type3", Id: "56"}, Relation: "relation2", Subject: &base.Subject{Type: "user", Id: "6", Relation: ""}},
	}

	contextualTuples := NewContextualTuples(tuples...)
	filter := &base.TupleFilter{Entity: &base.EntityFilter{Type: "", Ids: []string{"1", "3", "12", "32", "43", "56"}}, Relation: "", Subject: &base.SubjectFilter{Type: "user", Ids: []string{"1", "6", "4", "5", "7"}, Relation: ""}}

	iterator, err := contextualTuples.QueryRelationships(filter, database.NewCursorPagination(database.Cursor("NA=="), database.Sort("subject_id")))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	count := 0
	for iterator.HasNext() {
		filteredTuple := iterator.GetNext()

		count++

		if !slices.Contains([]string{
			"type1:3#relation1@user:4",
			"type2:32#relation1@user:5",
			"type1:3#relation1@user:6",
			"type2:12#relation2@user:6",
			"type3:43#relation2@user:6",
			"type3:56#relation2@user:6",
		}, tuple.ToString(filteredTuple)) {
			t.Errorf("Unexpected tuple: %+v", filteredTuple)
		}
	}

	if count != 6 {
		t.Errorf("Unexpected count")
	}
}
