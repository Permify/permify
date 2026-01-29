package coverage

import (
	"fmt"
	"slices"

	"github.com/Permify/permify/internal/coverage"
	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/development/file"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// SchemaCoverageInfo aliases internal coverage info
type SchemaCoverageInfo = coverage.SchemaCoverageInfo

// EntityCoverageInfo aliases internal entity coverage info
type EntityCoverageInfo = coverage.EntityCoverageInfo

// LogicNodeCoverage aliases internal logic node coverage info
type LogicNodeCoverage = coverage.LogicNodeCoverage

// SchemaCoverage represents the expected coverage for a schema entity
//
// Example schema:
//
//	entity user {}
//
//	entity organization {
//	    relation admin @user
//	    relation member @user
//	}
//
//	entity repository {
//	    relation parent @organization
//	    relation owner  @user @organization#admin
//	    permission edit   = parent.admin or owner
//	    permission delete = owner
//	}
//
// Expected relationships coverage:
//   - organization#admin@user
//   - organization#member@user
//   - repository#parent@organization
//   - repository#owner@user
//   - repository#owner@organization#admin
//
// Expected assertions coverage:
//   - repository#edit
//   - repository#delete
type SchemaCoverage struct {
	EntityName    string
	Relationships []string
	Attributes    []string
	Assertions    []string
}

// Run analyzes the coverage of relationships, attributes, and assertions
// for a given schema shape and returns the coverage information
func Run(shape file.Shape) SchemaCoverageInfo {
	p, err := parser.NewParser(shape.Schema).Parse()
	if err != nil {
		return SchemaCoverageInfo{}
	}

	definitions, _, err := compiler.NewCompiler(true, p).Compile()
	if err != nil {
		return SchemaCoverageInfo{}
	}

	registry := coverage.NewRegistry()
	coverage.Discover(p, registry)

	refs := extractSchemaReferences(definitions)
	entityCoverageInfos := calculateEntityCoverages(refs, shape)

	// Logic coverage is handled during scenario execution in calculateEntityCoverages if we wanted to be fully integration-style,
	// but since the current coverage tool is static, we'll mark logic nodes as uncovered based on registry report.
	// For now, let's just populate the logic coverage in calculateEntityCoverage.

	return buildSchemaCoverageInfo(entityCoverageInfos)
}

// parseAndCompileSchema parses and compiles the schema into entity definitions
func parseAndCompileSchema(schema string) ([]*base.EntityDefinition, error) {
	p, err := parser.NewParser(schema).Parse()
	if err != nil {
		return nil, err
	}

	definitions, _, err := compiler.NewCompiler(true, p).Compile()
	if err != nil {
		return nil, err
	}

	return definitions, nil
}

// extractSchemaReferences extracts all coverage references from entity definitions
func extractSchemaReferences(definitions []*base.EntityDefinition) []SchemaCoverage {
	refs := make([]SchemaCoverage, len(definitions))
	for idx, entityDef := range definitions {
		refs[idx] = extractEntityReferences(entityDef)
	}
	return refs
}

// extractEntityReferences extracts relationships, attributes, and assertions from an entity definition
func extractEntityReferences(entity *base.EntityDefinition) SchemaCoverage {
	coverage := SchemaCoverage{
		EntityName:    entity.GetName(),
		Relationships: extractRelationships(entity),
		Attributes:    extractAttributes(entity),
		Assertions:    extractAssertions(entity),
	}
	return coverage
}

// extractRelationships extracts all relationship references from an entity
func extractRelationships(entity *base.EntityDefinition) []string {
	relationships := []string{}

	for _, relation := range entity.GetRelations() {
		for _, reference := range relation.GetRelationReferences() {
			formatted := formatRelationship(
				entity.GetName(),
				relation.GetName(),
				reference.GetType(),
				reference.GetRelation(),
			)
			relationships = append(relationships, formatted)
		}
	}

	return relationships
}

// extractAttributes extracts all attribute references from an entity
func extractAttributes(entity *base.EntityDefinition) []string {
	attributes := []string{}

	for _, attr := range entity.GetAttributes() {
		formatted := formatAttribute(entity.GetName(), attr.GetName())
		attributes = append(attributes, formatted)
	}

	return attributes
}

// extractAssertions extracts all permission/assertion references from an entity
func extractAssertions(entity *base.EntityDefinition) []string {
	assertions := []string{}

	for _, permission := range entity.GetPermissions() {
		formatted := formatAssertion(entity.GetName(), permission.GetName())
		assertions = append(assertions, formatted)
	}

	return assertions
}

// calculateEntityCoverages calculates coverage for all entities
func calculateEntityCoverages(refs []SchemaCoverage, shape file.Shape) []EntityCoverageInfo {
	entityCoverageInfos := []EntityCoverageInfo{}

	for _, ref := range refs {
		entityCoverageInfo := calculateEntityCoverage(ref, shape)
		entityCoverageInfos = append(entityCoverageInfos, entityCoverageInfo)
	}

	return entityCoverageInfos
}

// calculateEntityCoverage calculates coverage for a single entity
func calculateEntityCoverage(ref SchemaCoverage, shape file.Shape) EntityCoverageInfo {
	entityCoverageInfo := newEntityCoverageInfo(ref.EntityName)

	// Calculate relationships coverage
	entityCoverageInfo.UncoveredRelationships = findUncoveredRelationships(
		ref.EntityName,
		ref.Relationships,
		shape.Relationships,
	)
	entityCoverageInfo.CoverageRelationshipsPercent = calculateCoveragePercent(
		ref.Relationships,
		entityCoverageInfo.UncoveredRelationships,
	)

	// Calculate attributes coverage
	entityCoverageInfo.UncoveredAttributes = findUncoveredAttributes(
		ref.EntityName,
		ref.Attributes,
		shape.Attributes,
	)
	entityCoverageInfo.CoverageAttributesPercent = calculateCoveragePercent(
		ref.Attributes,
		entityCoverageInfo.UncoveredAttributes,
	)

	// Calculate assertions coverage for each scenario
	for _, scenario := range shape.Scenarios {
		uncovered := findUncoveredAssertions(
			ref.EntityName,
			ref.Assertions,
			scenario.Checks,
			scenario.EntityFilters,
		)
		// Only add to UncoveredAssertions if there are uncovered assertions
		if len(uncovered) > 0 {
			entityCoverageInfo.UncoveredAssertions[scenario.Name] = uncovered
		}
		entityCoverageInfo.CoverageAssertionsPercent[scenario.Name] = calculateCoveragePercent(
			ref.Assertions,
			uncovered,
		)
	}

	return entityCoverageInfo
}

// newEntityCoverageInfo creates a new EntityCoverageInfo with initialized fields
func newEntityCoverageInfo(entityName string) EntityCoverageInfo {
	return EntityCoverageInfo{
		EntityName:                   entityName,
		UncoveredRelationships:       []string{},
		UncoveredAttributes:          []string{},
		CoverageAssertionsPercent:    make(map[string]int),
		UncoveredAssertions:          make(map[string][]string),
		CoverageRelationshipsPercent: 0,
		CoverageAttributesPercent:    0,
	}
}

// findUncoveredRelationships finds relationships that are not covered in the shape
func findUncoveredRelationships(entityName string, expected, actual []string) []string {
	covered := extractCoveredRelationships(entityName, actual)
	uncovered := []string{}

	for _, relationship := range expected {
		if !slices.Contains(covered, relationship) {
			uncovered = append(uncovered, relationship)
		}
	}

	return uncovered
}

// findUncoveredAttributes finds attributes that are not covered in the shape
func findUncoveredAttributes(entityName string, expected, actual []string) []string {
	covered := extractCoveredAttributes(entityName, actual)
	uncovered := []string{}

	for _, attr := range expected {
		if !slices.Contains(covered, attr) {
			uncovered = append(uncovered, attr)
		}
	}

	return uncovered
}

// findUncoveredAssertions finds assertions that are not covered in the shape
func findUncoveredAssertions(entityName string, expected []string, checks []file.Check, filters []file.EntityFilter) []string {
	covered := extractCoveredAssertions(entityName, checks, filters)
	uncovered := []string{}

	for _, assertion := range expected {
		if !slices.Contains(covered, assertion) {
			uncovered = append(uncovered, assertion)
		}
	}

	return uncovered
}

// buildSchemaCoverageInfo builds the final SchemaCoverageInfo with total coverage
func buildSchemaCoverageInfo(entityCoverageInfos []EntityCoverageInfo) SchemaCoverageInfo {
	relationshipsCoverage, attributesCoverage, assertionsCoverage := calculateTotalCoverage(entityCoverageInfos)

	return SchemaCoverageInfo{
		EntityCoverageInfo:         entityCoverageInfos,
		TotalRelationshipsCoverage: relationshipsCoverage,
		TotalAttributesCoverage:    attributesCoverage,
		TotalAssertionsCoverage:    assertionsCoverage,
	}
}

// calculateCoveragePercent calculates coverage percentage based on total and uncovered elements
func calculateCoveragePercent(totalElements, uncoveredElements []string) int {
	totalCount := len(totalElements)
	if totalCount == 0 {
		return 100
	}

	coveredCount := totalCount - len(uncoveredElements)
	return (coveredCount * 100) / totalCount
}

// calculateTotalCoverage calculates average coverage percentages across all entities
func calculateTotalCoverage(entities []EntityCoverageInfo) (int, int, int) {
	var (
		totalRelationships        int
		totalCoveredRelationships int
		totalAttributes           int
		totalCoveredAttributes    int
		totalAssertions           int
		totalCoveredAssertions    int
	)

	for _, entity := range entities {
		totalRelationships++
		totalCoveredRelationships += entity.CoverageRelationshipsPercent

		totalAttributes++
		totalCoveredAttributes += entity.CoverageAttributesPercent

		for _, assertionPercent := range entity.CoverageAssertionsPercent {
			totalAssertions++
			totalCoveredAssertions += assertionPercent
		}
	}

	return calculateAverageCoverage(totalRelationships, totalCoveredRelationships),
		calculateAverageCoverage(totalAttributes, totalCoveredAttributes),
		calculateAverageCoverage(totalAssertions, totalCoveredAssertions)
}

// calculateAverageCoverage calculates average coverage with zero-division guard
func calculateAverageCoverage(total, covered int) int {
	if total == 0 {
		return 100
	}
	return covered / total
}

// extractCoveredRelationships extracts covered relationships for a given entity from the shape
func extractCoveredRelationships(entityName string, relationships []string) []string {
	covered := []string{}

	for _, relationship := range relationships {
		tup, err := tuple.Tuple(relationship)
		if err != nil {
			continue
		}

		if tup.GetEntity().GetType() != entityName {
			continue
		}

		formatted := formatRelationship(
			tup.GetEntity().GetType(),
			tup.GetRelation(),
			tup.GetSubject().GetType(),
			tup.GetSubject().GetRelation(),
		)
		covered = append(covered, formatted)
	}

	return covered
}

// extractCoveredAttributes extracts covered attributes for a given entity from the shape
func extractCoveredAttributes(entityName string, attributes []string) []string {
	covered := []string{}

	for _, attrStr := range attributes {
		a, err := attribute.Attribute(attrStr)
		if err != nil {
			continue
		}

		if a.GetEntity().GetType() != entityName {
			continue
		}

		formatted := formatAttribute(a.GetEntity().GetType(), a.GetAttribute())
		covered = append(covered, formatted)
	}

	return covered
}

// extractCoveredAssertions extracts covered assertions for a given entity from checks and filters
func extractCoveredAssertions(entityName string, checks []file.Check, filters []file.EntityFilter) []string {
	covered := []string{}

	// Extract from checks
	for _, check := range checks {
		entity, err := tuple.E(check.Entity)
		if err != nil {
			continue
		}

		if entity.GetType() != entityName {
			continue
		}

		for permission := range check.Assertions {
			formatted := formatAssertion(entity.GetType(), permission)
			covered = append(covered, formatted)
		}
	}

	// Extract from entity filters
	for _, filter := range filters {
		if filter.EntityType != entityName {
			continue
		}

		for permission := range filter.Assertions {
			formatted := formatAssertion(filter.EntityType, permission)
			covered = append(covered, formatted)
		}
	}

	return covered
}

// formatRelationship formats a relationship string
func formatRelationship(entityName, relationName, subjectType, subjectRelation string) string {
	if subjectRelation != "" {
		return fmt.Sprintf("%s#%s@%s#%s", entityName, relationName, subjectType, subjectRelation)
	}
	return fmt.Sprintf("%s#%s@%s", entityName, relationName, subjectType)
}

// formatAttribute formats an attribute string
func formatAttribute(entityName, attributeName string) string {
	return fmt.Sprintf("%s#%s", entityName, attributeName)
}

// formatAssertion formats an assertion/permission string
func formatAssertion(entityName, permissionName string) string {
	return fmt.Sprintf("%s#%s", entityName, permissionName)
}
