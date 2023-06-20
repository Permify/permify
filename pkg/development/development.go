package development

import (
	"context"
	"fmt"
	"sort"

	"github.com/gookit/color"
	"github.com/rs/xid"
	"gopkg.in/yaml.v3"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/engines"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/servers"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/validation"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/development/file"
	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/telemetry"
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
		telemetry.NewNoopMeter(),
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
			storage.NewNoopWatcher(),
		),
	}
}

// Progresses - progress list
type Progresses struct {
	Pass       bool       `json:"pass"`
	ErrorCount int        `json:"error_count"`
	Messages   []Progress `json:"messages"`
}

// Progress - progress
type Progress struct {
	IsError bool   `json:"is_error"`
	Message string `json:"message"`
}

// AddError - add error to error list
func (l *Progresses) AddError(message string) {
	l.Pass = false
	l.ErrorCount++
	l.Messages = append(l.Messages, Progress{
		IsError: true,
		Message: message,
	})
}

// AddProgress - add progress to progress list
func (l *Progresses) AddProgress(message string) {
	l.Messages = append(l.Messages, Progress{
		IsError: false,
		Message: message,
	})
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
		relationships = append(relationships, tup)
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
	_, err = compiler.NewCompiler(true, sch).Compile()
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

// Validate performs validation and compilation of schema definitions. It takes a context and a shape map
// and returns a Progresses object which includes the results of each validation check.
func (c *Development) Validate(ctx context.Context, shape map[string]interface{}) *Progresses {
	// Initial setup of the Progresses object
	list := &Progresses{
		Pass:       true,
		ErrorCount: 0,
		Messages:   make([]Progress, 0),
	}

	// Marshal the shape map into YAML format
	out, err := yaml.Marshal(shape)
	if err != nil {
		list.AddError(err.Error())
		return list
	}

	// Unmarshal the YAML data into a file.Shape object
	s := &file.Shape{}
	err = yaml.Unmarshal(out, &s)
	if err != nil {
		list.AddError(err.Error())
		return list
	}

	// Start parsing the schema
	list.AddProgress("schema is creating... 🚀")

	// Parse the schema using the parser library
	sch, err := parser.NewParser(s.Schema).Parse()
	if err != nil {
		list.AddError(err.Error())
		return list
	}

	// Compile the parsed schema
	_, err = compiler.NewCompiler(true, sch).Compile()
	if err != nil {
		list.AddError(err.Error())
		return list
	}

	// Generate a new unique ID for this version of the schema
	version := xid.New().String()

	// Create a slice of SchemaDefinitions, one for each statement in the schema
	cnf := make([]storage.SchemaDefinition, 0, len(sch.Statements))
	for _, st := range sch.Statements {
		cnf = append(cnf, storage.SchemaDefinition{
			TenantID:             "t1",
			Version:              version,
			EntityType:           st.(*ast.EntityStatement).Name.Literal,
			SerializedDefinition: []byte(st.String()),
		})
	}

	// Write the schema definitions into the storage
	err = c.Container.SW.WriteSchema(ctx, cnf)
	if err != nil {
		list.AddError(err.Error())
		return list
	}

	// Indicate the schema was created successfully
	list.AddProgress("schema successfully created")

	// Start the process of creating relationships
	list.AddProgress("relationships are creating... 🚀")

	// Each item in the Relationships slice is processed individually
	for _, t := range s.Relationships {
		tup, err := tuple.Tuple(t)
		if err != nil {
			list.AddError(err.Error())
			continue
		}

		// Read the schema definition for this relationship
		definition, _, err := c.Container.SR.ReadSchemaDefinition(ctx, "t1", tup.GetEntity().GetType(), version)
		if err != nil {
			list.AddError(err.Error())
			return list
		}

		// Validate the relationship tuple against the schema definition
		err = validation.ValidateTuple(definition, tup)
		if err != nil {
			list.AddError(err.Error())
			return list
		}

		// Write the relationship to the database
		_, err = c.Container.RW.WriteRelationships(ctx, "t1", database.NewTupleCollection(tup))
		// Continue to the next relationship if an error occurred
		if err != nil {
			list.AddError(fmt.Sprintf("%s failed %s", t, err.Error()))
			continue
		}
	}

	list.AddProgress("checking scenarios... 🚀")

	// Each item in the Scenarios slice is processed individually
	for sn, scenario := range s.Scenarios {
		list.AddProgress(fmt.Sprintf("%v.scenario: %s - %s", sn+1, scenario.Name, scenario.Description))
		list.AddProgress("checks:")

		// Each Check in the current scenario is processed
		for _, check := range scenario.Checks {
			entity, err := tuple.E(check.Entity)
			if err != nil {
				list.AddError(err.Error())
				continue
			}

			ear, err := tuple.EAR(check.Subject)
			if err != nil {
				list.AddError(err.Error())
				continue
			}

			subject := &v1.Subject{
				Type:     ear.GetEntity().GetType(),
				Id:       ear.GetEntity().GetId(),
				Relation: ear.GetRelation(),
			}

			// Each Assertion in the current check is processed
			for permission, expected := range check.Assertions {
				exp := v1.PermissionCheckResponse_RESULT_ALLOWED
				if !expected {
					exp = v1.PermissionCheckResponse_RESULT_DENIED
				}

				// A Permission Check is made for the current entity, permission and subject
				res, err := c.Container.Invoker.Check(ctx, &v1.PermissionCheckRequest{
					TenantId: "t1",
					Metadata: &v1.PermissionCheckRequestMetadata{
						SchemaVersion: version,
						SnapToken:     token.NewNoopToken().Encode().String(),
						Depth:         100,
					},
					Entity:     entity,
					Permission: permission,
					Subject:    subject,
				})
				if err != nil {
					list.AddError(err.Error())
					continue
				}

				query := tuple.SubjectToString(subject) + " " + permission + " " + tuple.EntityToString(entity)

				// Check if the permission check result matches the expected result
				if res.Can == exp {
					list.AddProgress(fmt.Sprintf("success: %s", query))
				} else {
					list.AddError(fmt.Sprintf("fail: %s ->", query))

					// Handle the case where the permission check result is ALLOWED but the expected result was DENIED
					if res.Can == v1.PermissionCheckResponse_RESULT_ALLOWED {
						list.AddError(fmt.Sprintf("fail: %s -> expected: DENIED actual: ALLOWED ", query))
					} else {
						// Handle the case where the permission check result is DENIED but the expected result was ALLOWED
						list.AddError(fmt.Sprintf("fail: %s -> expected: ALLOWED actual: DENIED ", query))
					}
				}
			}
		}

		list.AddProgress("entity_filters:")

		// Each EntityFilter in the current scenario is processed
		for _, filter := range scenario.EntityFilters {
			ear, err := tuple.EAR(filter.Subject)
			if err != nil {
				list.AddError(err.Error())
				continue
			}

			subject := &v1.Subject{
				Type:     ear.GetEntity().GetType(),
				Id:       ear.GetEntity().GetId(),
				Relation: ear.GetRelation(),
			}

			// Each Assertion in the current filter is processed

			for permission, expected := range filter.Assertions {
				// Perform a lookup for the entity with the given subject and permission
				res, err := c.Container.Invoker.LookupEntity(ctx, &v1.PermissionLookupEntityRequest{
					TenantId: "t1",
					Metadata: &v1.PermissionLookupEntityRequestMetadata{
						SchemaVersion: version,
						SnapToken:     token.NewNoopToken().Encode().String(),
						Depth:         100,
					},
					EntityType: filter.EntityType,
					Permission: permission,
					Subject:    subject,
				})
				if err != nil {
					list.AddError(err.Error())
					continue
				}

				query := tuple.SubjectToString(subject) + " " + permission + " " + filter.EntityType

				// Check if the actual result of the entity lookup matches the expected result
				if isSameArray(res.GetEntityIds(), expected) {
					list.AddProgress(fmt.Sprintf("success: %v", query))
				} else {
					list.AddError(fmt.Sprintf("fail: %s -> expected: %+v actual: %+v", query, expected, res.GetEntityIds()))
				}
			}
		}

		color.Notice.Println("subject_filters:")

		// Each SubjectFilter in the current scenario is processed
		for _, filter := range scenario.SubjectFilters {

			subjectReference := tuple.RelationReference(filter.SubjectReference)
			if err != nil {
				list.AddError(err.Error())
				continue
			}

			var entity *v1.Entity
			entity, err = tuple.E(filter.Entity)
			if err != nil {
				list.AddError(err.Error())
				continue
			}

			// Each Assertion in the current filter is processed
			for permission, expected := range filter.Assertions {
				// Perform a lookup for the subject with the given entity and permission
				res, err := c.Container.Invoker.LookupSubject(ctx, &v1.PermissionLookupSubjectRequest{
					TenantId: "t1",
					Metadata: &v1.PermissionLookupSubjectRequestMetadata{
						SchemaVersion: version,
						SnapToken:     token.NewNoopToken().Encode().String(),
					},
					SubjectReference: subjectReference,
					Permission:       permission,
					Entity:           entity,
				})
				if err != nil {
					list.AddError(err.Error())
					continue
				}

				query := tuple.EntityToString(entity) + " " + permission + " " + filter.SubjectReference

				// Check if the actual result of the subject lookup matches the expected result
				if isSameArray(res.GetSubjectIds(), expected) {
					list.AddProgress(fmt.Sprintf("success: %v", query))
				} else {
					list.AddError(fmt.Sprintf("fail: %s -> expected: %+v actual: %+v", query, expected, res.GetSubjectIds()))
				}
			}
		}
	}

	// Return the results of all checks and validations
	return list
}

// isSameArray - check if two arrays are the same
func isSameArray(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sortedA := make([]string, len(a))
	copy(sortedA, a)
	sort.Strings(sortedA)

	sortedB := make([]string, len(b))
	copy(sortedB, b)
	sort.Strings(sortedB)

	for i := range sortedA {
		if sortedA[i] != sortedB[i] {
			return false
		}
	}

	return true
}
