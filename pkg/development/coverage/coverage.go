package coverage

import (
	"fmt"
	"slices"

	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/development/file"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// SchemaCoverageInfo - Schema coverage info
type SchemaCoverageInfo struct {
	EntityCoverageInfo         []EntityCoverageInfo
	TotalRelationshipsCoverage int
	TotalAttributesCoverage    int
	TotalAssertionsCoverage    int
}

// EntityCoverageInfo - Entity coverage info
type EntityCoverageInfo struct {
	EntityName string

	UncoveredRelationships       []string
	CoverageRelationshipsPercent int

	UncoveredAttributes       []string
	CoverageAttributesPercent int

	UncoveredAssertions       map[string][]string
	CoverageAssertionsPercent map[string]int
}

// SchemaCoverage
//
// schema:
//
//	entity user {}
//
//	entity organization {
//	    // organizational roles
//	    relation admin @user
//	    relation member @user
//	}
//
//	entity repository {
//	    // represents repositories parent organization
//	    relation parent @organization
//
//	    // represents owner of this repository
//	    relation owner  @user @organization#admin
//
//	    // permissions
//	    permission edit   = parent.admin or owner
//	    permission delete = owner
//	}
//
// - relationships coverage
//
// organization#admin@user
// organization#member@user
// repository#parent@organization
// repository#owner@user
// repository#owner@organization#admin
//
// - assertions coverage
//
// repository#edit
// repository#delete
type SchemaCoverage struct {
	EntityName    string
	Relationships []string
	Attributes    []string
	Assertions    []string
}

func Run(shape file.Shape) SchemaCoverageInfo {
	p, err := parser.NewParser(shape.Schema).Parse()
	if err != nil {
		return SchemaCoverageInfo{}
	}

	definitions, _, err := compiler.NewCompiler(true, p).Compile()
	if err != nil {
		return SchemaCoverageInfo{}
	}

	schemaCoverageInfo := SchemaCoverageInfo{}

	refs := make([]SchemaCoverage, len(definitions))
	for i, en := range definitions {
		refs[i] = references(en)
	}

	// Iterate through the schema coverage references
	for _, ref := range refs {
		// Initialize EntityCoverageInfo for the current entity
		entityCoverageInfo := EntityCoverageInfo{
			EntityName:                   ref.EntityName,
			UncoveredRelationships:       []string{},
			UncoveredAttributes:          []string{},
			CoverageAssertionsPercent:    map[string]int{},
			UncoveredAssertions:          map[string][]string{},
			CoverageRelationshipsPercent: 0,
			CoverageAttributesPercent:    0,
		}

		// Calculate relationships coverage
		er := relationships(ref.EntityName, shape.Relationships)

		for _, relationship := range ref.Relationships {
			if !slices.Contains(er, relationship) {
				entityCoverageInfo.UncoveredRelationships = append(entityCoverageInfo.UncoveredRelationships, relationship)
			}
		}

		entityCoverageInfo.CoverageRelationshipsPercent = calculateCoveragePercent(
			ref.Relationships,
			entityCoverageInfo.UncoveredRelationships,
		)

		// Calculate attributes coverage
		at := attributes(ref.EntityName, shape.Attributes)

		for _, attr := range ref.Attributes {
			if !slices.Contains(at, attr) {
				entityCoverageInfo.UncoveredAttributes = append(entityCoverageInfo.UncoveredAttributes, attr)
			}
		}

		entityCoverageInfo.CoverageAttributesPercent = calculateCoveragePercent(
			ref.Attributes,
			entityCoverageInfo.UncoveredAttributes,
		)

		// Calculate assertions coverage for each scenario
		for _, s := range shape.Scenarios {
			ca := assertions(ref.EntityName, s.Checks, s.EntityFilters)

			for _, assertion := range ref.Assertions {
				if !slices.Contains(ca, assertion) {
					entityCoverageInfo.UncoveredAssertions[s.Name] = append(entityCoverageInfo.UncoveredAssertions[s.Name], assertion)
				}
			}

			entityCoverageInfo.CoverageAssertionsPercent[s.Name] = calculateCoveragePercent(
				ref.Assertions,
				entityCoverageInfo.UncoveredAssertions[s.Name],
			)
		}

		schemaCoverageInfo.EntityCoverageInfo = append(schemaCoverageInfo.EntityCoverageInfo, entityCoverageInfo)
	}

	// Calculate and assign the total relationships and assertions coverage to the schemaCoverageInfo
	relationshipsCoverage, attributesCoverage, assertionsCoverage := calculateTotalCoverage(schemaCoverageInfo.EntityCoverageInfo)
	schemaCoverageInfo.TotalRelationshipsCoverage = relationshipsCoverage
	schemaCoverageInfo.TotalAttributesCoverage = attributesCoverage
	schemaCoverageInfo.TotalAssertionsCoverage = assertionsCoverage

	return schemaCoverageInfo
}

// calculateCoveragePercent - Calculate coverage percentage based on total and uncovered elements
func calculateCoveragePercent(totalElements, uncoveredElements []string) int {
	coveragePercent := 100
	totalCount := len(totalElements)

	if totalCount != 0 {
		coveredCount := totalCount - len(uncoveredElements)
		coveragePercent = (coveredCount * 100) / totalCount
	}

	return coveragePercent
}

// calculateTotalCoverage - Calculate total relationships and assertions coverage
func calculateTotalCoverage(entities []EntityCoverageInfo) (int, int, int) {
	totalRelationships := 0
	totalCoveredRelationships := 0
	totalAttributes := 0
	totalCoveredAttributes := 0
	totalAssertions := 0
	totalCoveredAssertions := 0

	// Iterate over each entity in the list
	for _, entity := range entities {
		totalRelationships++
		totalCoveredRelationships += entity.CoverageRelationshipsPercent

		totalAttributes++
		totalCoveredAttributes += entity.CoverageAttributesPercent

		for _, assertionsPercent := range entity.CoverageAssertionsPercent {
			totalAssertions++
			totalCoveredAssertions += assertionsPercent
		}
	}

	// Calculate the coverage percentages
	totalRelationshipsCoverage := totalCoveredRelationships / totalRelationships
	totalAttributesCoverage := totalCoveredAttributes / totalAttributes
	totalAssertionsCoverage := totalCoveredAssertions / totalAssertions

	// Return the coverage percentages
	return totalRelationshipsCoverage, totalAttributesCoverage, totalAssertionsCoverage
}

// References - Get references for a given entity
func references(entity *base.EntityDefinition) (coverage SchemaCoverage) {
	// Set the entity name in the coverage struct
	coverage.EntityName = entity.GetName()
	// Iterate over all relations in the entity
	for _, relation := range entity.GetRelations() {
		// Iterate over all references within each relation
		for _, reference := range relation.GetRelationReferences() {
			if reference.GetRelation() != "" {
				// Format and append the relationship to the coverage struct
				formattedRelationship := fmt.Sprintf("%s#%s@%s#%s", entity.GetName(), relation.GetName(), reference.GetType(), reference.GetRelation())
				coverage.Relationships = append(coverage.Relationships, formattedRelationship)
			} else {
				formattedRelationship := fmt.Sprintf("%s#%s@%s", entity.GetName(), relation.GetName(), reference.GetType())
				coverage.Relationships = append(coverage.Relationships, formattedRelationship)
			}
		}
	}
	// Iterate over all attributes in the entity
	for _, attr := range entity.GetAttributes() {
		// Format and append the attribute to the coverage struct
		formattedAttribute := fmt.Sprintf("%s#%s", entity.GetName(), attr.GetName())
		coverage.Attributes = append(coverage.Attributes, formattedAttribute)
	}
	// Iterate over all permissions in the entity
	for _, permission := range entity.GetPermissions() {
		// Format and append the permission to the coverage struct
		formattedPermission := fmt.Sprintf("%s#%s", entity.GetName(), permission.GetName())
		coverage.Assertions = append(coverage.Assertions, formattedPermission)
	}
	// Return the coverage struct
	return
}

// relationships - Get relationships for a given entity
func relationships(en string, relationships []string) []string {
	var rels []string
	for _, relationship := range relationships {
		tup, err := tuple.Tuple(relationship)
		if err != nil {
			return []string{}
		}
		if tup.GetEntity().GetType() != en {
			continue
		}
		// Check if the reference has a relation name
		if tup.GetSubject().GetRelation() != "" {
			// Format and append the relationship to the coverage struct
			rels = append(rels, fmt.Sprintf("%s#%s@%s#%s", tup.GetEntity().GetType(), tup.GetRelation(), tup.GetSubject().GetType(), tup.GetSubject().GetRelation()))
		} else {
			rels = append(rels, fmt.Sprintf("%s#%s@%s", tup.GetEntity().GetType(), tup.GetRelation(), tup.GetSubject().GetType()))
		}
		// Format ad append the relationship without the relation name to the coverage struct
	}
	return rels
}

// attributes - Get attributes for a given entity
func attributes(en string, attributes []string) []string {
	attrs := make([]string, len(attributes))
	for i, attr := range attributes {
		a, err := attribute.Attribute(attr)
		if err != nil {
			return []string{}
		}
		if a.GetEntity().GetType() != en {
			continue
		}
		attrs[i] = fmt.Sprintf("%s#%s", a.GetEntity().GetType(), a.GetAttribute())
	}
	return attrs
}

// assertions - Get assertions for a given entity
func assertions(en string, checks []file.Check, filters []file.EntityFilter) []string {
	// Initialize an empty slice to store the resulting assertions
	var asrts []string

	// Iterate over each check in the checks slice
	for _, assertion := range checks {
		// Get the corresponding entity object for the current assertion
		ca, err := tuple.E(assertion.Entity)
		if err != nil {
			// If there's an error, return an empty slice
			return []string{}
		}

		// If the current entity type doesn't match the given entity type, continue to the next check
		if ca.GetType() != en {
			continue
		}

		// Iterate over the keys (permissions) in the Assertions map
		for permission := range assertion.Assertions {
			// Append the formatted permission string to the asrts slice
			asrts = append(asrts, fmt.Sprintf("%s#%s", ca.GetType(), permission))
		}
	}

	// Iterate over each entity filter in the filters slice
	for _, assertion := range filters {
		// If the current entity type doesn't match the given entity type, continue to the next filter
		if assertion.EntityType != en {
			continue
		}

		// Iterate over the keys (permissions) in the Assertions map
		for permission := range assertion.Assertions {
			// Append the formatted permission string to the asrts slice
			asrts = append(asrts, fmt.Sprintf("%s#%s", assertion.EntityType, permission))
		}
	}

	// Return the asrts slice containing the collected assertions
	return asrts
}
