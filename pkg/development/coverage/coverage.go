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

type SchemaCoverageInfo struct {
	EntityCoverageInfo         []EntityCoverageInfo
	TotalRelationshipsCoverage int
	TotalAttributesCoverage    int
	TotalAssertionsCoverage    int
}

type EntityCoverageInfo struct {
	EntityName string

	UncoveredRelationships       []string
	CoverageRelationshipsPercent int

	UncoveredAttributes       []string
	CoverageAttributesPercent int

	UncoveredAssertions       map[string][]string
	CoverageAssertionsPercent map[string]int

	PermissionConditionCoverage map[string]map[string]*ConditionCoverageInfo
}

type ConditionCoverageInfo struct {
	PermissionName      string
	AllComponents       []ConditionComponent
	CoveredComponents   []ConditionComponent
	UncoveredComponents []ConditionComponent
	CoveragePercent     int
}

type ConditionComponent struct {
	Name string
	Type string
}

type SchemaCoverage struct {
	EntityName    string
	Relationships []string
	Attributes    []string
	Assertions    []string
}

func Run(shape file.Shape) SchemaCoverageInfo {
	definitions, err := parseAndCompileSchema(shape.Schema)
	if err != nil {
		return SchemaCoverageInfo{}
	}
	refs := extractSchemaReferences(definitions)
	entityCoverageInfos := calculateEntityCoverages(refs, shape, definitions)
	return buildSchemaCoverageInfo(entityCoverageInfos)
}

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

func extractSchemaReferences(definitions []*base.EntityDefinition) []SchemaCoverage {
	refs := make([]SchemaCoverage, len(definitions))
	for idx, entityDef := range definitions {
		refs[idx] = extractEntityReferences(entityDef)
	}
	return refs
}

func extractEntityReferences(entity *base.EntityDefinition) SchemaCoverage {
	return SchemaCoverage{
		EntityName:    entity.GetName(),
		Relationships: extractRelationships(entity),
		Attributes:    extractAttributes(entity),
		Assertions:    extractAssertions(entity),
	}
}

func extractRelationships(entity *base.EntityDefinition) []string {
	relationships := []string{}
	for _, relation := range entity.GetRelations() {
		for _, reference := range relation.GetRelationReferences() {
			formatted := formatRelationship(entity.GetName(), relation.GetName(), reference.GetType(), reference.GetRelation())
			relationships = append(relationships, formatted)
		}
	}
	return relationships
}

func extractAttributes(entity *base.EntityDefinition) []string {
	attributes := []string{}
	for _, attr := range entity.GetAttributes() {
		formatted := formatAttribute(entity.GetName(), attr.GetName())
		attributes = append(attributes, formatted)
	}
	return attributes
}

func extractAssertions(entity *base.EntityDefinition) []string {
	assertions := []string{}
	for _, permission := range entity.GetPermissions() {
		formatted := formatAssertion(entity.GetName(), permission.GetName())
		assertions = append(assertions, formatted)
	}
	return assertions
}

func calculateEntityCoverages(refs []SchemaCoverage, shape file.Shape, definitions []*base.EntityDefinition) []EntityCoverageInfo {
	entityCoverageInfos := []EntityCoverageInfo{}
	defMap := make(map[string]*base.EntityDefinition, len(definitions))
	for _, def := range definitions {
		defMap[def.GetName()] = def
	}
	for _, ref := range refs {
		entityCoverageInfo := calculateEntityCoverage(ref, shape, defMap[ref.EntityName])
		entityCoverageInfos = append(entityCoverageInfos, entityCoverageInfo)
	}
	return entityCoverageInfos
}

