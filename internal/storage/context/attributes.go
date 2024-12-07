package context

import (
	"slices"
	"sort"

	"github.com/Permify/permify/internal/storage/context/utils"
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
	filtered := c.filterAttributes(filter, "", "")
	if len(filtered) > 0 {
		return filtered[0], nil
	}
	return nil, nil
}

// QueryAttributes filters the attributes based on the provided filter,
// and returns an iterator to traverse through the filtered attributes.
func (c *ContextualAttributes) QueryAttributes(filter *base.AttributeFilter, pagination database.CursorPagination) (*database.AttributeIterator, error) {
	// Sort tuples based on the provided order field
	sort.Slice(c.Attributes, func(i, j int) bool {
		switch pagination.Sort() {
		case "entity_id":
			return c.Attributes[i].GetEntity().GetId() < c.Attributes[j].GetEntity().GetId()
		default:
			return false // If no valid order is provided, no sorting is applied
		}
	})

	cursor := ""
	if pagination.Cursor() != "" {
		t, err := utils.EncodedContinuousToken{Value: pagination.Cursor()}.Decode()
		if err != nil {
			return nil, err
		}
		cursor = t.(utils.ContinuousToken).Value
	}

	filtered := c.filterAttributes(filter, cursor, pagination.Sort())
	return database.NewAttributeIterator(filtered...), nil
}

// filterTuples applies the provided filter to c's Tuples and returns a slice of Tuples that match the filter.
func (c *ContextualAttributes) filterAttributes(filter *base.AttributeFilter, cursor, order string) []*base.Attribute {
	var filtered []*base.Attribute // Initialize a slice to hold the filtered tuples

	// Iterate over the tuples
	for _, attribute := range c.Attributes {

		// Skip tuples that come before the cursor based on the specified order field
		if cursor != "" && !isAttributeAfterCursor(attribute, cursor, order) {
			continue
		}

		// If a tuple matches the Entity, Relation, and Subject filters, add it to the filtered slice
		if matchesEntityFilterForAttributes(attribute, filter.GetEntity()) &&
			matchesAttributeFilter(attribute, filter.GetAttributes()) {
			filtered = append(filtered, attribute)
		}
	}

	return filtered // Return the filtered tuples
}

// isAfterCursor checks if the tuple's ID (based on the order field) comes after the cursor.
func isAttributeAfterCursor(attr *base.Attribute, cursor, order string) bool {
	switch order {
	case "entity_id":
		return attr.GetEntity().GetId() >= cursor
	default:
		// If the order field is not recognized, default to not skipping any tuples
		return true
	}
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
