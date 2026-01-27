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

	db, _ := factories.DatabaseFactory(config.Database{Engine: "memory"})
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

	checkEngine := NewCheckEngine(sr, dr)
	checkEngine.SetRegistry(registry)

	invoker := invoke.NewDirectInvoker(sr, dr, checkEngine, nil, nil, nil)
	checkEngine.SetInvoker(invoker)

	// Add owner
	tup, _ := tuple.Tuple("repository:1#owner@user:1")
	dw.Write(context.Background(), "t1", database.NewTupleCollection(tup), database.NewAttributeCollection())

	// Check repository:1#edit@user:1
	// This should match 'owner' (short-circuit OR)
	entity, _ := tuple.E("repository:1")
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
		if node.Path == "repository#edit.1" { // .1 is 'admin' in 'owner or admin'
			foundAdmin = true
		}
	}

	if !foundAdmin {
		t.Errorf("expected repository#edit.1 (admin) to be uncovered, but it wasn't in the report")
	}
}