func calculateEntityCoverage(ref SchemaCoverage, shape file.Shape, entityDef *base.EntityDefinition) EntityCoverageInfo {
	entityCoverageInfo := newEntityCoverageInfo(ref.EntityName)
	entityCoverageInfo.UncoveredRelationships = findUncoveredRelationships(ref.EntityName, ref.Relationships, shape.Relationships)
	entityCoverageInfo.CoverageRelationshipsPercent = calculateCoveragePercent(ref.Relationships, entityCoverageInfo.UncoveredRelationships)
	entityCoverageInfo.UncoveredAttributes = findUncoveredAttributes(ref.EntityName, ref.Attributes, shape.Attributes)
	entityCoverageInfo.CoverageAttributesPercent = calculateCoveragePercent(ref.Attributes, entityCoverageInfo.UncoveredAttributes)

	for _, scenario := range shape.Scenarios {
		uncovered := findUncoveredAssertions(ref.EntityName, ref.Assertions, scenario.Checks, scenario.EntityFilters)
		if len(uncovered) > 0 {
			entityCoverageInfo.UncoveredAssertions[scenario.Name] = uncovered
		}
		entityCoverageInfo.CoverageAssertionsPercent[scenario.Name] = calculateCoveragePercent(ref.Assertions, uncovered)
		if entityDef != nil {
			conditionCoverage := calculateConditionCoverage(ref.EntityName, entityDef, scenario, shape.Relationships, shape.Attributes)
			if len(conditionCoverage) > 0 {
				entityCoverageInfo.PermissionConditionCoverage[scenario.Name] = conditionCoverage
			}
		}
	}
	return entityCoverageInfo
}

func newEntityCoverageInfo(entityName string) EntityCoverageInfo {
	return EntityCoverageInfo{
		EntityName:                  entityName,
		UncoveredRelationships:      []string{},
		UncoveredAttributes:         []string{},
		CoverageAssertionsPercent:   make(map[string]int),
		UncoveredAssertions:         make(map[string][]string),
		PermissionConditionCoverage: make(map[string]map[string]*ConditionCoverageInfo),
	}
}

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

func buildSchemaCoverageInfo(entityCoverageInfos []EntityCoverageInfo) SchemaCoverageInfo {
	relationshipsCoverage, attributesCoverage, assertionsCoverage := calculateTotalCoverage(entityCoverageInfos)
	return SchemaCoverageInfo{
		EntityCoverageInfo:         entityCoverageInfos,
		TotalRelationshipsCoverage: relationshipsCoverage,
		TotalAttributesCoverage:    attributesCoverage,
		TotalAssertionsCoverage:    assertionsCoverage,
	}
}

func calculateCoveragePercent(totalElements, uncoveredElements []string) int {
	totalCount := len(totalElements)
	if totalCount == 0 {
		return 100
	}
	coveredCount := totalCount - len(uncoveredElements)
	return (coveredCount * 100) / totalCount
}

func calculateTotalCoverage(entities []EntityCoverageInfo) (int, int, int) {
	var totalRelationships, totalCoveredRelationships, totalAttributes, totalCoveredAttributes, totalAssertions, totalCoveredAssertions int
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

func calculateAverageCoverage(total, covered int) int {
	if total == 0 {
		return 100
	}
	return covered / total
}

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
		formatted := formatRelationship(tup.GetEntity().GetType(), tup.GetRelation(), tup.GetSubject().GetType(), tup.GetSubject().GetRelation())
		covered = append(covered, formatted)
	}
	return covered
}

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

func extractCoveredAssertions(entityName string, checks []file.Check, filters []file.EntityFilter) []string {
	covered := []string{}
	for _, check := range checks {
		entity, err := tuple.E(check.Entity)
		if err != nil {
			continue
		}
		if entity.GetType() != entityName {
			continue
		}
		for permission := range check.Assertions {
			covered = append(covered, formatAssertion(entity.GetType(), permission))
		}
	}
	for _, filter := range filters {
		if filter.EntityType != entityName {
			continue
		}
		for permission := range filter.Assertions {
			covered = append(covered, formatAssertion(filter.EntityType, permission))
		}
	}
	return covered
}

func formatRelationship(entityName, relationName, subjectType, subjectRelation string) string {
	if subjectRelation != "" {
		return fmt.Sprintf("%s#%s@%s#%s", entityName, relationName, subjectType, subjectRelation)
	}
	return fmt.Sprintf("%s#%s@%s", entityName, relationName, subjectType)
}

func formatAttribute(entityName, attributeName string) string {
	return fmt.Sprintf("%s#%s", entityName, attributeName)
}

func formatAssertion(entityName, permissionName string) string {
	return fmt.Sprintf("%s#%s", entityName, permissionName)
}

