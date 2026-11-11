package engines

import (
	"context"
	"reflect"
	"testing"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

// TestLookupEntityExpandsRecursivePathChainCandidates verifies that recursive
// descendants are returned once across consecutive lookup pages.
func TestLookupEntityExpandsRecursivePathChainCandidates(t *testing.T) {
	schema := `
	entity user {}

	entity role {
		attribute allow_view boolean
		permission view = allow_view
	}

	entity role_assignment {
		relation assignee @user
		relation role @role
		permission view = assignee and role.view
	}

	entity retailer {
		relation grants @role_assignment
		permission view = grants.view
	}

	entity retailer_entity_group {
		relation parent @retailer @retailer_entity_group
		relation grants @role_assignment
		permission view = grants.view or parent.view
	}
	`

	db, err := factories.DatabaseFactory(config.Database{Engine: "memory"})
	if err != nil {
		t.Fatal(err)
	}
	conf, err := newSchema(schema)
	if err != nil {
		t.Fatal(err)
	}
	if err = factories.SchemaWriterFactory(db).WriteSchema(context.Background(), conf); err != nil {
		t.Fatal(err)
	}

	relationships := []string{
		// The lookup subject inherits view from the retailer root.
		"role_assignment:target#assignee@user:target",
		"role_assignment:target#role@role:viewer",
		"retailer:root#grants@role_assignment:target",
		// An unrelated subject has the same role directly on an intermediate group.
		"role_assignment:unrelated#assignee@user:unrelated",
		"role_assignment:unrelated#role@role:viewer",
		"retailer_entity_group:2#grants@role_assignment:unrelated",
		"retailer_entity_group:1#parent@retailer:root",
		"retailer_entity_group:2#parent@retailer_entity_group:1",
		"retailer_entity_group:3#parent@retailer_entity_group:2",
		"retailer_entity_group:4#parent@retailer_entity_group:3",
	}
	tuples := make([]*base.Tuple, 0, len(relationships))
	for _, relationship := range relationships {
		parsed, parseErr := tuple.Tuple(relationship)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		tuples = append(tuples, parsed)
	}
	roleAttribute, err := attribute.Attribute("role:viewer$allow_view|boolean:true")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = factories.DataWriterFactory(db).Write(
		context.Background(),
		"t1",
		database.NewTupleCollection(tuples...),
		database.NewAttributeCollection(roleAttribute),
	); err != nil {
		t.Fatal(err)
	}

	schemaReader := factories.SchemaReaderFactory(db)
	dataReader := factories.DataReaderFactory(db)
	checkEngine := NewCheckEngine(schemaReader, dataReader)
	lookupEngine := NewLookupEngine(checkEngine, schemaReader, dataReader)
	invoker := invoke.NewDirectInvoker(schemaReader, dataReader, checkEngine, nil, lookupEngine, nil)
	checkEngine.SetInvoker(invoker)

	lookup := func(continuousToken string) *base.PermissionLookupEntityResponse {
		response, lookupErr := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
			TenantId:        "t1",
			EntityType:      "retailer_entity_group",
			Subject:         &base.Subject{Type: "user", Id: "target"},
			Permission:      "view",
			ContinuousToken: continuousToken,
			Metadata: &base.PermissionLookupEntityRequestMetadata{
				SnapToken: token.NewNoopToken().Encode().String(),
				Depth:     20,
			},
			PageSize: 2,
		})
		if lookupErr != nil {
			t.Fatal(lookupErr)
		}
		return response
	}

	firstPage := lookup("")
	if want := []string{"1", "2"}; !reflect.DeepEqual(firstPage.GetEntityIds(), want) {
		t.Fatalf("got first page %v; want %v", firstPage.GetEntityIds(), want)
	}
	if firstPage.GetContinuousToken() == "" {
		t.Fatal("first page returned an empty continuous token")
	}

	secondPage := lookup(firstPage.GetContinuousToken())
	if want := []string{"3", "4"}; !reflect.DeepEqual(secondPage.GetEntityIds(), want) {
		t.Fatalf("got second page %v; want %v", secondPage.GetEntityIds(), want)
	}
	if secondPage.GetContinuousToken() != "" {
		t.Fatalf("second page returned unexpected continuous token %q", secondPage.GetContinuousToken())
	}
}
