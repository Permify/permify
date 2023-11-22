package development

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"

	"gopkg.in/yaml.v3"

	"github.com/rs/xid"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/engines"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/servers"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/validation"
	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/development/file"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Create instances of storage using the factories package
	dataReader := factories.DataReaderFactory(db)
	dataWriter := factories.DataWriterFactory(db)
	schemaReader := factories.SchemaReaderFactory(db)
	schemaWriter := factories.SchemaWriterFactory(db)
	tenantReader := factories.TenantReaderFactory(db)
	tenantWriter := factories.TenantWriterFactory(db)

	// Create instances of engines
	checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
	expandEngine := engines.NewExpandEngine(schemaReader, dataReader)
	lookupEngine := engines.NewLookupEngine(checkEngine, schemaReader, dataReader)
	subjectPermissionEngine := engines.NewSubjectPermission(checkEngine, schemaReader)

	invoker := invoke.NewDirectInvoker(
		schemaReader,
		dataReader,
		checkEngine,
		expandEngine,
		lookupEngine,
		subjectPermissionEngine,
		telemetry.NewNoopMeter(),
	)

	checkEngine.SetInvoker(invoker)

	// Create a new container instance with engines, storage, and other dependencies
	return &Development{
		Container: servers.NewContainer(
			invoker,
			dataReader,
			dataWriter,
			schemaReader,
			schemaWriter,
			tenantReader,
			tenantWriter,
			storage.NewNoopWatcher(),
		),
	}
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

type Error struct {
	Type    string `json:"type"`
	Key     any    `json:"key"`
	Message string `json:"message"`
}

