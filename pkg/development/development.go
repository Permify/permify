package development

import (
	"context"
	"fmt"

	"github.com/rs/xid"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/engines"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/servers"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/validation"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

type Development struct {
	Container *servers.Container
}

func NewContainer() *Development {
	var err error

	// Create a new in-memory database using the factories package
	var db database.Database
	db, err = factories.DatabaseFactory(config.Database{Engine: database.MEMORY.String()})
	if err != nil {
		fmt.Println(err)
	}

	// Create a new logger instance
	l := logger.New("debug")

	// Create instances of storage using the factories package
	relationshipReader := factories.RelationshipReaderFactory(db, l)
	relationshipWriter := factories.RelationshipWriterFactory(db, l)
	schemaReader := factories.SchemaReaderFactory(db, l)
	schemaWriter := factories.SchemaWriterFactory(db, l)
	tenantReader := factories.TenantReaderFactory(db, l)
	tenantWriter := factories.TenantWriterFactory(db, l)

	// Create instances of engines
	checkEngine := engines.NewCheckEngine(schemaReader, relationshipReader)
	expandEngine := engines.NewExpandEngine(schemaReader, relationshipReader)
	entityFilterEngine := engines.NewEntityFilterEngine(schemaReader, relationshipReader)
	lookupEntityEngine := engines.NewLookupEntityEngine(checkEngine, entityFilterEngine)
	lookupSubjectEngine := engines.NewLookupSubjectEngine(schemaReader, relationshipReader)

	invoker := invoke.NewDirectInvoker(
		schemaReader,
		relationshipReader,
		checkEngine,
		expandEngine,
		lookupEntityEngine,
		lookupSubjectEngine,
	)

	checkEngine.SetInvoker(invoker)

	// Create a new container instance with engines, storage, and other dependencies
	return &Development{
		Container: servers.NewContainer(
			invoker,
			relationshipReader,
			relationshipWriter,
			schemaReader,
			schemaWriter,
			tenantReader,
			tenantWriter,
		),
	}
}

// Check - Creates new permission check request
func (c *Development) Check(ctx context.Context, subject *v1.Subject, action string, entity *v1.Entity) (*v1.PermissionCheckResponse, error) {
	// Create a new permission check request with the given subject, action, entity, and metadata
	req := &v1.PermissionCheckRequest{
		TenantId:   "t1",
		Entity:     entity,
		Subject:    subject,
		Permission: action,
		Metadata: &v1.PermissionCheckRequestMetadata{
			SchemaVersion: "",
			SnapToken:     "",
			Depth:         20,
		},
	}

	// Invoke the permission check using the container's invoker and return the response
	return c.Container.Invoker.Check(ctx, req)
}

// LookupEntity - Looks up an entity's permissions for a given subject and permission
func (c *Development) LookupEntity(ctx context.Context, subject *v1.Subject, permission, entityType string) (res *v1.PermissionLookupEntityResponse, err error) {
	// Create a new permission lookup entity request with the given subject, permission, entity type, and metadata
	req := &v1.PermissionLookupEntityRequest{
		TenantId:   "t1",
		EntityType: entityType,
		Subject:    subject,
		Permission: permission,
		Metadata: &v1.PermissionLookupEntityRequestMetadata{
			SchemaVersion: "",
			SnapToken:     "",
			Depth:         20,
		},
	}

	// Invoke the permission lookup entity using the container's invoker and return the response
	return c.Container.Invoker.LookupEntity(ctx, req)
}

// LookupSubject - Looks up a subject's permissions for a given entıty and permission
func (c *Development) LookupSubject(ctx context.Context, entity *v1.Entity, permission string, subjectReference *v1.RelationReference) (res *v1.PermissionLookupSubjectResponse, err error) {
	// Create a new permission lookup entity request with the given subject, permission, entity type, and metadata
	req := &v1.PermissionLookupSubjectRequest{
		TenantId:         "t1",
		Entity:           entity,
		Permission:       permission,
		SubjectReference: subjectReference,
		Metadata: &v1.PermissionLookupSubjectRequestMetadata{
			SchemaVersion: "",
			SnapToken:     "",
		},
	}

	// Invoke the permission lookup entity using the container's invoker and return the response
	return c.Container.Invoker.LookupSubject(ctx, req)
}

// ReadTuple - Creates new read API request
func (c *Development) ReadTuple(ctx context.Context, filter *v1.TupleFilter) (tuples *database.TupleCollection, continuousToken database.EncodedContinuousToken, err error) {
	// Get the head snapshot of the "t1" schema from the schema repository
	snap, err := c.Container.RR.HeadSnapshot(ctx, "t1")
	if err != nil {
		return nil, nil, err
	}

	// Create a new read relationships request with the given tuple filter, snapshot token, and pagination
	return c.Container.RR.ReadRelationships(ctx, "t1", filter, snap.Encode().String(), database.NewPagination())
}

// WriteTuple - Creates new write API request
func (c *Development) WriteTuple(ctx context.Context, tuples []*v1.Tuple) (err error) {
	// Get the head version of the "t1" schema from the schema repository
	version, err := c.Container.SR.HeadVersion(ctx, "t1")
	if err != nil {
		return err
	}

	// Create a new slice to hold the validated tuples
	relationships := make([]*v1.Tuple, 0, len(tuples))

	// Validate each tuple and append it to the relationships slice
	for _, tup := range tuples {

		// Set the subject relation to ellipsis if the subject is not a user and there is no relation
		subject := tuple.SetSubjectRelationToEllipsisIfNonUserAndNoRelation(tup.GetSubject())

		// Read the schema definition for the tuple's entity type and version from the schema repository
		definition, _, err := c.Container.SR.ReadSchemaDefinition(ctx, "t1", tup.GetEntity().GetType(), version)
		if err != nil {
			return err
		}

		// Validate the tuple against the schema definition
		err = validation.ValidateTuple(definition, tup)
		if err != nil {
			return err
		}

		// Append the validated tuple to the relationships slice
		relationships = append(relationships, &v1.Tuple{
			Entity:   tup.GetEntity(),
			Relation: tup.GetRelation(),
			Subject:  subject,
		})
	}

	// Write the relationships to the relationship repository and return the encoded snap token
	_, err = c.Container.RW.WriteRelationships(ctx, "t1", database.NewTupleCollection(relationships...))
	return err
}

// DeleteTuple - Creates new delete relation tuple request
func (c *Development) DeleteTuple(ctx context.Context, filter *v1.TupleFilter) (token token.EncodedSnapToken, err error) {
	return c.Container.RW.DeleteRelationships(ctx, "t1", filter)
}

// WriteSchema - Creates new write schema request
func (c *Development) WriteSchema(ctx context.Context, schema string) (err error) {
	// Parse the schema string into an abstract syntax tree (AST)
	sch, err := parser.NewParser(schema).Parse()
	if err != nil {
		return err
	}

	// Compile the AST into a set of schema definitions
	_, err = compiler.NewCompiler(false, sch).Compile()
	if err != nil {
		return err
	}

	// Generate a new version ID for the schema definition
	version := xid.New().String()

	// Create a new slice to hold the schema definitions
	cnf := make([]storage.SchemaDefinition, 0, len(sch.Statements))

	// Convert each statement in the AST into a schema definition and append it to the cnf slice
	for _, st := range sch.Statements {
		cnf = append(cnf, storage.SchemaDefinition{
			TenantID:             "t1",
			Version:              version,
			EntityType:           st.(*ast.EntityStatement).Name.Literal,
			SerializedDefinition: []byte(st.String()),
		})
	}

	// Write the schema definitions to the schema repository
	return c.Container.SW.WriteSchema(ctx, cnf)
}

// ReadSchema - Creates new read schema request
func (c *Development) ReadSchema(ctx context.Context) (sch *v1.SchemaDefinition, err error) {
	// Get the head version of the "t1" schema from the schema repository
	version, err := c.Container.SR.HeadVersion(ctx, "t1")
	if err != nil {
		return nil, err
	}

	// Read the schema definition for the given schema and version from the schema repository
	return c.Container.SR.ReadSchema(ctx, "t1", version)
}
