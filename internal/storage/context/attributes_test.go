package context

import (
	"testing"

	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

func TestQueryAttributes1(t *testing.T) {
	attributes := []*base.Attribute{
		{Entity: &base.Entity{Type: "type1", Id: "1"}, Attribute: "attribute1", Value: nil},
		{Entity: &base.Entity{Type: "type3", Id: "3"}, Attribute: "attribute2", Value: nil},
		{Entity: &base.Entity{Type: "type5", Id: "5"}, Attribute: "attribute3", Value: nil},
	}

	contextualAttributes := NewContextualAttributes(attributes...)
	filter := &base.AttributeFilter{Entity: &base.EntityFilter{Type: "type1", Ids: []string{"1"}}, Attributes: []string{"attribute1"}}

	iterator, err := contextualAttributes.QueryAttributes(filter, database.NewCursorPagination())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !iterator.HasNext() {
		t.Errorf("Expected at least one attribute, got none")
	}

	filteredAttribute := iterator.GetNext()
	if filteredAttribute.Entity.Type != "type1" || filteredAttribute.Entity.Id != "1" || filteredAttribute.Attribute != "attribute1" {
		t.Errorf("Unexpected attribute: %+v", filteredAttribute)
	}

	if iterator.HasNext() {
		t.Errorf("Expected exactly one attribute, got more")
	}
}

func TestQueryAttributes2(t *testing.T) {
	attributes := []*base.Attribute{
		{Entity: &base.Entity{Type: "type1", Id: "1"}, Attribute: "attribute1", Value: nil},
		{Entity: &base.Entity{Type: "type3", Id: "3"}, Attribute: "attribute2", Value: nil},
		{Entity: &base.Entity{Type: "type5", Id: "5"}, Attribute: "attribute3", Value: nil},
	}

	contextualAttributes := NewContextualAttributes(attributes...)
	filter := &base.AttributeFilter{Entity: &base.EntityFilter{Type: "type3", Ids: []string{"3"}}, Attributes: []string{"attribute1", "attribute2"}}

	iterator, err := contextualAttributes.QueryAttributes(filter, database.NewCursorPagination())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !iterator.HasNext() {
		t.Errorf("Expected at least one attributes, got none")
	}

	filteredTuple := iterator.GetNext()
	if filteredTuple.Entity.Type != "type3" || filteredTuple.Entity.Id != "3" || filteredTuple.Attribute != "attribute2" {
		t.Errorf("Unexpected attribute: %+v", filteredTuple)
	}

	if iterator.HasNext() {
		t.Errorf("Expected exactly one attribute, got more")
	}
}

func TestQueryAttributes3(t *testing.T) {
	attributes := []*base.Attribute{
		{Entity: &base.Entity{Type: "type1", Id: "1"}, Attribute: "attribute1", Value: nil},
		{Entity: &base.Entity{Type: "type3", Id: "3"}, Attribute: "attribute2", Value: nil},
		{Entity: &base.Entity{Type: "type5", Id: "5"}, Attribute: "attribute3", Value: nil},
	}

	contextualAttributes := NewContextualAttributes(attributes...)
	filter := &base.AttributeFilter{Entity: &base.EntityFilter{Type: "type3", Ids: []string{"3"}}, Attributes: []string{"attribute1", "attribute2"}}

	attr, err := contextualAttributes.QuerySingleAttribute(filter)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if attr.Entity.Type != "type3" || attr.Entity.Id != "3" || attr.Attribute != "attribute2" {
		t.Errorf("Unexpected attribute: %+v", attr)
	}
}