func (c *Development) Run(ctx context.Context, shape map[string]interface{}) (errors []Error) {
	// Marshal the shape map into YAML format
	out, err := yaml.Marshal(shape)
	if err != nil {
		errors = append(errors, Error{
			Type:    "file_validation",
			Key:     "",
			Message: err.Error(),
		})
		return
	}

	// Unmarshal the YAML data into a file.Shape object
	s := &file.Shape{}
	err = yaml.Unmarshal(out, &s)
	if err != nil {
		errors = append(errors, Error{
			Type:    "file_validation",
			Key:     "",
			Message: err.Error(),
		})
		return
	}

	// Parse the schema using the parser library
	sch, err := parser.NewParser(s.Schema).Parse()
	if err != nil {
		errors = append(errors, Error{
			Type:    "schema",
			Key:     "",
			Message: err.Error(),
		})
		return
	}

	// Compile the parsed schema
	_, _, err = compiler.NewCompiler(true, sch).Compile()
	if err != nil {
		errors = append(errors, Error{
			Type:    "schema",
			Key:     "",
			Message: err.Error(),
		})
		return
	}

	// Generate a new unique ID for this version of the schema
	version := xid.New().String()

	// Create a slice of SchemaDefinitions, one for each statement in the schema
	cnf := make([]storage.SchemaDefinition, 0, len(sch.Statements))
	for _, st := range sch.Statements {
		cnf = append(cnf, storage.SchemaDefinition{
			TenantID:             "t1",
			Version:              version,
			Name:                 st.GetName(),
			SerializedDefinition: []byte(st.String()),
		})
	}

	// Write the schema definitions into the storage
	err = c.Container.SW.WriteSchema(ctx, cnf)
	if err != nil {
		errors = append(errors, Error{
			Type:    "schema",
			Key:     "",
			Message: err.Error(),
		})
		return
	}

	// Each item in the Relationships slice is processed individually
	for _, t := range s.Relationships {
		tup, err := tuple.Tuple(t)
		if err != nil {
			errors = append(errors, Error{
				Type:    "relationships",
				Key:     t,
				Message: err.Error(),
			})
			continue
		}

		// Read the schema definition for this relationship
		definition, _, err := c.Container.SR.ReadEntityDefinition(ctx, "t1", tup.GetEntity().GetType(), version)
		if err != nil {
			errors = append(errors, Error{
				Type:    "relationships",
				Key:     t,
				Message: err.Error(),
			})
			continue
		}

		// Validate the relationship tuple against the schema definition
		err = validation.ValidateTuple(definition, tup)
		if err != nil {
			errors = append(errors, Error{
				Type:    "relationships",
				Key:     t,
				Message: err.Error(),
			})
			continue
		}

		// Write the relationship to the database
		_, err = c.Container.DW.Write(ctx, "t1", database.NewTupleCollection(tup), database.NewAttributeCollection())
		// Continue to the next relationship if an error occurred
		if err != nil {
			errors = append(errors, Error{
				Type:    "relationships",
				Key:     t,
				Message: err.Error(),
			})
			continue
		}
	}

	// Each item in the Attributes slice is processed individually
	for _, a := range s.Attributes {
		attr, err := attribute.Attribute(a)
		if err != nil {
			errors = append(errors, Error{
				Type:    "attributes",
				Key:     a,
				Message: err.Error(),
			})
			continue
		}

		// Read the schema definition for this attribute
		definition, _, err := c.Container.SR.ReadEntityDefinition(ctx, "t1", attr.GetEntity().GetType(), version)
		if err != nil {
			errors = append(errors, Error{
				Type:    "attributes",
				Key:     a,
				Message: err.Error(),
			})
			continue
		}

		// Validate the attribute against the schema definition
		err = validation.ValidateAttribute(definition, attr)
		if err != nil {
			errors = append(errors, Error{
				Type:    "attributes",
				Key:     a,
				Message: err.Error(),
			})
			continue
		}

		// Write the attribute to the database
		_, err = c.Container.DW.Write(ctx, "t1", database.NewTupleCollection(), database.NewAttributeCollection(attr))
		// Continue to the next attribute if an error occurred
		if err != nil {
			errors = append(errors, Error{
				Type:    "attributes",
				Key:     a,
				Message: err.Error(),
			})
			continue
		}
	}

	// Each item in the Scenarios slice is processed individually
	for i, scenario := range s.Scenarios {

		var createdScenarioRelationshipsTups []*v1.Tuple
		// Add scenario specific local relationships
		for _, t := range scenario.Relationships {
			tup, err := tuple.Tuple(t)
			if err != nil {
				errors = append(errors, Error{
					Type:    "relationships",
					Key:     t,
					Message: err.Error(),
				})
				continue
			}

			// Read the schema definition for this relationship
			definition, _, err := c.Container.SR.ReadEntityDefinition(ctx, "t1", tup.GetEntity().GetType(), version)
			if err != nil {
				errors = append(errors, Error{
					Type:    "relationships",
					Key:     t,
					Message: err.Error(),
				})
				continue
			}

			// Validate the relationship tuple against the schema definition
			err = validation.ValidateTuple(definition, tup)
			if err != nil {
				errors = append(errors, Error{
					Type:    "relationships",
					Key:     t,
					Message: err.Error(),
				})
				continue
			}

			// Write the relationship to the database
			_, err = c.Container.DW.Write(ctx, "t1", database.NewTupleCollection(tup), database.NewAttributeCollection())
			// Continue to the next relationship if an error occurred
			if err != nil {
				errors = append(errors, Error{
					Type:    "relationships",
					Key:     t,
					Message: err.Error(),
				})
				continue
			}
			createdScenarioRelationshipsTups = append(createdScenarioRelationshipsTups, tup)
		}

		// Add scenario specific attributes
		var createdScenarioAttributes []*v1.Attribute
		for _, a := range scenario.Attributes {
			attr, err := attribute.Attribute(a)
			if err != nil {
				errors = append(errors, Error{
					Type:    "attributes",
					Key:     a,
					Message: err.Error(),
				})
				continue
			}

			// Read the schema definition for this attribute
			definition, _, err := c.Container.SR.ReadEntityDefinition(ctx, "t1", attr.GetEntity().GetType(), version)
			if err != nil {
				errors = append(errors, Error{
					Type:    "attributes",
					Key:     a,
					Message: err.Error(),
				})
				continue
			}

			// Validate the attribute against the schema definition
			err = validation.ValidateAttribute(definition, attr)
			if err != nil {
				errors = append(errors, Error{
					Type:    "attributes",
					Key:     a,
					Message: err.Error(),
				})
				continue
			}

			// Write the attribute to the database
			_, err = c.Container.DW.Write(ctx, "t1", database.NewTupleCollection(), database.NewAttributeCollection(attr))
			// Continue to the next attribute if an error occurred
			if err != nil {
				errors = append(errors, Error{
					Type:    "attributes",
					Key:     a,
					Message: err.Error(),
				})
				continue
			}
			createdScenarioAttributes = append(createdScenarioAttributes, attr)
		}
		// Each Check in the current scenario is processed
		for _, check := range scenario.Checks {
			entity, err := tuple.E(check.Entity)
			if err != nil {
				errors = append(errors, Error{
					Type:    "scenarios",
					Key:     i,
					Message: err.Error(),
				})
				continue
			}

			ear, err := tuple.EAR(check.Subject)
			if err != nil {
				errors = append(errors, Error{
					Type:    "scenarios",
					Key:     i,
					Message: err.Error(),
				})
				continue
			}

			cont, err := Context(check.Context)
			if err != nil {
				errors = append(errors, Error{
					Type:    "scenarios",
					Key:     i,
					Message: err.Error(),
				})
				continue
			}

			subject := &v1.Subject{
				Type:     ear.GetEntity().GetType(),
				Id:       ear.GetEntity().GetId(),
				Relation: ear.GetRelation(),
			}

			// Each Assertion in the current check is processed
			for permission, expected := range check.Assertions {
				exp := v1.CheckResult_CHECK_RESULT_ALLOWED
				if !expected {
					exp = v1.CheckResult_CHECK_RESULT_DENIED
				}

				// A Permission Check is made for the current entity, permission and subject
				res, err := c.Container.Invoker.Check(ctx, &v1.PermissionCheckRequest{
					TenantId: "t1",
					Metadata: &v1.PermissionCheckRequestMetadata{
						SchemaVersion: version,
						SnapToken:     token.NewNoopToken().Encode().String(),
						Depth:         100,
					},
					Context:    cont,
					Entity:     entity,
					Permission: permission,
					Subject:    subject,
				})
				if err != nil {
					errors = append(errors, Error{
						Type:    "checks",
						Key:     i,
						Message: err.Error(),
					})
					continue
				}

				query := tuple.SubjectToString(subject) + " " + permission + " " + tuple.EntityToString(entity)

				// Check if the permission check result matches the expected result
				if res.Can != exp {
					var expectedStr, actualStr string
					if exp == v1.CheckResult_CHECK_RESULT_ALLOWED {
						expectedStr = "true"
					} else {
						expectedStr = "false"
					}

					if res.Can == v1.CheckResult_CHECK_RESULT_ALLOWED {
						actualStr = "true"
					} else {
						actualStr = "false"
					}

					// Construct a detailed error message with the expected result, actual result, and the query
					errorMsg := fmt.Sprintf("Query: %s, Expected: %s, Actual: %s", query, expectedStr, actualStr)

					errors = append(errors, Error{
						Type:    "scenarios",
						Key:     i,
						Message: errorMsg,
					})
				}
			}
		}

		// Each EntityFilter in the current scenario is processed
		for _, filter := range scenario.EntityFilters {
			ear, err := tuple.EAR(filter.Subject)
			if err != nil {
				errors = append(errors, Error{
					Type:    "scenarios",
					Key:     i,
					Message: err.Error(),
				})
				continue
			}

			cont, err := Context(filter.Context)
			if err != nil {
				errors = append(errors, Error{
					Type:    "scenarios",
					Key:     i,
					Message: err.Error(),
				})
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
					Context:    cont,
					EntityType: filter.EntityType,
					Permission: permission,
					Subject:    subject,
				})
				if err != nil {
					errors = append(errors, Error{
						Type:    "scenarios",
						Key:     i,
						Message: err.Error(),
					})
					continue
				}

				query := tuple.SubjectToString(subject) + " " + permission + " " + filter.EntityType

				// Check if the actual result of the entity lookup does NOT match the expected result
				if !isSameArray(res.GetEntityIds(), expected) {
					expectedStr := strings.Join(expected, ", ")
					actualStr := strings.Join(res.GetEntityIds(), ", ")

					errorMsg := fmt.Sprintf("Query: %s, Expected: [%s], Actual: [%s]", query, expectedStr, actualStr)

					errors = append(errors, Error{
						Type:    "scenarios",
						Key:     i,
						Message: errorMsg,
					})
				}
			}
		}

		// Each SubjectFilter in the current scenario is processed
		for _, filter := range scenario.SubjectFilters {

			subjectReference := tuple.RelationReference(filter.SubjectReference)
			if err != nil {
				errors = append(errors, Error{
					Type:    "scenarios",
					Key:     i,
					Message: err.Error(),
				})
				continue
			}

			cont, err := Context(filter.Context)
			if err != nil {
				errors = append(errors, Error{
					Type:    "scenarios",
					Key:     i,
					Message: err.Error(),
				})
				continue
			}

			var entity *v1.Entity
			entity, err = tuple.E(filter.Entity)
			if err != nil {
				errors = append(errors, Error{
					Type:    "scenarios",
					Key:     i,
					Message: err.Error(),
				})
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
						Depth:         100,
					},
					Context:          cont,
					SubjectReference: subjectReference,
					Permission:       permission,
					Entity:           entity,
				})
				if err != nil {
					errors = append(errors, Error{
						Type:    "scenarios",
						Key:     i,
						Message: err.Error(),
					})
					continue
				}

				query := tuple.EntityToString(entity) + " " + permission + " " + filter.SubjectReference

				// Check if the actual result of the subject lookup does NOT match the expected result
				if !isSameArray(res.GetSubjectIds(), expected) {
					expectedStr := strings.Join(expected, ", ")
					actualStr := strings.Join(res.GetSubjectIds(), ", ")

					errorMsg := fmt.Sprintf("Query: %s, Expected: [%s], Actual: [%s]", query, expectedStr, actualStr)

					errors = append(errors, Error{
						Type:    "scenarios",
						Key:     i,
						Message: errorMsg,
					})
				}
			}
		}

		// Once the scenarios are tested, delete created scenario specific relationships.
		for _, t := range createdScenarioRelationshipsTups {
			// Write the relationship to the database
			_, err = c.Container.DW.Delete(ctx, "t1", &v1.TupleFilter{
				Entity: &v1.EntityFilter{
					Type: t.GetEntity().Type,
					Ids:  []string{t.Entity.Id},
				},
				Relation: t.GetRelation(),
			},
				&v1.AttributeFilter{},
			)
			// Continue to the next relationship if an error occurred
			if err != nil {
				// If an error occurs, add it to the list and continue to the next filter.
				errors = append(errors, Error{
					Type:    "scenarios",
					Key:     i,
					Message: err.Error(),
				})
				continue
			}
		}

		// Once the scenarios are tested, delete created scenario specific attributes.
		for _, a := range createdScenarioAttributes {
			// Write the relationship to the database
			_, err = c.Container.DW.Delete(ctx, "t1", &v1.TupleFilter{},
				&v1.AttributeFilter{
					Entity: &v1.EntityFilter{
						Type: a.GetEntity().Type,
						Ids:  []string{a.Entity.Id},
					},
					Attributes: []string{a.GetAttribute()},
				},
			)
			// Continue to the next relationship if an error occurred
			if err != nil {
				// If an error occurs, add it to the list and continue to the next filter.
				errors = append(errors, Error{
					Type:    "scenarios",
					Key:     i,
					Message: err.Error(),
				})
				continue
			}
		}
	}

	return
}

