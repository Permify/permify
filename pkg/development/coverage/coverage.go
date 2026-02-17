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

// SchemaCoverageInfo represents the overall coverage information for a schema
type SchemaCoverageInfo struct {
	EntityCoverageInfo         []EntityCoverageInfo
	TotalRelationshipsCoverage int
	TotalAttributesCoverage    int
	TotalAssertionsCoverage    int
}

// ConditionComponent represents a single component (leaf) in a permission condition tree
type ConditionComponent struct {
	Name string // The identifier, e.g. "owner", "parent.admin", "is_public"
	Type string // "relation", "tuple_to_userset", "attribute", "call"
}

// ConditionCoverageInfo represents the coverage information for a single permission's condition
type ConditionCoverageInfo struct {
	PermissionName      string
	AllComponents       []ConditionComponent
	CoveredComponents   []ConditionComponent
	UncoveredComponents []ConditionComponent
	CoveragePercent     int
}

// EntityCoverageInfo represents coverage information for a single entity
type EntityCoverageInfo struct {
	EntityName string

	UncoveredRelationships       []string
	CoverageRelationshipsPercent int

	UncoveredAttributes       []string
	CoverageAttributesPercent int

	UncoveredAssertions       map[string][]string
	CoverageAssertionsPercent map[string]int

	// PermissionConditionCoverage maps scenario name -> permission name -> condition coverage info
	PermissionConditionCoverage map[string]map[string]ConditionCoverageInfo
}

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
	definitions, err := parseAndCompileSchema(shape.Schema)
	if err != nil {
		return SchemaCoverageInfo{}
	}

	refs := extractSchemaReferences(definitions)
	entityCoverageInfos := calculateEntityCoverages(refs, shape, definitions)

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
func calculateEntityCoverages(refs []SchemaCoverage, shape file.Shape, definitions []*base.EntityDefinition) []EntityCoverageInfo {
	entityCoverageInfos := []EntityCoverageInfo{}

	// Build a map from entity name to definition for condition coverage
	defMap := make(map[string]*base.EntityDefinition)
	for _, def := range definitions {
		defMap[def.GetName()] = def
	}

	for _, ref := range refs {
		entityDef := defMap[ref.EntityName]
		entityCoverageInfo := calculateEntityCoverage(ref, shape, entityDef)
		entityCoverageInfos = append(entityCoverageInfos, entityCoverageInfo)
	}

	return entityCoverageInfos
}

