package context

import (
	"golang.org/x/exp/slices"

	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// ContextualAttributes - A collection of attributes with context.
type ContextualAttributes struct {
	Attributes []*base.Attribute
}

// NewContextualAttributes - Creates a new collection of attributes with context.
func NewContextualAttributes(attributes ...*base.Attribute) *ContextualAttributes {
	return &ContextualAttributes{
		Attributes: attributes,
	}
}

// QuerySingleAttribute filters the attributes based on the provided filter,
// and returns the first attribute from the filtered attributes, if any exist.
// If no attributes match the filter, it returns nil.
func (c *ContextualAttributes) QuerySingleAttribute(filter *base.AttributeFilter) (*base.Attribute, error) {
	filtered := c.filterAttributes(filter)
	if len(filtered) > 0 {
		return filtered[0], nil
	}
	return nil, nil
}

// QueryAttributes filters the attributes based on the provided filter,
// and returns an iterator to traverse through the filtered attributes.
func (c *ContextualAttributes) QueryAttributes(filter *base.AttributeFilter) (*database.AttributeIterator, error) {
	filtered := c.filterAttributes(filter)
	return database.NewAttributeIterator(filtered...), nil
}

// filterTuples applies the provided filter to c's Tuples and returns a slice of Tuples that match the filter.
func (c *ContextualAttributes) filterAttributes(filter *base.AttributeFilter) []*base.Attribute {
	var filtered []*base.Attribute // Initialize a slice to hold the filtered tuples

	// Iterate over the tuples
	for _, attribute := range c.Attributes {
		// If a tuple matches the Entity, Relation, and Subject filters, add it to the filtered slice
		if matchesEntityFilterForAttributes(attribute, filter.GetEntity()) &&
			matchesAttributeFilter(attribute, filter.GetAttributes()) {
			filtered = append(filtered, attribute)
		}
	}

	return filtered // Return the filtered tuples
}

// matchesEntityFilterForAttributes checks if an attribute matches an entity filter.
func matchesEntityFilterForAttributes(attribute *base.Attribute, filter *base.EntityFilter) bool {
	typeMatches := filter.GetType() == "" || attribute.GetEntity().GetType() == filter.GetType()
	idMatches := len(filter.GetIds()) == 0 || slices.Contains(filter.GetIds(), attribute.GetEntity().GetId())
	return typeMatches && idMatches
}

// matchesAttributeFilter checks if an attribute matches the provided filter.
func matchesAttributeFilter(attribute *base.Attribute, filter []string) bool {
	return len(filter) == 0 || slices.Contains(filter, attribute.GetAttribute())
}