// Context is a function that takes a file context and returns a base context and an error.
func Context(fileContext file.Context) (cont *v1.Context, err error) {
	// Initialize an empty base context to be populated from the file context.
	cont = &v1.Context{
		Tuples:     []*v1.Tuple{},
		Attributes: []*v1.Attribute{},
		Data:       nil,
	}

	// Convert the file context's data to a Struct object.
	st, err := structpb.NewStruct(fileContext.Data)
	if err != nil {
		// If an error occurs, return it.
		return nil, err
	}

	// Assign the Struct object to the context's data field.
	cont.Data = st

	// Iterate over the file context's tuples.
	for _, t := range fileContext.Tuples {
		// Convert each tuple to a base tuple.
		tup, err := tuple.Tuple(t)
		if err != nil {
			// If an error occurs, return it.
			return nil, err
		}

		// Add the converted tuple to the context's tuples slice.
		cont.Tuples = append(cont.Tuples, tup)
	}

	// Iterate over the file context's attributes.
	for _, t := range fileContext.Attributes {
		// Convert each attribute to a base attribute.
		attr, err := attribute.Attribute(t)
		if err != nil {
			// If an error occurs, return it.
			return nil, err
		}

		// Add the converted attribute to the context's attributes slice.
		cont.Attributes = append(cont.Attributes, attr)
	}

	// If everything goes well, return the context and a nil error.
	return cont, nil
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