// calculateEntityCoverage calculates coverage for a single entity
func calculateEntityCoverage(ref SchemaCoverage, shape file.Shape, entityDef *base.EntityDefinition) EntityCoverageInfo {
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

	// Calculate assertions coverage and condition coverage for each scenario
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

		// Calculate condition coverage for permissions that are asserted in this scenario
		if entityDef != nil {
			condCov := calculateConditionCoverage(
				entityDef,
				ref.EntityName,
				scenario,
				shape.Relationships,
				shape.Attributes,
			)
			if len(condCov) > 0 {
				entityCoverageInfo.PermissionConditionCoverage[scenario.Name] = condCov
			}
		}
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
		PermissionConditionCoverage:  make(map[string]map[string]ConditionCoverageInfo),
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

// calculateConditionCoverage analyzes the condition tree for each permission that is asserted
// in the given scenario, and checks which leaf components are covered by the test data.
func calculateConditionCoverage(
	entityDef *base.EntityDefinition,
	entityName string,
	scenario file.Scenario,
	relationships []string,
	attributes []string,
) map[string]ConditionCoverageInfo {
	result := make(map[string]ConditionCoverageInfo)

	// Find which permissions are asserted for this entity in this scenario
	assertedPerms := extractAssertedPermissions(entityName, scenario)
	if len(assertedPerms) == 0 {
		return result
	}

	// Build sets of covered relations and attributes for this entity
	coveredRelations := buildCoveredRelationSet(entityName, relationships)
	coveredAttributes := buildCoveredAttributeSet(entityName, attributes)

	// For each asserted permission, walk its condition tree
	for _, permName := range assertedPerms {
		permDef, ok := entityDef.GetPermissions()[permName]
		if !ok || permDef.GetChild() == nil {
			continue
		}

		allComponents := extractConditionComponents(permDef.GetChild())
		if len(allComponents) == 0 {
			continue
		}

		var covered []ConditionComponent
		var uncovered []ConditionComponent

		for _, comp := range allComponents {
			if isComponentCovered(comp, coveredRelations, coveredAttributes) {
				covered = append(covered, comp)
			} else {
				uncovered = append(uncovered, comp)
			}
		}

		coveragePercent := (len(covered) * 100) / len(allComponents)

		result[permName] = ConditionCoverageInfo{
			PermissionName:      permName,
			AllComponents:       allComponents,
			CoveredComponents:   covered,
			UncoveredComponents: uncovered,
			CoveragePercent:     coveragePercent,
		}
	}

	return result
}

// extractAssertedPermissions finds all permission names that are asserted for the given entity
// in a scenario's checks and entity filters.
func extractAssertedPermissions(entityName string, scenario file.Scenario) []string {
	seen := make(map[string]bool)
	var perms []string

	for _, check := range scenario.Checks {
		entity, err := tuple.E(check.Entity)
		if err != nil || entity.GetType() != entityName {
			continue
		}
		for permName := range check.Assertions {
			if !seen[permName] {
				seen[permName] = true
				perms = append(perms, permName)
			}
		}
	}

	for _, filter := range scenario.EntityFilters {
		if filter.EntityType != entityName {
			continue
		}
		for permName := range filter.Assertions {
			if !seen[permName] {
				seen[permName] = true
				perms = append(perms, permName)
			}
		}
	}

	return perms
}

// extractConditionComponents recursively walks a Child tree and extracts all leaf components
func extractConditionComponents(child *base.Child) []ConditionComponent {
	if child == nil {
		return nil
	}

	// Check if this is a leaf node
	if leaf := child.GetLeaf(); leaf != nil {
		comp := leafToComponent(leaf)
		if comp != nil {
			return []ConditionComponent{*comp}
		}
		return nil
	}

	// Check if this is a rewrite (union/intersection/exclusion)
	if rewrite := child.GetRewrite(); rewrite != nil {
		var components []ConditionComponent
		for _, ch := range rewrite.GetChildren() {
			components = append(components, extractConditionComponents(ch)...)
		}
		return components
	}

	return nil
}

// leafToComponent converts a Leaf node to a ConditionComponent
func leafToComponent(leaf *base.Leaf) *ConditionComponent {
	if cus := leaf.GetComputedUserSet(); cus != nil {
		return &ConditionComponent{
			Name: cus.GetRelation(),
			Type: "relation",
		}
	}

	if ttus := leaf.GetTupleToUserSet(); ttus != nil {
		tupleSetRelation := ""
		if ttus.GetTupleSet() != nil {
			tupleSetRelation = ttus.GetTupleSet().GetRelation()
		}
		computedRelation := ""
		if ttus.GetComputed() != nil {
			computedRelation = ttus.GetComputed().GetRelation()
		}
		name := tupleSetRelation
		if computedRelation != "" {
			name = tupleSetRelation + "." + computedRelation
		}
		return &ConditionComponent{
			Name: name,
			Type: "tuple_to_userset",
		}
	}

	if ca := leaf.GetComputedAttribute(); ca != nil {
		return &ConditionComponent{
			Name: ca.GetName(),
			Type: "attribute",
		}
	}

	if call := leaf.GetCall(); call != nil {
		return &ConditionComponent{
			Name: call.GetRuleName(),
			Type: "call",
		}
	}

	return nil
}

// buildCoveredRelationSet builds a set of relation names that are covered by test relationships
// for a given entity. Returns a map[relationName]bool.
func buildCoveredRelationSet(entityName string, relationships []string) map[string]bool {
	covered := make(map[string]bool)

	for _, rel := range relationships {
		tup, err := tuple.Tuple(rel)
		if err != nil {
			continue
		}
		if tup.GetEntity().GetType() == entityName {
			covered[tup.GetRelation()] = true
		}
	}

	return covered
}

// buildCoveredAttributeSet builds a set of attribute names that are covered by test attributes
// for a given entity. Returns a map[attributeName]bool.
func buildCoveredAttributeSet(entityName string, attributes []string) map[string]bool {
	covered := make(map[string]bool)

	for _, attrStr := range attributes {
		a, err := attribute.Attribute(attrStr)
		if err != nil {
			continue
		}
		if a.GetEntity().GetType() == entityName {
			covered[a.GetAttribute()] = true
		}
	}

	return covered
}

// isComponentCovered checks if a condition component is covered by the test data.
// A "relation" component is covered if the relation name exists in coveredRelations.
// A "tuple_to_userset" component is covered if the tuple set relation exists in coveredRelations.
// An "attribute" component is covered if the attribute name exists in coveredAttributes.
// A "call" component is always considered covered (rule calls depend on runtime evaluation).
func isComponentCovered(comp ConditionComponent, coveredRelations, coveredAttributes map[string]bool) bool {
	switch comp.Type {
	case "relation":
		return coveredRelations[comp.Name]
	case "tuple_to_userset":
		// For tuple_to_userset like "parent.admin", check that the tuple set relation ("parent") is covered
		tupleSetRelation := comp.Name
		// If it contains a dot, split to get the tuple set relation
		for i, c := range comp.Name {
			if c == '.' {
				tupleSetRelation = comp.Name[:i]
				break
			}
		}
		return coveredRelations[tupleSetRelation]
	case "attribute":
		return coveredAttributes[comp.Name]
	case "call":
		// Rule calls are considered covered if they can be evaluated
		return true
	}
	return false
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
