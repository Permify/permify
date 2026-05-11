package coverage

import (
	"fmt"
	"slices"
	"sort"
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

const (
	componentRelation       = "relation"
	componentAttribute      = "attribute"
	componentTupleToUserset = "tuple_to_userset"
	componentCall           = "call"
	componentPermission     = "permission"
)

// ConditionComponent represents a single leaf component in a permission condition tree.
type ConditionComponent struct {
	Name string
	Type string
}

// ConditionCoverageInfo represents condition component coverage for one permission.
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
	definitionsByEntity := make(map[string]*base.EntityDefinition, len(definitions))

	for _, definition := range definitions {
		definitionsByEntity[definition.GetName()] = definition
	}

	for _, ref := range refs {
		entityCoverageInfo := calculateEntityCoverage(ref, shape, definitionsByEntity[ref.EntityName])
		entityCoverageInfos = append(entityCoverageInfos, entityCoverageInfo)
	}

	return entityCoverageInfos
}

// calculateEntityCoverage calculates coverage for a single entity
func calculateEntityCoverage(ref SchemaCoverage, shape file.Shape, entityDefinition *base.EntityDefinition) EntityCoverageInfo {
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

		conditionCoverage := calculateConditionCoverage(
			entityDefinition,
			scenario,
			shape.Relationships,
			shape.Attributes,
		)
		if len(conditionCoverage) > 0 {
			entityCoverageInfo.PermissionConditionCoverage[scenario.Name] = conditionCoverage
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

// calculateConditionCoverage analyzes asserted permissions in a scenario and reports which
// leaf components inside each permission condition have no matching test data.
func calculateConditionCoverage(
	entityDefinition *base.EntityDefinition,
	scenario file.Scenario,
	relationships []string,
	attributes []string,
) map[string]ConditionCoverageInfo {
	result := make(map[string]ConditionCoverageInfo)
	if entityDefinition == nil {
		return result
	}

	assertedPermissions := extractAssertedPermissions(entityDefinition.GetName(), scenario)
	if len(assertedPermissions) == 0 {
		return result
	}

	coverageData := newComponentCoverageData(relationships, attributes, scenario)

	for _, permissionName := range sortedPermissionNames(assertedPermissions) {
		permission, ok := entityDefinition.GetPermissions()[permissionName]
		if !ok || permission.GetChild() == nil {
			continue
		}

		components := extractConditionComponents(entityDefinition, permission.GetChild(), map[string]bool{
			permissionName: true,
		})
		components = uniqueConditionComponents(components)
		if len(components) == 0 {
			continue
		}

		targets := assertedPermissions[permissionName]
		covered := make([]ConditionComponent, 0, len(components))
		uncovered := make([]ConditionComponent, 0)

		for _, component := range components {
			if coverageData.isComponentCovered(entityDefinition.GetName(), component, targets) {
				covered = append(covered, component)
			} else {
				uncovered = append(uncovered, component)
			}
		}

		result[permissionName] = ConditionCoverageInfo{
			PermissionName:      permissionName,
			AllComponents:       components,
			CoveredComponents:   covered,
			UncoveredComponents: uncovered,
			CoveragePercent:     calculateCoveragePercent(conditionComponentNames(components), conditionComponentNames(uncovered)),
		}
	}

	return result
}

type assertionTarget struct {
	EntityID     string
	EntityIDOnly bool
}

// extractAssertedPermissions returns permissions asserted for an entity and the concrete
// entity IDs they were asserted against when the scenario provides them.
func extractAssertedPermissions(entityName string, scenario file.Scenario) map[string][]assertionTarget {
	asserted := make(map[string][]assertionTarget)

	for _, check := range scenario.Checks {
		entity, err := tuple.E(check.Entity)
		if err != nil || entity.GetType() != entityName {
			continue
		}

		for permission := range check.Assertions {
			asserted[permission] = appendAssertionTarget(asserted[permission], assertionTarget{
				EntityID:     entity.GetId(),
				EntityIDOnly: true,
			})
		}
	}

	for _, filter := range scenario.EntityFilters {
		if filter.EntityType != entityName {
			continue
		}

		for permission := range filter.Assertions {
			asserted[permission] = appendAssertionTarget(asserted[permission], assertionTarget{})
		}
	}

	return asserted
}

func appendAssertionTarget(targets []assertionTarget, target assertionTarget) []assertionTarget {
	for _, existing := range targets {
		if existing == target {
			return targets
		}
	}
	return append(targets, target)
}

func sortedPermissionNames(asserted map[string][]assertionTarget) []string {
	names := make([]string, 0, len(asserted))
	for name := range asserted {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func extractConditionComponents(
	entityDefinition *base.EntityDefinition,
	child *base.Child,
	visitedPermissions map[string]bool,
) []ConditionComponent {
	if child == nil {
		return nil
	}

	if leaf := child.GetLeaf(); leaf != nil {
		return leafToComponents(entityDefinition, leaf, visitedPermissions)
	}

	if rewrite := child.GetRewrite(); rewrite != nil {
		components := []ConditionComponent{}
		for _, child := range rewrite.GetChildren() {
			components = append(components, extractConditionComponents(entityDefinition, child, visitedPermissions)...)
		}
		return components
	}

	return nil
}

func leafToComponents(
	entityDefinition *base.EntityDefinition,
	leaf *base.Leaf,
	visitedPermissions map[string]bool,
) []ConditionComponent {
	if computedUserSet := leaf.GetComputedUserSet(); computedUserSet != nil {
		name := computedUserSet.GetRelation()
		if entityDefinition.GetReferences()[name] == base.EntityDefinition_REFERENCE_PERMISSION {
			nestedPermission := entityDefinition.GetPermissions()[name]
			if nestedPermission != nil && !visitedPermissions[name] {
				visitedPermissions[name] = true
				components := extractConditionComponents(entityDefinition, nestedPermission.GetChild(), visitedPermissions)
				delete(visitedPermissions, name)
				return components
			}

			return []ConditionComponent{{
				Name: name,
				Type: componentPermission,
			}}
		}

		return []ConditionComponent{{
			Name: name,
			Type: componentRelation,
		}}
	}

	if tupleToUserset := leaf.GetTupleToUserSet(); tupleToUserset != nil {
		tupleSetRelation := ""
		if tupleToUserset.GetTupleSet() != nil {
			tupleSetRelation = tupleToUserset.GetTupleSet().GetRelation()
		}

		computedRelation := ""
		if tupleToUserset.GetComputed() != nil {
			computedRelation = tupleToUserset.GetComputed().GetRelation()
		}

		return []ConditionComponent{{
			Name: strings.Join(nonEmptyStrings(tupleSetRelation, computedRelation), "."),
			Type: componentTupleToUserset,
		}}
	}

	if computedAttribute := leaf.GetComputedAttribute(); computedAttribute != nil {
		return []ConditionComponent{{
			Name: computedAttribute.GetName(),
			Type: componentAttribute,
		}}
	}

	if call := leaf.GetCall(); call != nil {
		components := []ConditionComponent{{
			Name: call.GetRuleName(),
			Type: componentCall,
		}}

		for _, argument := range call.GetArguments() {
			if computedAttribute := argument.GetComputedAttribute(); computedAttribute != nil {
				components = append(components, ConditionComponent{
					Name: computedAttribute.GetName(),
					Type: componentAttribute,
				})
			}
		}

		return components
	}

	return nil
}

func nonEmptyStrings(values ...string) []string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func uniqueConditionComponents(components []ConditionComponent) []ConditionComponent {
	seen := make(map[ConditionComponent]bool, len(components))
	unique := make([]ConditionComponent, 0, len(components))

	for _, component := range components {
		if component.Name == "" || seen[component] {
			continue
		}
		seen[component] = true
		unique = append(unique, component)
	}

	return unique
}

func conditionComponentNames(components []ConditionComponent) []string {
	names := make([]string, 0, len(components))
	for _, component := range components {
		names = append(names, fmt.Sprintf("%s:%s", component.Type, component.Name))
	}
	return names
}

type componentCoverageData struct {
	relationships map[string]map[string]map[string]bool
	attributes    map[string]map[string]map[string]bool
}

func newComponentCoverageData(relationships, attributes []string, scenario file.Scenario) componentCoverageData {
	data := componentCoverageData{
		relationships: make(map[string]map[string]map[string]bool),
		attributes:    make(map[string]map[string]map[string]bool),
	}

	for _, relationship := range relationships {
		data.addRelationship(relationship)
	}

	for _, attribute := range attributes {
		data.addAttribute(attribute)
	}

	for _, check := range scenario.Checks {
		data.addContext(check.Context)
	}

	for _, filter := range scenario.EntityFilters {
		data.addContext(filter.Context)
	}

	for _, filter := range scenario.SubjectFilters {
		data.addContext(filter.Context)
	}

	return data
}

func (data componentCoverageData) addContext(context file.Context) {
	for _, relationship := range context.Tuples {
		data.addRelationship(relationship)
	}

	for _, attribute := range context.Attributes {
		data.addAttribute(attribute)
	}
}

func (data componentCoverageData) addRelationship(relationship string) {
	tup, err := tuple.Tuple(relationship)
	if err != nil {
		return
	}

	entityType := tup.GetEntity().GetType()
	relation := tup.GetRelation()
	entityID := tup.GetEntity().GetId()

	if data.relationships[entityType] == nil {
		data.relationships[entityType] = make(map[string]map[string]bool)
	}
	if data.relationships[entityType][relation] == nil {
		data.relationships[entityType][relation] = make(map[string]bool)
	}
	data.relationships[entityType][relation][entityID] = true
}

func (data componentCoverageData) addAttribute(attributeString string) {
	attr, err := attribute.Attribute(attributeString)
	if err != nil {
		return
	}

	entityType := attr.GetEntity().GetType()
	attributeName := attr.GetAttribute()
	entityID := attr.GetEntity().GetId()

	if data.attributes[entityType] == nil {
		data.attributes[entityType] = make(map[string]map[string]bool)
	}
	if data.attributes[entityType][attributeName] == nil {
		data.attributes[entityType][attributeName] = make(map[string]bool)
	}
	data.attributes[entityType][attributeName][entityID] = true
}

func (data componentCoverageData) isComponentCovered(entityName string, component ConditionComponent, targets []assertionTarget) bool {
	switch component.Type {
	case componentRelation:
		return data.hasRelationship(entityName, component.Name, targets)
	case componentTupleToUserset:
		tupleSetRelation, _, _ := strings.Cut(component.Name, ".")
		return data.hasRelationship(entityName, tupleSetRelation, targets)
	case componentAttribute:
		return data.hasAttribute(entityName, component.Name, targets)
	case componentCall:
		return true
	case componentPermission:
		return false
	default:
		return false
	}
}

func (data componentCoverageData) hasRelationship(entityName, relation string, targets []assertionTarget) bool {
	return hasCoverageForTargets(data.relationships[entityName][relation], targets)
}

func (data componentCoverageData) hasAttribute(entityName, attributeName string, targets []assertionTarget) bool {
	return hasCoverageForTargets(data.attributes[entityName][attributeName], targets)
}

func hasCoverageForTargets(coveredEntityIDs map[string]bool, targets []assertionTarget) bool {
	if len(coveredEntityIDs) == 0 {
		return false
	}

	for _, target := range targets {
		if !target.EntityIDOnly {
			return true
		}
		if coveredEntityIDs[target.EntityID] {
			return true
		}
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
