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

func TestQueryAttributes4(t *testing.T) {
	attributes := []*base.Attribute{
		{Entity: &base.Entity{Type: "type1", Id: "1"}, Attribute: "attribute1", Value: nil},
		{Entity: &base.Entity{Type: "type3", Id: "3"}, Attribute: "attribute2", Value: nil},
		{Entity: &base.Entity{Type: "type5", Id: "5"}, Attribute: "attribute3", Value: nil},
		{Entity: &base.Entity{Type: "type5", Id: "4"}, Attribute: "attribute4", Value: nil},
		{Entity: &base.Entity{Type: "type5", Id: "12"}, Attribute: "attribute12", Value: nil},
		{Entity: &base.Entity{Type: "type5", Id: "22"}, Attribute: "attribute22", Value: nil},
	}

	contextualAttributes := NewContextualAttributes(attributes...)
	filter := &base.AttributeFilter{Entity: &base.EntityFilter{Type: "type5", Ids: []string{}}, Attributes: []string{}}

	iterator, err := contextualAttributes.QueryAttributes(filter, database.NewCursorPagination(database.Cursor("MjI="), database.Sort("entity_id")))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !iterator.HasNext() {
		t.Errorf("Expected at least one attribute, got none")
	}

	filteredAttribute := iterator.GetNext()
	if filteredAttribute.Entity.Type != "type5" || filteredAttribute.Entity.Id != "22" || filteredAttribute.Attribute != "attribute22" {
		t.Errorf("Unexpected attribute: %+v", filteredAttribute)
	}

	if !iterator.HasNext() {
		t.Errorf("Expected at least one attribute, got none")
	}

	filteredAttribute2 := iterator.GetNext()
	if filteredAttribute2.Entity.Type != "type5" || filteredAttribute2.Entity.Id != "4" || filteredAttribute2.Attribute != "attribute4" {
		t.Errorf("Unexpected attribute: %+v", filteredAttribute2)
	}

	if !iterator.HasNext() {
		t.Errorf("Expected at least one attribute, got none")
	}

	filteredAttribute3 := iterator.GetNext()
	if filteredAttribute3.Entity.Type != "type5" || filteredAttribute3.Entity.Id != "5" || filteredAttribute3.Attribute != "attribute3" {
		t.Errorf("Unexpected attribute: %+v", filteredAttribute3)
	}
}