func calculateConditionCoverage(entityName string, entityDef *base.EntityDefinition, scenario file.Scenario, relationships []string, attributes []string) map[string]*ConditionCoverageInfo {
	result := make(map[string]*ConditionCoverageInfo)
	assertedPermissions := extractAssertedPermissions(entityName, scenario)
	for _, perm := range entityDef.GetPermissions() {
		permName := perm.GetName()
		if _, ok := assertedPermissions[permName]; !ok {
			continue
		}
		components := extractConditionComponents(perm.GetChild())
		if len(components) == 0 {
			continue
		}
		coveredRelations := buildCoveredRelationSet(entityName, relationships)
		coveredAttrs := buildCoveredAttributeSet(entityName, attributes)
		for _, check := range scenario.Checks {
			entity, err := tuple.E(check.Entity)
			if err != nil || entity.GetType() != entityName {
				continue
			}
			for _, ctxTuple := range check.Context.Tuples {
				tup, err := tuple.Tuple(ctxTuple)
				if err != nil {
					continue
				}
				if tup.GetEntity().GetType() == entityName {
					coveredRelations[tup.GetRelation()] = true
				}
			}
			for _, ctxAttr := range check.Context.Attributes {
				a, err := attribute.Attribute(ctxAttr)
				if err != nil {
					continue
				}
				if a.GetEntity().GetType() == entityName {
					coveredAttrs[a.GetAttribute()] = true
				}
			}
		}
		var covered, uncovered []ConditionComponent
		for _, comp := range components {
			if isComponentCovered(comp, coveredRelations, coveredAttrs) {
				covered = append(covered, comp)
			} else {
				uncovered = append(uncovered, comp)
			}
		}
		coveragePercent := 100
		if len(components) > 0 {
			coveragePercent = (len(covered) * 100) / len(components)
		}
		result[permName] = &ConditionCoverageInfo{
			PermissionName:      permName,
			AllComponents:       components,
			CoveredComponents:   covered,
			UncoveredComponents: uncovered,
			CoveragePercent:     coveragePercent,
		}
	}
	return result
}

func extractAssertedPermissions(entityName string, scenario file.Scenario) map[string]bool {
	asserted := make(map[string]bool)
	for _, check := range scenario.Checks {
		entity, err := tuple.E(check.Entity)
		if err != nil || entity.GetType() != entityName {
			continue
		}
		for permName := range check.Assertions {
			asserted[permName] = true
		}
	}
	for _, filter := range scenario.EntityFilters {
		if filter.EntityType != entityName {
			continue
		}
		for permName := range filter.Assertions {
			asserted[permName] = true
		}
	}
	return asserted
}

func extractConditionComponents(child *base.Child) []ConditionComponent {
	if child == nil {
		return nil
	}
	if leaf := child.GetLeaf(); leaf != nil {
		comp := leafToComponent(leaf)
		if comp.Name != "" {
			return []ConditionComponent{comp}
		}
		return nil
	}
	if rewrite := child.GetRewrite(); rewrite != nil {
		var components []ConditionComponent
		for _, ch := range rewrite.GetChildren() {
			components = append(components, extractConditionComponents(ch)...)
		}
		return components
	}
	return nil
}

func leafToComponent(leaf *base.Leaf) ConditionComponent {
	if cus := leaf.GetComputedUserSet(); cus != nil {
		return ConditionComponent{Name: cus.GetRelation(), Type: "relation"}
	}
	if ttus := leaf.GetTupleToUserSet(); ttus != nil {
		tupleRel := ""
		if ts := ttus.GetTupleSet(); ts != nil {
			tupleRel = ts.GetRelation()
		}
		computedRel := ""
		if c := ttus.GetComputed(); c != nil {
			computedRel = c.GetRelation()
		}
		return ConditionComponent{Name: fmt.Sprintf("%s.%s", tupleRel, computedRel), Type: "tuple_to_userset"}
	}
	if ca := leaf.GetComputedAttribute(); ca != nil {
		return ConditionComponent{Name: ca.GetName(), Type: "attribute"}
	}
	if call := leaf.GetCall(); call != nil {
		return ConditionComponent{Name: fmt.Sprintf("call:%s", call.GetRuleName()), Type: "call"}
	}
	return ConditionComponent{}
}

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

func buildCoveredAttributeSet(entityName string, attrs []string) map[string]bool {
	covered := make(map[string]bool)
	for _, attrStr := range attrs {
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

func isComponentCovered(comp ConditionComponent, coveredRelations, coveredAttrs map[string]bool) bool {
	switch comp.Type {
	case "relation":
		return coveredRelations[comp.Name]
	case "tuple_to_userset":
		for i, ch := range comp.Name {
			if ch == '.' {
				return coveredRelations[comp.Name[:i]]
			}
		}
		return false
	case "attribute":
		return coveredAttrs[comp.Name]
	case "call":
		return true
	default:
		return false
	}
}
