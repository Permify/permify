package coverage

import (
	"fmt"
	"slices"
	"strings"

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

// EntityCoverageInfo represents coverage information for a single entity
type EntityCoverageInfo struct {
	EntityName string

	UncoveredRelationships       []string
	CoverageRelationshipsPercent int

	UncoveredAttributes       []string
	CoverageAttributesPercent int

	UncoveredAssertions       map[string][]string
	CoverageAssertionsPercent map[string]int

	UncoveredAssertionComponents       map[string][]string
	CoverageAssertionComponentsPercent map[string]int
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
	EntityName          string
	Relationships       []string
	Attributes          []string
	Assertions          []string
	AssertionComponents []string
}

type assertionCoverageContext struct {
	EntityID string
	Subject  *base.Subject
}

// Run analyzes the coverage of relationships, attributes, and assertions
// for a given schema shape and returns the coverage information
func Run(shape file.Shape) SchemaCoverageInfo {
	definitions, err := parseAndCompileSchema(shape.Schema)
	if err != nil {
		return SchemaCoverageInfo{}
	}

	refs := extractSchemaReferences(definitions)
	entityCoverageInfos := calculateEntityCoverages(refs, shape)

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
		EntityName:          entity.GetName(),
		Relationships:       extractRelationships(entity),
		Attributes:          extractAttributes(entity),
		Assertions:          extractAssertions(entity),
		AssertionComponents: extractAssertionComponents(entity),
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

// extractAssertionComponents extracts permission condition leaves from an entity.
func extractAssertionComponents(entity *base.EntityDefinition) []string {
	components := []string{}
	permissions := make(map[string]*base.Child)

	for _, permission := range entity.GetPermissions() {
		permissions[permission.GetName()] = permission.GetChild()
	}

	for _, permission := range entity.GetPermissions() {
		components = append(components, extractChildAssertionComponents(
			entity.GetName(),
			permission.GetName(),
			permission.GetChild(),
			permissions,
			map[string]bool{permission.GetName(): true},
		)...)
	}

	return components
}

func extractChildAssertionComponents(entityName, permissionName string, child *base.Child, permissions map[string]*base.Child, visited map[string]bool) []string {
	if child == nil {
		return []string{formatAssertion(entityName, permissionName)}
	}

	if leaf := child.GetLeaf(); leaf != nil {
		if computed := leaf.GetComputedUserSet(); computed != nil {
			relationName := computed.GetRelation()
			if nestedChild, ok := permissions[relationName]; ok && !visited[relationName] {
				nextVisited := cloneVisitedPermissions(visited)
				nextVisited[relationName] = true
				return extractChildAssertionComponents(entityName, permissionName, nestedChild, permissions, nextVisited)
			}
		}
		if component := formatLeafAssertionComponent(entityName, permissionName, leaf); component != "" {
			return []string{component}
		}
		return []string{formatAssertion(entityName, permissionName)}
	}

	if rewrite := child.GetRewrite(); rewrite != nil {
		components := []string{}
		for _, rewriteChild := range rewrite.GetChildren() {
			components = append(components, extractChildAssertionComponents(entityName, permissionName, rewriteChild, permissions, visited)...)
		}
		return components
	}

	return []string{formatAssertion(entityName, permissionName)}
}

func cloneVisitedPermissions(visited map[string]bool) map[string]bool {
	cloned := make(map[string]bool, len(visited)+1)
	for name, isVisited := range visited {
		cloned[name] = isVisited
	}
	return cloned
}

func formatLeafAssertionComponent(entityName, permissionName string, leaf *base.Leaf) string {
	switch leaf.GetType().(type) {
	case *base.Leaf_ComputedUserSet:
		if computed := leaf.GetComputedUserSet(); computed != nil {
			return formatAssertionComponent(entityName, permissionName, computed.GetRelation())
		}
	case *base.Leaf_TupleToUserSet:
		tupleToUserSet := leaf.GetTupleToUserSet()
		if tupleToUserSet != nil && tupleToUserSet.GetTupleSet() != nil && tupleToUserSet.GetComputed() != nil {
			return formatAssertionComponent(
				entityName,
				permissionName,
				fmt.Sprintf("%s.%s", tupleToUserSet.GetTupleSet().GetRelation(), tupleToUserSet.GetComputed().GetRelation()),
			)
		}
	case *base.Leaf_ComputedAttribute:
		if computed := leaf.GetComputedAttribute(); computed != nil {
			return formatAssertionComponent(entityName, permissionName, computed.GetName())
		}
	case *base.Leaf_Call:
		if call := leaf.GetCall(); call != nil {
			return formatAssertionComponent(entityName, permissionName, fmt.Sprintf("%s()", call.GetRuleName()))
		}
	default:
		return ""
	}
	return ""
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

		uncoveredComponents := findUncoveredAssertionComponents(
			ref.EntityName,
			ref.AssertionComponents,
			scenario.Checks,
			scenario.EntityFilters,
			shape.Relationships,
			shape.Attributes,
		)
		if len(uncoveredComponents) > 0 {
			entityCoverageInfo.UncoveredAssertionComponents[scenario.Name] = uncoveredComponents
		}
		entityCoverageInfo.CoverageAssertionComponentsPercent[scenario.Name] = calculateCoveragePercent(
			ref.AssertionComponents,
			uncoveredComponents,
		)
	}

	return entityCoverageInfo
}

// newEntityCoverageInfo creates a new EntityCoverageInfo with initialized fields
func newEntityCoverageInfo(entityName string) EntityCoverageInfo {
	return EntityCoverageInfo{
		EntityName:                         entityName,
		UncoveredRelationships:             []string{},
		UncoveredAttributes:                []string{},
		CoverageAssertionsPercent:          make(map[string]int),
		UncoveredAssertions:                make(map[string][]string),
		CoverageAssertionComponentsPercent: make(map[string]int),
		UncoveredAssertionComponents:       make(map[string][]string),
		CoverageRelationshipsPercent:       0,
		CoverageAttributesPercent:          0,
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

func findUncoveredAssertionComponents(entityName string, expected []string, checks []file.Check, filters []file.EntityFilter, relationships, attributes []string) []string {
	coverageContexts := extractAssertionCoverageContexts(entityName, checks, filters)
	uncovered := []string{}

	for _, component := range expected {
		assertion := assertionFromComponent(component)
		if !assertionComponentHasSupportingData(entityName, component, coverageContexts[assertion], relationships, attributes) {
			uncovered = append(uncovered, component)
		}
	}

	return uncovered
}

func extractAssertionCoverageContexts(entityName string, checks []file.Check, filters []file.EntityFilter) map[string][]assertionCoverageContext {
	contexts := make(map[string][]assertionCoverageContext)

	for _, check := range checks {
		entity, err := tuple.E(check.Entity)
		if err != nil || entity.GetType() != entityName {
			continue
		}

		subject := parseSubject(check.Subject)
		for permission := range check.Assertions {
			assertion := formatAssertion(entity.GetType(), permission)
			contexts[assertion] = append(contexts[assertion], assertionCoverageContext{
				EntityID: entity.GetId(),
				Subject:  subject,
			})
		}
	}

	for _, filter := range filters {
		if filter.EntityType != entityName {
			continue
		}

		subject := parseSubject(filter.Subject)
		for permission, entityIDs := range filter.Assertions {
			assertion := formatAssertion(filter.EntityType, permission)
			for _, entityID := range entityIDs {
				contexts[assertion] = append(contexts[assertion], assertionCoverageContext{
					EntityID: entityID,
					Subject:  subject,
				})
			}
		}
	}

	return contexts
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

		for _, assertionPercent := range entity.CoverageAssertionComponentsPercent {
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

func formatAssertionComponent(entityName, permissionName, componentName string) string {
	return fmt.Sprintf("%s#%s[%s]", entityName, permissionName, componentName)
}

func assertionFromComponent(component string) string {
	for idx, ch := range component {
		if ch == '[' {
			return component[:idx]
		}
	}
	return component
}

func componentName(component string) string {
	start := -1
	for idx, ch := range component {
		if ch == '[' {
			start = idx + 1
			continue
		}
		if ch == ']' && start >= 0 {
			return component[start:idx]
		}
	}
	return ""
}

func assertionComponentHasSupportingData(entityName, component string, contexts []assertionCoverageContext, relationships, attributes []string) bool {
	if len(contexts) == 0 {
		return false
	}

	name := componentName(component)
	if name == "" {
		return true
	}

	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		return tupleToUserSetComponentCovered(entityName, parts[0], parts[1], contexts, relationships)
	}

	if strings.HasSuffix(name, "()") {
		return true
	}

	for _, context := range contexts {
		if attributeCovered(entityName, context.EntityID, name, attributes) || relationshipRelationCovered(entityName, context.EntityID, name, relationships) {
			return true
		}
	}

	return false
}

func relationshipRelationCovered(entityName, entityID, relationName string, relationships []string) bool {
	for _, relationship := range relationships {
		tup, err := tuple.Tuple(relationship)
		if err != nil {
			continue
		}
		if tup.GetEntity().GetType() == entityName && tup.GetEntity().GetId() == entityID && tup.GetRelation() == relationName {
			return true
		}
	}
	return false
}

func attributeCovered(entityName, entityID, attributeName string, attributes []string) bool {
	for _, attrStr := range attributes {
		attr, err := attribute.Attribute(attrStr)
		if err != nil {
			continue
		}
		if attr.GetEntity().GetType() == entityName && attr.GetEntity().GetId() == entityID && attr.GetAttribute() == attributeName {
			return true
		}
	}
	return false
}

func tupleToUserSetComponentCovered(entityName, tupleRelationName, computedRelationName string, contexts []assertionCoverageContext, relationships []string) bool {
	for _, context := range contexts {
		for _, relationship := range relationships {
			tup, err := tuple.Tuple(relationship)
			if err != nil {
				continue
			}
			if tup.GetEntity().GetType() != entityName || tup.GetEntity().GetId() != context.EntityID || tup.GetRelation() != tupleRelationName {
				continue
			}
			if computedRelationCovered(tup.GetSubject(), computedRelationName, context.Subject, relationships) {
				return true
			}
		}
	}
	return false
}

func computedRelationCovered(source *base.Subject, computedRelationName string, expectedSubject *base.Subject, relationships []string) bool {
	for _, relationship := range relationships {
		tup, err := tuple.Tuple(relationship)
		if err != nil {
			continue
		}
		if tup.GetEntity().GetType() != source.GetType() || tup.GetEntity().GetId() != source.GetId() || tup.GetRelation() != computedRelationName {
			continue
		}
		if expectedSubject == nil || tuple.AreSubjectsEqual(tup.GetSubject(), expectedSubject) {
			return true
		}
	}
	return false
}

func parseSubject(subject string) *base.Subject {
	ear, err := tuple.EAR(subject)
	if err != nil {
		return nil
	}
	return &base.Subject{
		Type:     ear.GetEntity().GetType(),
		Id:       ear.GetEntity().GetId(),
		Relation: ear.GetRelation(),
	}
}
