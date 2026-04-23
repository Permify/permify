package development

import (
	"context"
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
		if e.Type == "scenarios" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected an error with Type=scenarios, got %+v", errs)
	}
}
