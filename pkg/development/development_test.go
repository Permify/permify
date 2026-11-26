package development

import (
	"context"
	"strings"
	"testing"

	"github.com/Permify/permify/pkg/development/file"
)

// TestScenarioScopedRelationshipsAndAttributes exercises the Scenario.Relationships
// and Scenario.Attributes fields end-to-end through RunWithShape. The scenario's
// relationships and attributes must be treated as contextual data: they are
// scoped to the scenario's checks and must not leak into later scenarios.
func TestScenarioScopedRelationshipsAndAttributes(t *testing.T) {
	schema := `
entity user {}

entity account {
    relation owner @user
    attribute balance integer

    permission withdraw = check_balance(balance) and owner
}

rule check_balance(balance integer) {
    (balance >= context.data.amount) && (context.data.amount <= 5000)
}
`

	shape := &file.Shape{
		Schema:        schema,
		Relationships: []string{},
		Attributes:    []string{},
		Scenarios: []file.Scenario{
			{
				Name: "scenario-scoped owner can withdraw",
				Relationships: []string{
					"account:1#owner@user:andrew",
				},
				Attributes: []string{
					"account:1$balance|integer:2000",
				},
				Checks: []file.Check{
					{
						Entity:  "account:1",
						Subject: "user:andrew",
						Context: file.Context{
							Data: map[string]interface{}{"amount": 1500},
						},
						Assertions: map[string]bool{"withdraw": true},
					},
				},
			},
			{
				Name: "scenario-scoped data does not leak across scenarios",
				// No scenario-scoped fixtures here. user:andrew should no longer be
				// an owner of account:1 and the balance attribute should not apply.
				Checks: []file.Check{
					{
						Entity:  "account:1",
						Subject: "user:andrew",
						Context: file.Context{
							Data: map[string]interface{}{"amount": 1500},
						},
						Assertions: map[string]bool{"withdraw": false},
					},
				},
				// Explicit negative lookups: if the previous scenario's owner tuple
				// leaked, account:1 would resolve for user:andrew and user:andrew
				// would resolve as a viewer of account:1. Both must be empty.
				EntityFilters: []file.EntityFilter{
					{
						EntityType: "account",
						Subject:    "user:andrew",
						Context: file.Context{
							Data: map[string]interface{}{"amount": 1500},
						},
						Assertions: map[string][]string{
							"withdraw": {},
						},
					},
				},
				SubjectFilters: []file.SubjectFilter{
					{
						SubjectReference: "user",
						Entity:           "account:1",
						Context: file.Context{
							Data: map[string]interface{}{"amount": 1500},
						},
						Assertions: map[string][]string{
							"withdraw": {},
						},
					},
				},
			},
		},
	}

	dev := NewContainer()
	errs := dev.RunWithShape(context.Background(), shape)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %+v", errs)
	}
}

// TestScenarioScopedFixturesApplyToFilters verifies that scenario-scoped
// relationships and attributes are merged into the context of both entity
// filters and subject filters, not just checks.
func TestScenarioScopedFixturesApplyToFilters(t *testing.T) {
	schema := `
entity user {}

entity account {
    relation owner @user
    attribute active boolean

    permission view = is_active(active) and owner
}

rule is_active(active boolean) {
    active == true
}
`

	shape := &file.Shape{
		Schema: schema,
		Scenarios: []file.Scenario{
			{
				Name: "filters see scenario-scoped fixtures",
				Relationships: []string{
					"account:1#owner@user:andrew",
					"account:2#owner@user:andrew",
				},
				Attributes: []string{
					"account:1$active|boolean:true",
					"account:2$active|boolean:true",
				},
				EntityFilters: []file.EntityFilter{
					{
						EntityType: "account",
						Subject:    "user:andrew",
						Assertions: map[string][]string{
							"view": {"1", "2"},
						},
					},
				},
				SubjectFilters: []file.SubjectFilter{
					{
						SubjectReference: "user",
						Entity:           "account:1",
						Assertions: map[string][]string{
							"view": {"andrew"},
						},
					},
				},
			},
		},
	}

	dev := NewContainer()
	errs := dev.RunWithShape(context.Background(), shape)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %+v", errs)
	}
}

// TestScenarioRelationshipsReportInvalidTuple verifies that a malformed
// scenario-scoped relationship is surfaced as a scenarios-scoped error rather
// than crashing the run.
func TestScenarioRelationshipsReportInvalidTuple(t *testing.T) {
	shape := &file.Shape{
		Schema: `
entity user {}
entity account {
    relation owner @user
    permission view = owner
}
`,
		Scenarios: []file.Scenario{
			{
				Name: "bad tuple",
				Relationships: []string{
					"not-a-valid-tuple",
				},
				Checks: []file.Check{
					{
						Entity:     "account:1",
						Subject:    "user:andrew",
						Assertions: map[string]bool{"view": false},
					},
				},
			},
		},
	}

	dev := NewContainer()
	errs := dev.RunWithShape(context.Background(), shape)
	if len(errs) == 0 {
		t.Fatalf("expected at least one error for invalid scenario relationship")
	}
	found := false
	for _, e := range errs {
		if e.Type == "scenarios" && e.Key == 0 && strings.Contains(e.Message, "invalid tuple") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a scenarios error at key 0 from the tuple parser, got %+v", errs)
	}
}

// TestScenarioAttributesReportInvalidAttribute is the attribute-path mirror of
// TestScenarioRelationshipsReportInvalidTuple: a malformed scenario-scoped
// attribute must be reported as a scenarios error instead of crashing.
func TestScenarioAttributesReportInvalidAttribute(t *testing.T) {
	shape := &file.Shape{
		Schema: `
entity user {}
entity account {
    attribute active boolean
    permission view = is_active(active)
}

rule is_active(active boolean) {
    active == true
}
`,
		Scenarios: []file.Scenario{
			{
				Name: "bad attribute",
				Attributes: []string{
					"not-a-valid-attribute",
				},
				Checks: []file.Check{
					{
						Entity:     "account:1",
						Subject:    "user:andrew",
						Assertions: map[string]bool{"view": false},
					},
				},
			},
		},
	}

	dev := NewContainer()
	errs := dev.RunWithShape(context.Background(), shape)
	if len(errs) == 0 {
		t.Fatalf("expected at least one error for invalid scenario attribute")
	}
	found := false
	for _, e := range errs {
		if e.Type == "scenarios" && e.Key == 0 && strings.Contains(e.Message, "invalid attribute") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a scenarios error at key 0 from the attribute parser, got %+v", errs)
	}
}
