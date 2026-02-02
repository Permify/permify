package engines

import (
	"context"
	"testing"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/coverage"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

func TestCheckEngineCoverage(t *testing.T) {
	schema := `
		entity user {}
		entity repository {
			relation owner @user
			relation admin @user
			permission edit = owner or admin
		}
	`

	p := parser.NewParser(schema)
	sch, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	c := compiler.NewCompiler(true, sch)
	entities, _, err := c.Compile()
	if err != nil {
		t.Fatal(err)
	}

	db, err := factories.DatabaseFactory(config.Database{Engine: "memory"})
	if err != nil {
		t.Fatal(err)
	}
	sw := factories.SchemaWriterFactory(db)

	for _, entity := range entities {
		err := sw.WriteSchema(context.Background(), []storage.SchemaDefinition{
			{
				TenantID:             "t1",
				Name:                 entity.Name,
				SerializedDefinition: []byte(schema),
				Version:              "v1",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	sr := factories.SchemaReaderFactory(db)
	dr := factories.DataReaderFactory(db)
	dw := factories.DataWriterFactory(db)

	registry := coverage.NewRegistry()
	coverage.Discover(sch, registry)

	// Concurrency limit 1 enables sequential execution and short-circuit detection.
	checkEngine := NewCheckEngine(sr, dr, CheckConcurrencyLimit(1))
	checkEngine.SetRegistry(registry)

	invoker := invoke.NewDirectInvoker(sr, dr, checkEngine, nil, nil, nil)
	checkEngine.SetInvoker(invoker)

	// Add owner. For OR, we check owner first - it succeeds. Short-circuit: admin never runs.
	tup, err := tuple.Tuple("repository:1#owner@user:1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dw.Write(context.Background(), "t1", database.NewTupleCollection(tup), database.NewAttributeCollection()); err != nil {
		t.Fatal(err)
	}

	// Check repository:1#edit@user:1 - owner matches (short-circuit), admin never evaluated.
	entity, err := tuple.E("repository:1")
	if err != nil {
		t.Fatal(err)
	}
	subject := &base.Subject{Type: "user", Id: "1"}

	_, err = invoker.Check(context.Background(), &base.PermissionCheckRequest{
		TenantId:   "t1",
		Entity:     entity,
		Subject:    subject,
		Permission: "edit",
		Metadata: &base.PermissionCheckRequestMetadata{
			SnapToken: token.NewNoopToken().Encode().String(),
			Depth:     20,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	report := registry.Report()

	// 'admin' should be uncovered because of short-circuit (owner was true)
	foundAdmin := false
	for _, node := range report {
		if node.Path == "repository#edit.op.1.leaf" { // .op.1.leaf is 'admin' in 'owner or admin'
			foundAdmin = true
		}
	}

	if !foundAdmin {
		t.Errorf("expected repository#edit.1 (admin) to be uncovered, but it wasn't in the report")
	}
}
