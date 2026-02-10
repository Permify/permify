package engines

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

// Mock implementations for testing

type mockSchemaReader struct{}

func (m *mockSchemaReader) ReadSchema(ctx context.Context, tenantID, version string) (*base.SchemaDefinition, error) {
	return nil, fmt.Errorf("mock schema reader error")
}

func (m *mockSchemaReader) ReadSchemaString(ctx context.Context, tenantID, version string) ([]string, error) {
	return nil, fmt.Errorf("mock schema reader error")
}

func (m *mockSchemaReader) ReadEntityDefinition(ctx context.Context, tenantID, entityName, version string) (*base.EntityDefinition, string, error) {
	return nil, "", fmt.Errorf("mock schema reader error")
}

func (m *mockSchemaReader) ReadRuleDefinition(ctx context.Context, tenantID, ruleName, version string) (*base.RuleDefinition, string, error) {
	return nil, "", fmt.Errorf("mock schema reader error")
}

func (m *mockSchemaReader) HeadVersion(ctx context.Context, tenantID string) (string, error) {
	return "", fmt.Errorf("mock schema reader error")
}

func (m *mockSchemaReader) ListSchemas(ctx context.Context, tenantID string, pagination database.Pagination) ([]*base.SchemaList, database.EncodedContinuousToken, error) {
	return nil, nil, fmt.Errorf("mock schema reader error")
}

var _ = Describe("lookup-entity-engine", func() {
	// DRIVE SAMPLE

	driveSchemaEntityFilter := `
	entity user {}
	
	entity organization {
		relation admin @user
	}
	
	entity folder {
		relation org @organization
		relation creator @user
		relation collaborator @user
	
		permission read = collaborator
		permission update = collaborator
		permission delete = creator or org.admin
	}
	
	entity doc {
		relation org @organization
		relation parent @folder
		relation owner @user
	
		permission read = (owner or parent.collaborator) or org.admin
		permission update = owner and org.admin
		permission delete = owner or org.admin
		permission share = update and (owner or parent.update)
	}
	`

	Context("Drive Sample: Entity Filter", func() {
		It("Drive Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				contextual    []string
				filters       []filter
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:1#folder@user:3",
					"folder:1#collaborator@user:1",
				},
				contextual: []string{
					"folder:1#collaborator@user:3",
					"organization:1#admin@user:1",
					"doc:1#org@organization:1#...",
				},
				filters: []filter{
					{
						entityType: "doc",
						subject:    "user:1",
						assertions: map[string][]string{
							"read": {"1"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			reqContext := &base.Context{
				Tuples:     []*base.Tuple{},
				Attributes: []*base.Attribute{},
			}

			for _, relationship := range tests.contextual {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				reqContext.Tuples = append(reqContext.Tuples, t)
			}

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
						Context: reqContext,
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 2", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"doc:2#owner@user:2",
					"doc:2#parent@folder:2#...",
					"folder:2#collaborator@user:3",
					"folder:2#creator@user:2",
					"organization:2#admin@user:2",
					"doc:2#org@organization:2#...",
				},
				filters: []filter{
					{
						entityType: "doc",
						subject:    "user:2",
						assertions: map[string][]string{
							"read":   {"2"},
							"update": {"2"},
							"delete": {"2"},
							"share":  {"2"},
						},
					},
					{
						entityType: "folder",
						subject:    "user:3",
						assertions: map[string][]string{
							"read":   {"2"},
							"update": {"2"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 3", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:1#parent@folder:1#...",
					"folder:1#collaborator@user:1",
					"folder:1#collaborator@user:3",
					"folder:1#creator@user:2",
					"organization:1#admin@user:1",
					"doc:1#org@organization:1#...",
				},
				filters: []filter{
					{
						entityType: "doc",
						subject:    "user:1",
						assertions: map[string][]string{
							"read": {"1"},
						},
					},
					{
						entityType: "folder",
						subject:    "user:2",
						assertions: map[string][]string{
							"delete": {"1"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 4", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:1#parent@folder:1#...",
					"folder:1#collaborator@user:1",
					"folder:1#collaborator@user:3",
					"folder:1#creator@user:2",
					"organization:1#admin@user:1",
					"doc:1#org@organization:1#...",
				},
				filters: []filter{
					{
						entityType: "doc",
						subject:    "user:1",
						assertions: map[string][]string{
							"read": {"1"},
						},
					},
					{
						entityType: "folder",
						subject:    "user:2",
						assertions: map[string][]string{
							"delete": {"1"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 5", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:2#owner@user:3",
					"doc:1#parent@folder:1#...",
					"doc:2#parent@folder:1#...",
					"folder:1#collaborator@user:1",
					"folder:1#collaborator@user:3",
					"folder:1#creator@user:2",
					"organization:1#admin@user:1",
					"doc:1#org@organization:1#...",
					"doc:2#org@organization:1#...",
				},
				filters: []filter{
					{
						entityType: "doc",
						subject:    "user:1",
						assertions: map[string][]string{
							"read": {"1", "2"},
						},
					},
					{
						entityType: "folder",
						subject:    "user:2",
						assertions: map[string][]string{
							"delete": {"1"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 6", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				contextual    []string
				filters       []filter
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:2#owner@user:3",
					"doc:3#owner@user:3",
					"doc:4#owner@user:2",
					"doc:5#owner@user:3",
					"doc:6#owner@user:2",
					"doc:1#parent@folder:1#...",
					"doc:2#parent@folder:1#...",
					"doc:3#parent@folder:1#...",
				},
				contextual: []string{
					"doc:4#parent@folder:1#...",
					"doc:5#parent@folder:1#...",
					"doc:6#parent@folder:1#...",
					"folder:1#collaborator@user:1",
					"folder:1#collaborator@user:3",
					"folder:1#creator@user:2",
					"organization:1#admin@user:1",
				},
				filters: []filter{
					{
						entityType: "doc",
						subject:    "user:1",
						assertions: map[string][]string{
							"read": {"1", "2", "3", "4", "5", "6"},
						},
					},
					{
						entityType: "folder",
						subject:    "user:2",
						assertions: map[string][]string{
							"delete": {"1"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			reqContext := &base.Context{
				Tuples:     []*base.Tuple{},
				Attributes: []*base.Attribute{},
			}

			for _, relationship := range tests.contextual {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				reqContext.Tuples = append(reqContext.Tuples, t)
			}

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
						Context: reqContext,
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 7", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:2#owner@user:3",
					"doc:3#owner@user:3",
					"doc:4#owner@user:2",
					"doc:5#owner@user:3",
					"doc:6#owner@user:2",
					"doc:7#owner@user:4",
					"doc:8#owner@user:4",
					"doc:9#owner@user:5",
					"doc:10#owner@user:5",
					"doc:1#parent@folder:1#...",
					"doc:2#parent@folder:1#...",
					"doc:3#parent@folder:2#...",
					"doc:4#parent@folder:2#...",
					"doc:5#parent@folder:3#...",
					"doc:6#parent@folder:3#...",
					"doc:7#parent@folder:4#...",
					"doc:8#parent@folder:4#...",
					"doc:9#parent@folder:5#...",
					"doc:10#parent@folder:5#...",
					"folder:1#collaborator@user:1",
					"folder:2#collaborator@user:1",
					"folder:3#collaborator@user:2",
					"folder:4#collaborator@user:2",
					"folder:5#collaborator@user:3",
					"folder:1#creator@user:2",
					"folder:2#creator@user:3",
					"folder:3#creator@user:4",
					"folder:4#creator@user:4",
					"folder:5#creator@user:5",
					"organization:1#admin@user:1",
					"organization:2#admin@user:2",
					"organization:3#admin@user:3",
					"doc:1#org@organization:1#...",
					"doc:2#org@organization:1#...",
					"doc:3#org@organization:2#...",
					"doc:4#org@organization:2#...",
					"doc:5#org@organization:3#...",
					"doc:6#org@organization:3#...",
					"doc:7#org@organization:1#...",
					"doc:8#org@organization:2#...",
					"doc:9#org@organization:3#...",
					"doc:10#org@organization:1#...",
				},
				filters: []filter{
					{
						entityType: "doc",
						subject:    "user:1",
						assertions: map[string][]string{
							"read": {"1", "10", "2", "3", "4", "7"},
						},
					},
					{
						entityType: "doc",
						subject:    "user:2",
						assertions: map[string][]string{
							"read":   {"1", "3", "4", "5", "6", "7", "8"},
							"delete": {"1", "3", "4", "6", "8"},
						},
					},
					{
						entityType: "doc",
						subject:    "user:3",
						assertions: map[string][]string{
							"read":   {"10", "2", "3", "5", "6", "9"},
							"update": {"5"},
						},
					},
					{
						entityType: "doc",
						subject:    "user:4",
						assertions: map[string][]string{
							"read":   {"7", "8"},
							"delete": {"7", "8"},
							"share":  nil,
						},
					},
					{
						entityType: "doc",
						subject:    "user:5",
						assertions: map[string][]string{
							"read":   {"10", "9"},
							"delete": {"10", "9"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 8 scope", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				scope      map[string]*base.StringArrayValue
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:2#owner@user:3",
					"doc:3#owner@user:3",
					"doc:4#owner@user:2",
					"doc:5#owner@user:3",
					"doc:6#owner@user:2",
					"doc:7#owner@user:4",
					"doc:8#owner@user:4",
					"doc:9#owner@user:5",
					"doc:10#owner@user:5",
					"doc:1#parent@folder:1#...",
					"doc:2#parent@folder:1#...",
					"doc:3#parent@folder:2#...",
					"doc:4#parent@folder:2#...",
					"doc:5#parent@folder:3#...",
					"doc:6#parent@folder:3#...",
					"doc:7#parent@folder:4#...",
					"doc:8#parent@folder:4#...",
					"doc:9#parent@folder:5#...",
					"doc:10#parent@folder:5#...",
					"folder:1#collaborator@user:1",
					"folder:2#collaborator@user:1",
					"folder:3#collaborator@user:2",
					"folder:4#collaborator@user:2",
					"folder:5#collaborator@user:3",
					"folder:1#creator@user:2",
					"folder:2#creator@user:3",
					"folder:3#creator@user:4",
					"folder:4#creator@user:4",
					"folder:5#creator@user:5",
					"organization:1#admin@user:1",
					"organization:2#admin@user:2",
					"organization:3#admin@user:3",
					"doc:1#org@organization:1#...",
					"doc:2#org@organization:1#...",
					"doc:3#org@organization:2#...",
					"doc:4#org@organization:2#...",
					"doc:5#org@organization:3#...",
					"doc:6#org@organization:3#...",
					"doc:7#org@organization:1#...",
					"doc:8#org@organization:2#...",
					"doc:9#org@organization:3#...",
					"doc:10#org@organization:1#...",
				},
				filters: []filter{
					{
						entityType: "doc",
						subject:    "user:1",
						scope: map[string]*base.StringArrayValue{
							"organization": {
								Data: []string{"2"},
							},
						},
						assertions: map[string][]string{
							"read": {"1", "2", "3", "4"},
						},
					},
					{
						entityType: "doc",
						subject:    "user:1",
						scope: map[string]*base.StringArrayValue{
							"organization": {
								Data: []string{"2"},
							},
							"folder": {
								Data: []string{"2"},
							},
						},
						assertions: map[string][]string{
							"read": {"3", "4"},
						},
					},
					{
						entityType: "doc",
						subject:    "user:2",
						scope: map[string]*base.StringArrayValue{
							"organization": {
								Data: []string{"1"},
							},
						},
						assertions: map[string][]string{
							"read":   {"1", "4", "5", "6", "7", "8"},
							"delete": {"1", "4", "6"},
						},
					},
					{
						entityType: "doc",
						subject:    "user:3",
						scope: map[string]*base.StringArrayValue{
							"organization": {
								Data: []string{"1", "2", "3"},
							},
						},
						assertions: map[string][]string{
							"read":   {"10", "2", "3", "5", "6", "9"},
							"update": {"5"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Scope:      filter.scope,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})
	})

	facebookGroupsSchemaEntityFilter := `
		entity user {}
	
		entity group {
	
		  // Relation to represent the members of the group
		  relation member @user
		  // Relation to represent the admins of the group
		  relation admin @user
		  // Relation to represent the moderators of the group
		  relation moderator @user
	
		  // Permissions for the group entity
		  action create = member
		  action join = member
		  action leave = member
		  action invite_to_group = admin
		  action remove_from_group = admin or moderator
		  action edit_settings = admin or moderator
		  action post_to_group = member
		  action comment_on_post = member
		  action view_group_insights = admin or moderator
		}
	
		entity post {
	
		  // Relation to represent the owner of the post
		  relation owner @user
		  // Relation to represent the group that the post belongs to
		  relation group @group
	
		  // Permissions for the post entity
		  action view_post = owner or group.member
		  action edit_post = owner or group.admin
		  action delete_post = owner or group.admin
	
		  permission group_member = group.member
		}
	
		entity comment {
	
		  // Relation to represent the owner of the comment
		  relation owner @user
	
		  // Relation to represent the post that the comment belongs to
		  relation post @post
	
		  // Permissions for the comment entity
		  action view_comment = owner or post.group_member
		  action edit_comment = owner
		  action delete_comment = owner
	
	     action remove = post.delete_post
		}
	
		entity like {
	
		  // Relation to represent the owner of the like
		  relation owner @user
	
		  // Relation to represent the post that the like belongs to
		  relation post @post
	
		  // Permissions for the like entity
		  action like_post = owner or post.group_member
		  action unlike_post = owner or post.group_member
		}
	
		entity poll {
	
		  // Relation to represent the owner of the poll
		  relation owner @user
	
		  // Relation to represent the group that the poll belongs to
		  relation group @group
	
		  // Permissions for the poll entity
		  action create_poll = owner or group.admin
		  action view_poll = owner or group.member
		  action edit_poll = owner or group.admin
		  action delete_poll = owner or group.admin
		}
	
		entity file {
	
		  // Relation to represent the owner of the file
		  relation owner @user
	
		  // Relation to represent the group that the file belongs to
		  relation group @group
	
		  // Permissions for the file entity
		  action upload_file = owner or group.member
		  action view_file = owner or group.member
		  action delete_file = owner or group.admin
		}
	
		entity event {
	
		  // Relation to represent the owner of the event
		  relation owner @user
		  // Relation to represent the group that the event belongs to
		  relation group @group
	
		  // Permissions for the event entity
		  action create_event = owner or group.admin
		  action view_event = owner or group.member
		  action edit_event = owner or group.admin
		  action delete_event = owner or group.admin
		  action RSVP_to_event = owner or group.member
		}
		`

	Context("Facebook Group Sample: Entity Filter", func() {
		It("Facebook Group Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(facebookGroupsSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"group:1#member@user:1",
					"group:1#admin@user:2",
					"group:1#moderator@user:3",
					"post:1#owner@user:1",
					"post:1#group@group:1#...",
					"comment:1#owner@user:1",
					"comment:1#post@post:1#...",
					"like:1#owner@user:1",
					"like:1#post@post:1#...",
					"poll:1#owner@user:2",
					"poll:1#group@group:1#...",
					"file:1#owner@user:3",
					"file:1#group@group:1#...",
					"event:1#owner@user:2",
					"event:1#group@group:1#...",
				},
				filters: []filter{
					{
						entityType: "group",
						subject:    "user:1",
						assertions: map[string][]string{
							"create":          {"1"},
							"join":            {"1"},
							"leave":           {"1"},
							"post_to_group":   {"1"},
							"comment_on_post": {"1"},
						},
					},
					{
						entityType: "post",
						subject:    "user:1",
						assertions: map[string][]string{
							"view_post": {"1"},
							"edit_post": {"1"},
						},
					},
					{
						entityType: "comment",
						subject:    "user:1",
						assertions: map[string][]string{
							"view_comment": {"1"},
							"edit_comment": {"1"},
						},
					},
					{
						entityType: "like",
						subject:    "user:1",
						assertions: map[string][]string{
							"like_post":   {"1"},
							"unlike_post": {"1"},
						},
					},
					{
						entityType: "poll",
						subject:    "user:2",
						assertions: map[string][]string{
							"create_poll": {"1"},
							"view_poll":   {"1"},
							"edit_poll":   {"1"},
						},
					},
					{
						entityType: "file",
						subject:    "user:3",
						assertions: map[string][]string{
							"upload_file": {"1"},
							"view_file":   {"1"},
						},
					},
					{
						entityType: "event",
						subject:    "user:2",
						assertions: map[string][]string{
							"create_event":  {"1"},
							"view_event":    {"1"},
							"edit_event":    {"1"},
							"RSVP_to_event": {"1"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})

		It("Facebook Group Sample: Case 2", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(facebookGroupsSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				contextual    []string
				filters       []filter
			}{
				relationships: []string{
					"group:1#member@user:1",
					"group:1#member@user:2",
					"group:1#member@user:3",
					"group:1#admin@user:4",
					"group:1#admin@user:5",
					"group:1#moderator@user:6",
					"group:1#moderator@user:7",
					"post:1#owner@user:1",
					"post:2#owner@user:2",
					"post:3#owner@user:3",
					"post:1#group@group:1#...",
					"post:2#group@group:1#...",
					"post:3#group@group:1#...",
					"comment:1#owner@user:1",
					"comment:2#owner@user:2",
					"comment:3#owner@user:3",
					"comment:1#post@post:1#...",
					"comment:2#post@post:2#...",
					"comment:3#post@post:3#...",
					"like:1#owner@user:1",
					"like:2#owner@user:2",
					"like:3#owner@user:3",
					"like:1#post@post:1#...",
					"like:2#post@post:2#...",
					"like:3#post@post:3#...",
					"poll:1#owner@user:4",
					"poll:2#owner@user:5",
					"poll:3#owner@user:6",
					"poll:1#group@group:1#...",
					"poll:2#group@group:1#...",
					"poll:3#group@group:1#...",
					"file:1#owner@user:7",
					"file:2#owner@user:8",
					"file:3#owner@user:9",
					"file:1#group@group:1#...",
					"file:2#group@group:1#...",
					"file:3#group@group:1#...",
					"event:1#owner@user:10",
					"event:2#owner@user:11",
					"event:3#owner@user:12",
					"event:1#group@group:1#...",
					"event:2#group@group:1#...",
					"event:3#group@group:1#...",
				},
				contextual: []string{
					"file:1#group@group:1#...",
					"file:2#group@group:1#...",
					"file:3#group@group:1#...",
					"event:1#owner@user:10",
					"event:2#owner@user:11",
					"event:3#owner@user:12",
					"event:1#group@group:1#...",
					"event:2#group@group:1#...",
					"event:3#group@group:1#...",
				},
				filters: []filter{
					{
						entityType: "group",
						subject:    "user:1",
						assertions: map[string][]string{
							"create":          {"1"},
							"join":            {"1"},
							"leave":           {"1"},
							"post_to_group":   {"1"},
							"comment_on_post": {"1"},
						},
					},
					{
						entityType: "post",
						subject:    "user:1",
						assertions: map[string][]string{
							"view_post": {"1", "2", "3"},
							"edit_post": {"1"},
						},
					},
					{
						entityType: "comment",
						subject:    "user:2",
						assertions: map[string][]string{
							"view_comment": {"1", "2", "3"},
							"edit_comment": {"2"},
						},
					},
					{
						entityType: "like",
						subject:    "user:3",
						assertions: map[string][]string{
							"like_post":   {"1", "2", "3"},
							"unlike_post": {"1", "2", "3"},
						},
					},
					{
						entityType: "poll",
						subject:    "user:4",
						assertions: map[string][]string{
							"create_poll": {"1", "2", "3"},
							"view_poll":   {"1"},
							"edit_poll":   {"1", "2", "3"},
						},
					},
					{
						entityType: "file",
						subject:    "user:5",
						assertions: map[string][]string{
							"upload_file": nil,
							"view_file":   nil,
							"delete_file": {"1", "2", "3"},
						},
					},
					{
						entityType: "event",
						subject:    "user:6",
						assertions: map[string][]string{
							"create_event":  nil,
							"view_event":    nil,
							"edit_event":    nil,
							"delete_event":  nil,
							"RSVP_to_event": nil,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			reqContext := &base.Context{
				Tuples:     []*base.Tuple{},
				Attributes: []*base.Attribute{},
			}

			for _, relationship := range tests.contextual {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				reqContext.Tuples = append(reqContext.Tuples, t)
			}

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
						Context: reqContext,
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})

		It("Facebook Group Sample: Case 3 pagination", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(facebookGroupsSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"group:1#member@user:1",
					"group:2#member@user:1",
					"group:3#member@user:1",
					"group:4#member@user:1",

					"post:99#group@group:1#...",
					"post:98#group@group:2#...",
					"post:97#group@group:3#...",
					"post:96#group@group:4#...",
					"post:96#group@group:4#...",
					"post:95#group@group:4#...",
					"post:94#group@group:4#...",
					"post:93#group@group:4#...",
					"post:92#group@group:4#...",
				},
				filters: []filter{
					{
						entityType: "post",
						subject:    "user:1",
						assertions: map[string][]string{
							"view_post": {"92", "93", "94", "95", "96", "97", "98", "99"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {

					ct := ""

					var ids []string

					for {
						response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
							TenantId:   "t1",
							EntityType: filter.entityType,
							Subject:    subject,
							Permission: permission,
							Metadata: &base.PermissionLookupEntityRequestMetadata{
								SnapToken:     token.NewNoopToken().Encode().String(),
								SchemaVersion: "",
								Depth:         100,
							},
							PageSize:        5,
							ContinuousToken: ct,
						})
						Expect(err).ShouldNot(HaveOccurred())

						ids = append(ids, response.GetEntityIds()...)

						ct = response.GetContinuousToken()

						if ct == "" {
							break
						}
					}

					Expect(ids).Should(Equal(res))
				}
			}
		})

		It("Facebook Group Sample: Case 4 pagination", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(facebookGroupsSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"group:1#admin@user:1",
					"group:2#admin@user:1",
					"group:3#admin@user:1",
					"group:4#admin@user:1",

					"post:59#group@group:1#...",
					"post:58#group@group:2#...",
					"post:57#group@group:3#...",
					"post:56#group@group:4#...",
					"post:55#group@group:4#...",
					"post:54#group@group:4#...",
					"post:53#group@group:4#...",
					"post:52#group@group:4#...",

					"comment:99#post@post:58#...",
					"comment:98#post@post:58#...",
					"comment:97#post@post:54#...",
					"comment:96#post@post:4#...",
					"comment:96#post@post:57#...",
					"comment:95#post@post:54#...",
					"comment:94#post@post:54#...",
					"comment:93#post@post:54#...",
					"comment:92#post@post:53#...",
					"comment:91#post@post:53#...",
					"comment:90#post@post:53#...",
					"comment:45#post@post:53#...",
					"comment:1#post@post:53#...",
				},
				filters: []filter{
					{
						entityType: "comment",
						subject:    "user:1",
						assertions: map[string][]string{
							"remove": {"1", "45", "90", "91", "92", "93", "94", "95", "96", "97", "98", "99"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {

					ct := ""

					var ids []string

					for {
						response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
							TenantId:   "t1",
							EntityType: filter.entityType,
							Subject:    subject,
							Permission: permission,
							Metadata: &base.PermissionLookupEntityRequestMetadata{
								SnapToken:     token.NewNoopToken().Encode().String(),
								SchemaVersion: "",
								Depth:         100,
							},
							PageSize:        5,
							ContinuousToken: ct,
						})
						Expect(err).ShouldNot(HaveOccurred())

						ids = append(ids, response.GetEntityIds()...)

						ct = response.GetContinuousToken()

						if ct == "" {
							break
						}
					}

					Expect(ids).Should(Equal(res))
				}
			}
		})
	})

	googleDocsSchemaEntityFilter := `
		entity user {}
	
		entity resource {
		  relation viewer  @user  @group#member @group#manager
		  relation manager @user @group#member @group#manager
	
		  action edit = manager
		  action view = viewer or manager
		}
	
		entity group {
		  relation manager @user @group#member @group#manager
		  relation member @user @group#member @group#manager
		}
	
		entity organization {
		  relation group @group
		  relation resource @resource
	
		  relation administrator @user @group#member @group#manager
		  relation direct_member @user
	
		  permission admin = administrator
		  permission member = direct_member or administrator or group.member
		}`

	Context("Google Docs Sample: Entity Filter", func() {
		It("Google Docs Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(googleDocsSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"resource:1#viewer@user:1",
					"resource:1#viewer@group:1#member",
					"resource:1#manager@user:2",
					"resource:1#manager@group:1#manager",
					"group:1#manager@user:3",
					"group:1#manager@group:1#member",
					"group:1#member@user:4",
					"group:1#member@group:1#manager",
					"organization:1#group@group:1#...",
					"organization:1#resource@resource:1#...",
					"organization:1#administrator@user:5",
					"organization:1#administrator@group:1#manager",
					"organization:1#direct_member@user:6",
				},
				filters: []filter{
					{
						entityType: "resource",
						subject:    "user:1",
						assertions: map[string][]string{
							"view": {"1"},
						},
					},
					{
						entityType: "resource",
						subject:    "group:1#member",
						assertions: map[string][]string{
							"view": {"1"},
						},
					},
					{
						entityType: "resource",
						subject:    "user:2",
						assertions: map[string][]string{
							"edit": {"1"},
							"view": {"1"},
						},
					},
					{
						entityType: "organization",
						subject:    "user:5",
						assertions: map[string][]string{
							"admin":  {"1"},
							"member": {"1"},
						},
					},
					{
						entityType: "organization",
						subject:    "group:1#manager",
						assertions: map[string][]string{
							"admin":  {"1"},
							"member": {"1"},
						},
					},
					{
						entityType: "organization",
						subject:    "user:6",
						assertions: map[string][]string{
							"member": {"1"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         500, // High depth for multi-hop group/org traversal
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})

		It("Google Docs Sample: Case 2", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(googleDocsSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"resource:1#viewer@user:1",
					"resource:1#viewer@group:1#member",
					"resource:1#viewer@group:2#manager",
					"resource:1#manager@user:2",
					"resource:1#manager@group:1#manager",
					"resource:1#manager@group:2#member",
					"resource:2#viewer@user:3",
					"resource:2#viewer@group:2#manager",
					"resource:2#manager@user:4",
					"resource:2#manager@group:1#manager",
					"group:1#manager@user:5",
					"group:1#manager@group:2#member",
					"group:1#member@user:6",
					"group:1#member@group:2#manager",
					"group:2#manager@user:7",
					"group:2#manager@group:1#member",
					"group:2#member@user:8",
					"group:2#member@group:1#manager",
					"organization:1#group@group:1#...",
					"organization:1#group@group:2#...",
					"organization:1#resource@resource:1#...",
					"organization:1#resource@resource:2#...",
					"organization:1#administrator@user:9",
					"organization:1#administrator@group:1#manager",
					"organization:1#administrator@group:2#member",
					"organization:1#direct_member@user:10",
				},
				filters: []filter{
					{
						entityType: "resource",
						subject:    "user:1",
						assertions: map[string][]string{
							"view": {"1"},
						},
					},
					{
						entityType: "resource",
						subject:    "group:2#manager",
						assertions: map[string][]string{
							"view": {"1", "2"},
						},
					},
					{
						entityType: "resource",
						subject:    "user:4",
						assertions: map[string][]string{
							"edit": {"2"},
							"view": {"2"},
						},
					},
					{
						entityType: "group",
						subject:    "user:5",
						assertions: map[string][]string{
							"member": {"2"},
						},
					},
					{
						entityType: "group",
						subject:    "group:1#manager",
						assertions: map[string][]string{
							"member": {"2"},
						},
					},
					{
						entityType: "organization",
						subject:    "user:9",
						assertions: map[string][]string{
							"admin":  {"1"},
							"member": {"1"},
						},
					},
					{
						entityType: "organization",
						subject:    "group:2#member",
						assertions: map[string][]string{
							"admin":  {"1"},
							"member": {"1"},
						},
					},
					{
						entityType: "organization",
						subject:    "user:10",
						assertions: map[string][]string{
							"member": {"1"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         500, // High depth for multi-hop group/org traversal
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})

		It("Google Docs Sample: Case 3", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(googleDocsSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{}

			for i := 1; i <= 50; i++ {
				relationship := fmt.Sprintf("resource:%d#viewer@user:1", i)
				tests.relationships = append(tests.relationships, relationship)
			}

			// Generate 50 manager relationships.
			for i := 51; i <= 100; i++ {
				relationship := fmt.Sprintf("resource:%d#manager@user:1", i)
				tests.relationships = append(tests.relationships, relationship)
			}

			tests.filters = []filter{
				{
					entityType: "resource",
					subject:    "user:1",
					assertions: map[string][]string{
						"view": {"1", "10", "100", "11", "12", "13", "14", "15", "16", "17", "18", "19", "2", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "3", "30", "31", "32", "33", "34", "35", "36", "37", "38", "39", "4", "40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "5", "50", "51", "52", "53", "54", "55", "56", "57", "58", "59", "6", "60", "61", "62", "63", "64", "65", "66", "67", "68", "69", "7", "70", "71", "72", "73", "74", "75", "76", "77", "78", "79", "8", "80", "81", "82", "83", "84", "85", "86", "87", "88", "89", "9", "90", "91", "92", "93", "94", "95", "96", "97", "98", "99"},
						"edit": {"100", "51", "52", "53", "54", "55", "56", "57", "58", "59", "60", "61", "62", "63", "64", "65", "66", "67", "68", "69", "70", "71", "72", "73", "74", "75", "76", "77", "78", "79", "80", "81", "82", "83", "84", "85", "86", "87", "88", "89", "90", "91", "92", "93", "94", "95", "96", "97", "98", "99"},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})

		It("Google Docs Sample: Case 4 pagination", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(googleDocsSchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{}

			for i := 1; i <= 50; i++ {
				relationship := fmt.Sprintf("resource:%d#viewer@user:1", i)
				tests.relationships = append(tests.relationships, relationship)
			}

			// Generate 50 manager relationships.
			for i := 51; i <= 100; i++ {
				relationship := fmt.Sprintf("resource:%d#manager@user:1", i)
				tests.relationships = append(tests.relationships, relationship)
			}

			tests.filters = []filter{
				{
					entityType: "resource",
					subject:    "user:1",
					assertions: map[string][]string{
						"view": {"1", "10", "100", "11", "12", "13", "14", "15", "16", "17", "18", "19", "2", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "3", "30", "31", "32", "33", "34", "35", "36", "37", "38", "39", "4", "40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "5", "50", "51", "52", "53", "54", "55", "56", "57", "58", "59", "6", "60", "61", "62", "63", "64", "65", "66", "67", "68", "69", "7", "70", "71", "72", "73", "74", "75", "76", "77", "78", "79", "8", "80", "81", "82", "83", "84", "85", "86", "87", "88", "89", "9", "90", "91", "92", "93", "94", "95", "96", "97", "98", "99"},
						"edit": {"100", "51", "52", "53", "54", "55", "56", "57", "58", "59", "60", "61", "62", "63", "64", "65", "66", "67", "68", "69", "70", "71", "72", "73", "74", "75", "76", "77", "78", "79", "80", "81", "82", "83", "84", "85", "86", "87", "88", "89", "90", "91", "92", "93", "94", "95", "96", "97", "98", "99"},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {

					ct := ""

					var ids []string

					for {
						response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
							TenantId:   "t1",
							EntityType: filter.entityType,
							Subject:    subject,
							Permission: permission,
							Metadata: &base.PermissionLookupEntityRequestMetadata{
								SnapToken:     token.NewNoopToken().Encode().String(),
								SchemaVersion: "",
								Depth:         100,
							},
							PageSize:        10,
							ContinuousToken: ct,
						})
						Expect(err).ShouldNot(HaveOccurred())

						ids = append(ids, response.GetEntityIds()...)

						ct = response.GetContinuousToken()

						if ct == "" {
							break
						}
					}

					Expect(ids).Should(Equal(res))
				}
			}
		})
	})

	workdaySchemaEntityFilter := `
			entity user {}
	
			entity organization {
	
				relation member @user
	
				attribute balance integer
	
				permission view = check_balance(balance) and member
			}
	
			entity repository {
	
				relation organization  @organization
	
				attribute is_public boolean
	
				permission view = is_public
				permission edit = organization.view
				permission delete = is_workday(is_public)
			}
	
			rule check_balance(balance integer) {
				balance > 5000
			}
	
			rule is_workday(is_public boolean) {
				  is_public && (context.data.day_of_week != 'saturday' && context.data.day_of_week != 'sunday')
			}
			`

	Context("Weekday Sample: Entity Filter", func() {
		It("Weekday Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(workdaySchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				attributes    []string
				filters       []filter
			}{
				relationships: []string{
					"organization:1#member@user:1",
					"organization:2#member@user:1",
					"organization:4#member@user:1",
					"organization:8#member@user:1",
					"organization:917#member@user:1",
					"organization:20#member@user:1",
					"organization:45#member@user:1",
					"repository:4#organization@organization:1",

					"organization:2#member@user:1",
				},
				attributes: []string{
					"repository:1$is_public|boolean:true",
					"repository:2$is_public|boolean:false",
					"repository:3$is_public|boolean:true",
					"repository:4$is_public|boolean:true",
					"repository:5$is_public|boolean:true",
					"repository:6$is_public|boolean:false",

					"organization:1$balance|integer:4000",
					"organization:2$balance|integer:6000",
					"organization:4$balance|integer:6000",
					"organization:8$balance|integer:6000",
					"organization:917$balance|integer:6000",
					"organization:20$balance|integer:6000",
					"organization:45$balance|integer:6000",
				},
				filters: []filter{
					{
						entityType: "repository",
						subject:    "user:1",
						assertions: map[string][]string{
							"view": {"1", "3", "4", "5"},
						},
					},
					{
						entityType: "organization",
						subject:    "user:1",
						assertions: map[string][]string{
							"view": {"2", "20", "4", "45", "8", "917"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			var attributes []*base.Attribute

			for _, attr := range tests.attributes {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})

		It("Weekday Sample: Case 2 pagination", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(workdaySchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				attributes    []string
				filters       []filter
			}{
				relationships: []string{
					"organization:1#member@user:1",
					"organization:2#member@user:1",
					"organization:4#member@user:1",
					"organization:8#member@user:1",
					"organization:917#member@user:1",
					"organization:20#member@user:1",
					"organization:45#member@user:1",
					"organization:22#member@user:1",
					"organization:43#member@user:1",
					"organization:84#member@user:1",
					"organization:9157#member@user:1",
					"organization:260#member@user:1",
					"organization:475#member@user:1",
					"repository:4#organization@organization:1",

					"organization:2#member@user:1",
				},
				attributes: []string{
					"repository:1$is_public|boolean:true",
					"repository:2$is_public|boolean:false",
					"repository:3$is_public|boolean:true",
					"repository:4$is_public|boolean:true",
					"repository:5$is_public|boolean:true",
					"repository:6$is_public|boolean:false",

					"organization:1$balance|integer:4000",
					"organization:2$balance|integer:6000",
					"organization:4$balance|integer:6000",
					"organization:8$balance|integer:6000",
					"organization:917$balance|integer:6000",
					"organization:20$balance|integer:6000",
					"organization:45$balance|integer:6000",

					"organization:22$balance|integer:6000",
					"organization:43$balance|integer:6000",
					"organization:84$balance|integer:6000",
					"organization:9157$balance|integer:6000",
					"organization:260$balance|integer:6000",
					"organization:475$balance|integer:6000",
				},
				filters: []filter{
					{
						entityType: "organization",
						subject:    "user:1",
						assertions: map[string][]string{
							"view": {"2", "20", "22", "260", "4", "43", "45", "475", "8", "84", "9157", "917"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			var attributes []*base.Attribute

			for _, attr := range tests.attributes {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {

					ct := ""

					var ids []string

					for {
						response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
							TenantId:   "t1",
							EntityType: filter.entityType,
							Subject:    subject,
							Permission: permission,
							Metadata: &base.PermissionLookupEntityRequestMetadata{
								SnapToken:     token.NewNoopToken().Encode().String(),
								SchemaVersion: "",
								Depth:         100,
							},
							PageSize:        10,
							ContinuousToken: ct,
						})
						Expect(err).ShouldNot(HaveOccurred())

						ids = append(ids, response.GetEntityIds()...)

						ct = response.GetContinuousToken()

						if ct == "" {
							break
						}
					}

					Expect(ids).Should(Equal(res))
				}
			}
		})

		It("Weekday Sample: Case 3 scope", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(workdaySchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				scope      map[string]*base.StringArrayValue
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				attributes    []string
				filters       []filter
			}{
				relationships: []string{
					"organization:1#member@user:1",
					"organization:2#member@user:1",
					"organization:4#member@user:1",
					"organization:8#member@user:1",
					"organization:917#member@user:1",
					"organization:20#member@user:1",
					"organization:45#member@user:1",
					"repository:4#organization@organization:1",

					"organization:2#member@user:1",
				},
				attributes: []string{
					"repository:1$is_public|boolean:true",
					"repository:2$is_public|boolean:false",
					"repository:3$is_public|boolean:true",
					"repository:4$is_public|boolean:true",
					"repository:5$is_public|boolean:true",
					"repository:6$is_public|boolean:false",

					"organization:1$balance|integer:4000",
					"organization:2$balance|integer:6000",
					"organization:4$balance|integer:6000",
					"organization:8$balance|integer:6000",
					"organization:917$balance|integer:6000",
					"organization:20$balance|integer:6000",
					"organization:45$balance|integer:6000",
				},
				filters: []filter{
					{
						entityType: "repository",
						subject:    "user:1",
						assertions: map[string][]string{
							"view": {"1", "3", "4", "5"},
						},
					},
					{
						entityType: "organization",
						subject:    "user:1",
						assertions: map[string][]string{
							"view": {"2", "20", "4", "45", "8", "917"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			var attributes []*base.Attribute

			for _, attr := range tests.attributes {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Scope:      filter.scope,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})
	})

	PropagationAcrossEntitiesEntityFilter := `
entity user {}

entity aaa {
    relation role__admin @user
    permission ccc__read = role__admin
}

entity bbb {
    relation resource__aaa @aaa
    relation role__admin @user
    attribute attr__is_public boolean
    permission ccc__read = role__admin or attr__is_public

}

entity ccc {
    relation resource__aaa @aaa
    relation resource__bbb @bbb
    permission ccc__read = resource__aaa.ccc__read or resource__bbb.ccc__read
}`

	Context("Propagation Across Entities: Entity Filter", func() {
		It("Drive Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(PropagationAcrossEntitiesEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				entityType string
				subject    string
				assertions map[string][]string
			}

			tests := struct {
				relationships []string
				attributes    []string
				filters       []filter
			}{
				relationships: []string{
					"aaa:a1#role__admin@user:u1",
					"bbb:b1#resource__aaa@aaa:a1",
					"ccc:c1#resource__aaa@aaa:a1",
					"ccc:c1#resource__bbb@bbb:b1",
				},
				attributes: []string{
					"bbb:b1$attr__is_public|boolean:true",
				},
				filters: []filter{
					{
						entityType: "ccc",
						subject:    "user:u1",
						assertions: map[string][]string{
							"ccc__read": {"c1"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			var attributes []*base.Attribute

			for _, attr := range tests.attributes {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				ear, err := tuple.EAR(filter.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range filter.assertions {
					response, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
						TenantId:   "t1",
						EntityType: filter.entityType,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionLookupEntityRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetEntityIds()).Should(Equal(res))
				}
			}
		})
	})

	driveSchemaSubjectFilter := `
	entity user {}
	
	entity organization {
		relation admin @user
	}
	
	entity folder {
		relation org @organization
		relation creator @user
		relation collaborator @user
	
		permission read = collaborator
		permission update = collaborator
		permission delete = creator or org.admin
		permission share = update
	}
	
	entity doc {
		relation org @organization
		relation parent @folder
	
		relation owner @user @organization#admin
		relation member @user
	
		permission read = owner or member
		permission update = owner and org.admin
		permission delete = owner or org.admin
		permission share = update and (member not parent.update)
		permission remove = owner or parent.delete
	}`

	Context("Drive Sample: Subject Filter", func() {
		It("Drive Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				subjectReference string
				entity           string
				assertions       map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:1#owner@user:1",
					"doc:1#member@user:1",
				},
				filters: []filter{
					{
						subjectReference: "user",
						entity:           "doc:1",
						assertions: map[string][]string{
							"read": {"1", "2"},
						},
					},
				},
			}

			// filters

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				entity, err := tuple.E(filter.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range filter.assertions {
					response, err := invoker.LookupSubject(context.Background(), &base.PermissionLookupSubjectRequest{
						TenantId:         "t1",
						SubjectReference: tuple.RelationReference(filter.subjectReference),
						Entity:           entity,
						Permission:       permission,
						Metadata: &base.PermissionLookupSubjectRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetSubjectIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 2", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				subjectReference string
				entity           string
				assertions       map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"organization:1#admin@user:1",
					"organization:1#admin@user:2",
					"doc:1#org@organization:1#...",
					"doc:1#owner@user:1",
					"doc:1#owner@user:2",
				},
				filters: []filter{
					{
						subjectReference: "user",
						entity:           "doc:1",
						assertions: map[string][]string{
							"update": {"1", "2"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				entity, err := tuple.E(filter.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range filter.assertions {
					response, err := invoker.LookupSubject(context.Background(), &base.PermissionLookupSubjectRequest{
						TenantId:         "t1",
						SubjectReference: tuple.RelationReference(filter.subjectReference),
						Entity:           entity,
						Permission:       permission,
						Metadata: &base.PermissionLookupSubjectRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetSubjectIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 3", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				subjectReference string
				entity           string
				assertions       map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"organization:1#admin@user:1",
					"organization:1#admin@user:2",
					"organization:1#admin@user:3",
					"doc:1#member@user:1",
					"doc:1#member@user:2",
					"doc:1#member@user:3",
					"doc:1#org@organization:1#...",
					"doc:1#owner@user:1",
					"doc:1#owner@user:2",
					"doc:1#owner@user:3",
					"doc:1#parent@folder:1#...",
					"folder:1#collaborator@user:3",
				},
				filters: []filter{
					{
						subjectReference: "user",
						entity:           "doc:1",
						assertions: map[string][]string{
							"share": {"1", "2"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				entity, err := tuple.E(filter.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range filter.assertions {
					response, err := invoker.LookupSubject(context.Background(), &base.PermissionLookupSubjectRequest{
						TenantId:         "t1",
						SubjectReference: tuple.RelationReference(filter.subjectReference),
						Entity:           entity,
						Permission:       permission,
						Metadata: &base.PermissionLookupSubjectRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetSubjectIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 4", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				subjectReference string
				entity           string
				assertions       map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"folder:1#collaborator@user:1",
					"folder:1#collaborator@user:2",
				},
				filters: []filter{
					{
						subjectReference: "user",
						entity:           "folder:1",
						assertions: map[string][]string{
							"share": {"1", "2"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				entity, err := tuple.E(filter.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range filter.assertions {
					response, err := invoker.LookupSubject(context.Background(), &base.PermissionLookupSubjectRequest{
						TenantId:         "t1",
						SubjectReference: tuple.RelationReference(filter.subjectReference),
						Entity:           entity,
						Permission:       permission,
						Metadata: &base.PermissionLookupSubjectRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetSubjectIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 5", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				subjectReference string
				entity           string
				assertions       map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"doc:1#owner@user:1",
					"doc:1#owner@user:3",
					"organization:1#admin@user:8",
					"doc:1#owner@organization:1#admin",
				},
				filters: []filter{
					{
						subjectReference: "user",
						entity:           "doc:1",
						assertions: map[string][]string{
							"delete": {"1", "3", "8"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				entity, err := tuple.E(filter.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range filter.assertions {
					response, err := invoker.LookupSubject(context.Background(), &base.PermissionLookupSubjectRequest{
						TenantId:         "t1",
						SubjectReference: tuple.RelationReference(filter.subjectReference),
						Entity:           entity,
						Permission:       permission,
						Metadata: &base.PermissionLookupSubjectRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetSubjectIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 6", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				subjectReference string
				entity           string
				assertions       map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"doc:1#owner@user:1",
					"doc:1#owner@user:3",
					"organization:1#admin@user:8",
					"doc:1#owner@organization:1#admin",
				},
				filters: []filter{
					{
						subjectReference: "organization#admin",
						entity:           "doc:1",
						assertions: map[string][]string{
							"delete": {"1"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				entity, err := tuple.E(filter.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range filter.assertions {
					response, err := invoker.LookupSubject(context.Background(), &base.PermissionLookupSubjectRequest{
						TenantId:         "t1",
						SubjectReference: tuple.RelationReference(filter.subjectReference),
						Entity:           entity,
						Permission:       permission,
						Metadata: &base.PermissionLookupSubjectRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetSubjectIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 7", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				subjectReference string
				entity           string
				assertions       map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"doc:1#owner@user:1",
					"doc:1#owner@user:3",
					"organization:1#admin@user:8",
					"organization:2#admin@user:32",
					"organization:3#admin@user:43",
					"organization:4#admin@user:65",
					"doc:1#owner@organization:1#admin",
					"doc:1#owner@organization:2#admin",
					"doc:1#owner@organization:3#admin",
				},
				filters: []filter{
					{
						subjectReference: "organization#admin",
						entity:           "doc:1",
						assertions: map[string][]string{
							"delete": {"1", "2", "3"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				entity, err := tuple.E(filter.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range filter.assertions {
					response, err := invoker.LookupSubject(context.Background(), &base.PermissionLookupSubjectRequest{
						TenantId:         "t1",
						SubjectReference: tuple.RelationReference(filter.subjectReference),
						Entity:           entity,
						Permission:       permission,
						Metadata: &base.PermissionLookupSubjectRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetSubjectIds()).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 8 pagination", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				subjectReference string
				entity           string
				assertions       map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"doc:1#owner@user:1",
					"doc:1#owner@user:3",
					"organization:1#admin@user:8",
					"organization:2#admin@user:32",
					"organization:3#admin@user:43",
					"organization:4#admin@user:65",
					"doc:1#owner@organization:1#admin",
					"doc:1#owner@organization:2#admin",
					"doc:1#owner@organization:3#admin",
					"doc:1#owner@organization:4#admin",
					"doc:1#owner@organization:5#admin",
					"doc:1#owner@organization:6#admin",
					"doc:1#owner@organization:7#admin",
				},
				filters: []filter{
					{
						subjectReference: "organization#admin",
						entity:           "doc:1",
						assertions: map[string][]string{
							"delete": {"1", "2", "3", "4", "5", "6", "7"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				entity, err := tuple.E(filter.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range filter.assertions {

					ct := ""

					var ids []string

					for {
						response, err := invoker.LookupSubject(context.Background(), &base.PermissionLookupSubjectRequest{
							TenantId:         "t1",
							SubjectReference: tuple.RelationReference(filter.subjectReference),
							Entity:           entity,
							Permission:       permission,
							Metadata: &base.PermissionLookupSubjectRequestMetadata{
								SnapToken:     token.NewNoopToken().Encode().String(),
								SchemaVersion: "",
							},
							ContinuousToken: ct,
							PageSize:        5,
						})
						Expect(err).ShouldNot(HaveOccurred())

						ids = append(ids, response.GetSubjectIds()...)

						ct = response.GetContinuousToken()

						if ct == "" {
							break
						}
					}

					Expect(ids).Should(Equal(res))
				}
			}
		})

		It("Drive Sample: Case 9 pagination", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				subjectReference string
				entity           string
				assertions       map[string][]string
			}

			tests := struct {
				relationships []string
				filters       []filter
			}{
				relationships: []string{
					"doc:99#owner@user:1",
					"doc:99#owner@user:3",

					"organization:98#admin@user:101",

					"organization:11#admin@user:99",
					"organization:12#admin@user:98",
					"organization:13#admin@user:97",
					"organization:14#admin@user:96",
					"organization:14#admin@user:95",

					"folder:1#org@organization:11",
					"folder:2#org@organization:12",
					"folder:3#org@organization:13",
					"folder:4#org@organization:14",

					"doc:99#parent@folder:1",
					"doc:99#parent@folder:2",
					"doc:99#parent@folder:3",
					"doc:99#parent@folder:4",

					"doc:99#owner@organization:98#admin",
				},
				filters: []filter{
					{
						subjectReference: "user",
						entity:           "doc:99",
						assertions: map[string][]string{
							"remove": {"1", "101", "3", "95", "96", "97", "98", "99"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				entity, err := tuple.E(filter.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range filter.assertions {

					ct := ""

					var ids []string

					for {
						response, err := invoker.LookupSubject(context.Background(), &base.PermissionLookupSubjectRequest{
							TenantId:         "t1",
							SubjectReference: tuple.RelationReference(filter.subjectReference),
							Entity:           entity,
							Permission:       permission,
							Metadata: &base.PermissionLookupSubjectRequestMetadata{
								SnapToken:     token.NewNoopToken().Encode().String(),
								SchemaVersion: "",
							},
							ContinuousToken: ct,
							PageSize:        5,
						})
						Expect(err).ShouldNot(HaveOccurred())

						ids = append(ids, response.GetSubjectIds()...)

						ct = response.GetContinuousToken()

						if ct == "" {
							break
						}
					}

					Expect(ids).Should(Equal(res))
				}
			}
		})
	})

	workdaySchemaSubjectFilter := `
			entity user {}
	
			entity organization {
	
				relation member @user
	
				attribute balance integer
	
				permission view = check_balance(balance) and member
				permission delete = check_balance(balance) not member
			}
	
			entity repository {
	
				relation organization  @organization
				relation member  @user
	
				attribute is_public boolean
	
				permission view = is_public
				permission edit = organization.view
				permission delete = is_workday(is_public)
				permission up = is_public not organization.member
				permission deploy = is_public not member
				permission check = is_public and organization.delete
			}
	
			rule check_balance(balance integer) {
				balance > 5000
			}
	
			rule is_workday(is_public boolean) {
				  is_public == true && (context.data.day_of_week != 'saturday' && context.data.day_of_week != 'sunday')
			}
			`

	Context("Weekday Sample: Subject Filter", func() {
		It("Weekday Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(workdaySchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				subjectReference string
				entity           string
				assertions       map[string][]string
			}

			tests := struct {
				relationships []string
				attributes    []string
				filters       []filter
			}{
				relationships: []string{
					"organization:1#member@user:1",
					"repository:4#organization@organization:1",

					"organization:2#member@user:1",
					"organization:2#member@user:3",
					"organization:5#member@user:2",
					"organization:5#member@user:5",
				},
				attributes: []string{
					"repository:1$is_public|boolean:true",
					"repository:2$is_public|boolean:false",
					"repository:3$is_public|boolean:true",

					"organization:1$balance|integer:4000",
					"organization:2$balance|integer:6000",
				},
				filters: []filter{
					{
						subjectReference: "user",
						entity:           "repository:1",
						assertions: map[string][]string{
							"view": {"1", "2", "3", "5"},
						},
					},
					{
						subjectReference: "user",
						entity:           "organization:2",
						assertions: map[string][]string{
							"view": {"1", "3"},
						},
					},
				},
			}

			// filters

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			var attributes []*base.Attribute

			for _, attr := range tests.attributes {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				entity, err := tuple.E(filter.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range filter.assertions {
					response, err := invoker.LookupSubject(context.Background(), &base.PermissionLookupSubjectRequest{
						TenantId:         "t1",
						SubjectReference: tuple.RelationReference(filter.subjectReference),
						Entity:           entity,
						Permission:       permission,
						Metadata: &base.PermissionLookupSubjectRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetSubjectIds()).Should(Equal(res))
				}
			}
		})

		It("Weekday Sample: Case 2", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(workdaySchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				subjectReference string
				entity           string
				assertions       map[string][]string
			}

			tests := struct {
				relationships []string
				attributes    []string
				filters       []filter
			}{
				relationships: []string{
					"organization:1#member@user:1",
					"repository:4#organization@organization:1",

					"repository:3#organization@organization:1",
					"repository:1#organization@organization:1",

					"organization:2#member@user:1",
					"organization:2#member@user:3",
					"organization:5#member@user:2",
					"organization:5#member@user:5",

					"repository:12#member@user:1",
					"repository:12#member@user:2",

					"repository:82#organization@organization:43",

					"organization:43#member@user:90",
					"organization:43#member@user:54",
				},
				attributes: []string{
					"repository:1$is_public|boolean:true",
					"repository:2$is_public|boolean:false",
					"repository:3$is_public|boolean:true",
					"repository:12$is_public|boolean:true",
					"repository:82$is_public|boolean:true",

					"organization:1$balance|integer:4000",
					"organization:2$balance|integer:6000",

					"organization:43$balance|integer:6000",
				},
				filters: []filter{
					{
						subjectReference: "user",
						entity:           "repository:1",
						assertions: map[string][]string{
							"up": {"2", "3", "5", "54", "90"},
						},
					},
					{
						subjectReference: "user",
						entity:           "repository:3",
						assertions: map[string][]string{
							"up": {"2", "3", "5", "54", "90"},
						},
					},
					{
						subjectReference: "user",
						entity:           "repository:12",
						assertions: map[string][]string{
							"deploy": {"3", "5", "54", "90"},
						},
					},
					{
						subjectReference: "user",
						entity:           "repository:82",
						assertions: map[string][]string{
							"check": {"1", "2", "3", "5"},
						},
					},
				},
			}

			// filters

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			var attributes []*base.Attribute

			for _, attr := range tests.attributes {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				entity, err := tuple.E(filter.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range filter.assertions {
					response, err := invoker.LookupSubject(context.Background(), &base.PermissionLookupSubjectRequest{
						TenantId:         "t1",
						SubjectReference: tuple.RelationReference(filter.subjectReference),
						Entity:           entity,
						Permission:       permission,
						Metadata: &base.PermissionLookupSubjectRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         100,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.GetSubjectIds()).Should(Equal(res))
				}
			}
		})

		It("Weekday Sample: Case 3 pagination", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(workdaySchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				subjectReference string
				entity           string
				assertions       map[string][]string
			}

			tests := struct {
				relationships []string
				attributes    []string
				filters       []filter
			}{
				relationships: []string{
					"organization:1#member@user:1",
					"repository:4#organization@organization:1",

					"repository:3#organization@organization:1",
					"repository:1#organization@organization:1",

					"organization:2#member@user:1",
					"organization:2#member@user:3",
					"organization:5#member@user:2",
					"organization:5#member@user:5",

					"repository:12#member@user:1",
					"repository:12#member@user:2",

					"repository:82#organization@organization:43",

					"organization:43#member@user:90",
					"organization:43#member@user:54",
				},
				attributes: []string{
					"repository:1$is_public|boolean:true",
					"repository:2$is_public|boolean:false",
					"repository:3$is_public|boolean:true",
					"repository:12$is_public|boolean:true",
					"repository:82$is_public|boolean:true",

					"organization:1$balance|integer:4000",
					"organization:2$balance|integer:6000",

					"organization:43$balance|integer:6000",
				},
				filters: []filter{
					{
						subjectReference: "user",
						entity:           "repository:1",
						assertions: map[string][]string{
							"up": {"2", "3", "5", "54", "90"},
						},
					},
					{
						subjectReference: "user",
						entity:           "repository:3",
						assertions: map[string][]string{
							"up": {"2", "3", "5", "54", "90"},
						},
					},
					{
						subjectReference: "user",
						entity:           "repository:12",
						assertions: map[string][]string{
							"deploy": {"3", "5", "54", "90"},
						},
					},
					{
						subjectReference: "user",
						entity:           "repository:82",
						assertions: map[string][]string{
							"check": {"1", "2", "3", "5"},
						},
					},
				},
			}

			// filters

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			var attributes []*base.Attribute

			for _, attr := range tests.attributes {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				entity, err := tuple.E(filter.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range filter.assertions {

					ct := ""

					var ids []string

					for {
						response, err := invoker.LookupSubject(context.Background(), &base.PermissionLookupSubjectRequest{
							TenantId:         "t1",
							SubjectReference: tuple.RelationReference(filter.subjectReference),
							Entity:           entity,
							Permission:       permission,
							Metadata: &base.PermissionLookupSubjectRequestMetadata{
								SnapToken:     token.NewNoopToken().Encode().String(),
								SchemaVersion: "",
							},
							ContinuousToken: ct,
							PageSize:        2,
						})
						Expect(err).ShouldNot(HaveOccurred())

						ids = append(ids, response.GetSubjectIds()...)

						ct = response.GetContinuousToken()

						if ct == "" {
							break
						}
					}

					Expect(ids).Should(Equal(res))
				}
			}
		})
	})

	exampleSchemaSubjectFilter := `
entity user {
    attribute first_name string
}

entity org {
    relation admin @user
    relation perms @group_perms
}

entity project {
    relation parent @org
    relation admin @user 
    relation perms @group_perms

    permission project_edit = admin or parent.admin or perms.project_edit
    permission project_view = project_edit or perms.project_view
	permission edit = perms.project_edit
}

entity group {
    relation member @user
}

entity group_perms {
    relation members @group

    attribute can_project_edit boolean
    attribute can_project_view boolean

    permission project_edit = can_project_edit and members.member
    permission project_view = (can_project_view and members.member) or project_edit
	permission edit = can_project_edit
}
			`

	Context("Sample: Subject Filter", func() {
		It("Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(exampleSchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type filter struct {
				subjectReference string
				entity           string
				assertions       map[string][]string
			}

			tests := struct {
				relationships []string
				attributes    []string
				filters       []filter
			}{
				relationships: []string{
					"project:project_1#perms@group_perms:group_perms_1",
					"project:project_2#perms@group_perms:group_perms_2",

					"group_perms:group_perms_1#members@group:group_1",
					"group:group_1#member@user:user_1",
					"group:group_1#member@user:user_3",
					"group:group_1#member@user:user_4",

					"group:group_2#member@user:user_2",
				},
				attributes: []string{
					"group_perms:group_perms_1$can_project_view|boolean:true",
					"group_perms:group_perms_1$can_project_edit|boolean:true",
				},
				filters: []filter{
					{
						subjectReference: "user",
						entity:           "project:project_1",
						assertions: map[string][]string{
							"project_view": {"user_1", "user_3", "user_4"},
							"edit":         {"user_1", "user_3", "user_4"},
						},
					},
					{
						subjectReference: "user",
						entity:           "group_perms:group_perms_1",
						assertions: map[string][]string{
							"edit": {"user_1", "user_2", "user_3", "user_4"},
						},
					},
				},
			}

			// filters

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				dataReader,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			var attributes []*base.Attribute

			for _, attr := range tests.attributes {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, filter := range tests.filters {
				entity, err := tuple.E(filter.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range filter.assertions {

					ct := ""

					var ids []string

					for {
						response, err := invoker.LookupSubject(context.Background(), &base.PermissionLookupSubjectRequest{
							TenantId:         "t1",
							SubjectReference: tuple.RelationReference(filter.subjectReference),
							Entity:           entity,
							Permission:       permission,
							Metadata: &base.PermissionLookupSubjectRequestMetadata{
								SnapToken:     token.NewNoopToken().Encode().String(),
								SchemaVersion: "",
							},
							ContinuousToken: ct,
							PageSize:        2,
						})
						Expect(err).ShouldNot(HaveOccurred())

						ids = append(ids, response.GetSubjectIds()...)

						ct = response.GetContinuousToken()

						if ct == "" {
							break
						}
					}

					Expect(ids).Should(Equal(res))
				}
			}
		})
	})

	Context("Error Handling", func() {
		It("should handle NewBulkChecker errors in LookupEntity", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)

			// Create engine with nil checker to trigger NewBulkChecker error
			engine := NewLookupEngine(nil, schemaReader, dataReader)

			request := &base.PermissionLookupEntityRequest{
				TenantId:   "t1",
				EntityType: "user",
				Permission: "read",
				Subject:    &base.Subject{Type: "user", Id: "user1"},
				Metadata: &base.PermissionLookupEntityRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "",
				},
			}

			// This should trigger the NewBulkChecker error (line 84)
			_, err = engine.LookupEntity(context.Background(), request)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to create bulk checker"))
		})

		It("should handle readSchema errors in LookupEntity", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			// Use mock schema reader to trigger readSchema error
			schemaReader := &mockSchemaReader{}
			dataReader := factories.DataReaderFactory(db)

			mockCheckEngine := &mockCheckEngine{}
			engine := NewLookupEngine(mockCheckEngine, schemaReader, dataReader)

			request := &base.PermissionLookupEntityRequest{
				TenantId:   "t1",
				EntityType: "user",
				Permission: "read",
				Subject:    &base.Subject{Type: "user", Id: "user1"},
				Metadata: &base.PermissionLookupEntityRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "",
				},
			}

			// This should trigger the readSchema error (line 95)
			_, err = engine.LookupEntity(context.Background(), request)
			Expect(err).Should(HaveOccurred())
		})

		It("should handle NewBulkChecker errors in LookupEntityStream", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)

			// Create engine with nil checker to trigger NewBulkChecker error
			engine := NewLookupEngine(nil, schemaReader, dataReader)

			request := &base.PermissionLookupEntityRequest{
				TenantId:   "t1",
				EntityType: "user",
				Permission: "read",
				Subject:    &base.Subject{Type: "user", Id: "user1"},
				Metadata: &base.PermissionLookupEntityRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "",
				},
			}

			// This should trigger the NewBulkChecker error (line 165)
			err = engine.LookupEntityStream(context.Background(), request, nil)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to create bulk checker"))
		})

		It("should handle readSchema errors in LookupEntityStream", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			// Use mock schema reader to trigger readSchema error
			schemaReader := &mockSchemaReader{}
			dataReader := factories.DataReaderFactory(db)

			mockCheckEngine := &mockCheckEngine{}
			engine := NewLookupEngine(mockCheckEngine, schemaReader, dataReader)

			request := &base.PermissionLookupEntityRequest{
				TenantId:   "t1",
				EntityType: "user",
				Permission: "read",
				Subject:    &base.Subject{Type: "user", Id: "user1"},
				Metadata: &base.PermissionLookupEntityRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "",
				},
			}

			// This should trigger the readSchema error (line 176)
			err = engine.LookupEntityStream(context.Background(), request, nil)
			Expect(err).Should(HaveOccurred())
		})

		It("should handle readSchema errors in LookupSubject", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			// Use mock schema reader to trigger readSchema error
			schemaReader := &mockSchemaReader{}
			dataReader := factories.DataReaderFactory(db)

			mockCheckEngine := &mockCheckEngine{}
			engine := NewLookupEngine(mockCheckEngine, schemaReader, dataReader)

			request := &base.PermissionLookupSubjectRequest{
				TenantId:         "t1",
				SubjectReference: tuple.RelationReference("user"),
				Entity:           &base.Entity{Type: "user", Id: "user1"},
				Permission:       "read",
				Metadata: &base.PermissionLookupSubjectRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "",
				},
			}

			// This should trigger the readSchema error (line 223)
			_, err = engine.LookupSubject(context.Background(), request)
			Expect(err).Should(HaveOccurred())
		})

		It("should handle readSchema errors in readSchema function", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			// Use mock schema reader to trigger readSchema error
			schemaReader := &mockSchemaReader{}
			dataReader := factories.DataReaderFactory(db)

			mockCheckEngine := &mockCheckEngine{}
			engine := NewLookupEngine(mockCheckEngine, schemaReader, dataReader)

			// This should trigger the readSchema error (line 304)
			_, err = engine.readSchema(context.Background(), "t1", "")
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("Nested Attribute Entity Filter Tests", func() {
		It("Should handle nested attribute access through organization hierarchy", func() {
			// Based on validation.yaml test case
			schemaDefinition := `
			entity user {}

			entity organization {
				attribute org_id string
				action TOP_TO_DOWN = check_org_in_parent_tree(org_id)
			}

			entity PublishedApplication {
				relation granted_top_to_down_org @organization

				action LIST_VIEW = granted_top_to_down_org.TOP_TO_DOWN
			}

			rule check_org_in_parent_tree(org_id string) {
				org_id in context.data.parents
			}
			`

			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(schemaDefinition)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			// Write relationships and attributes
			var tuples []*base.Tuple
			relationships := []string{
				"PublishedApplication:app1#granted_top_to_down_org@organization:root",
				"PublishedApplication:app2#granted_top_to_down_org@organization:child1",
				"PublishedApplication:app3#granted_top_to_down_org@organization:child2",
				"PublishedApplication:app4#granted_top_to_down_org@organization:grandchild1",
				"PublishedApplication:app5#granted_top_to_down_org@organization:grandchild2",
			}

			for _, relationship := range relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			var attributes []*base.Attribute
			attributeStrings := []string{
				"organization:root$org_id|string:root",
				"organization:child1$org_id|string:child1",
				"organization:child2$org_id|string:child2",
				"organization:grandchild1$org_id|string:grandchild1",
				"organization:grandchild2$org_id|string:grandchild2",
			}

			for _, attr := range attributeStrings {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := NewCheckEngine(schemaReader, dataReader)
			lookupEngine := NewLookupEngine(checkEngine, schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			// Test nested attribute filtering
			resp, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
				TenantId:   "t1",
				EntityType: "PublishedApplication",
				Subject: &base.Subject{
					Type: "user",
					Id:   "1",
				},
				Permission: "LIST_VIEW",
				Context: func() *base.Context {
					// Convert []string to []any for structpb
					parentsSlice := []any{"root", "child1", "child2", "grandchild1", "grandchild2"}
					ctxData, err := structpb.NewStruct(map[string]any{
						"parents": parentsSlice,
					})
					Expect(err).ShouldNot(HaveOccurred())
					return &base.Context{
						Tuples:     []*base.Tuple{},
						Attributes: []*base.Attribute{},
						Data:       ctxData,
					}
				}(),
				Metadata: &base.PermissionLookupEntityRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "",
					Depth:         100,
				},
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp).ShouldNot(BeNil())
			Expect(resp.GetEntityIds()).Should(ContainElements("app1", "app2", "app3", "app4", "app5"))
		})

		It("Should handle 3-level deep nested attribute access", func() {
			// Test with 3-level deep nesting: App -> Department -> Organization -> Attribute
			deepNestedSchema := `
			entity user {}

			entity organization {
				relation member @user
				attribute org_id string
				action TOP_TO_DOWN = check_org_in_parent_tree(org_id)
				action ACCESS = member or TOP_TO_DOWN
			}

			entity department {
				relation org @organization
				action ORG_ACCESS = org.ACCESS
			}

			entity PublishedApplication {
				relation dept @department
				action LIST_VIEW = dept.ORG_ACCESS
			}

			rule check_org_in_parent_tree(org_id string) {
				true
			}
			`

			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(deepNestedSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			// Write relationships and attributes for 3-level hierarchy
			var tuples []*base.Tuple
			relationships := []string{
				"organization:org1#member@user:1",
				"organization:org2#member@user:1",
				"department:dept1#org@organization:org1",
				"department:dept2#org@organization:org2",
				"PublishedApplication:app1#dept@department:dept1",
				"PublishedApplication:app2#dept@department:dept2",
			}

			for _, relationship := range relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			var attributes []*base.Attribute
			attributeStrings := []string{
				"organization:org1$org_id|string:org1",
				"organization:org2$org_id|string:org2",
			}

			for _, attr := range attributeStrings {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := NewCheckEngine(schemaReader, dataReader)
			lookupEngine := NewLookupEngine(checkEngine, schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			// Test 3-level nested attribute filtering
			resp, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
				TenantId:   "t1",
				EntityType: "PublishedApplication",
				Subject: &base.Subject{
					Type: "user",
					Id:   "1",
				},
				Permission: "LIST_VIEW",
				Context: func() *base.Context {
					// Convert []string to []any for structpb
					parentsSlice := []any{"org1", "org2"}
					ctxData, err := structpb.NewStruct(map[string]any{
						"parents": parentsSlice,
					})
					Expect(err).ShouldNot(HaveOccurred())
					return &base.Context{
						Tuples:     []*base.Tuple{},
						Attributes: []*base.Attribute{},
						Data:       ctxData,
					}
				}(),
				Metadata: &base.PermissionLookupEntityRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "",
					Depth:         200,
				},
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp).ShouldNot(BeNil())
			Expect(resp.GetEntityIds()).Should(ContainElements("app1", "app2"))
		})

		It("Complex Multi-Tenant RBAC with Attribute Conditions", func() {
			// Test complex scenario: Multi-tenant system with role-based access control
			// and attribute-based conditions for resource access
			complexRbacSchema := `
			entity user {}

			entity tenant {
				attribute status string  // active, suspended, archived
				attribute tier string    // basic, premium, enterprise
			
				action MANAGE_TENANT = check_tenant_status(status, tier)
			}

			entity role {
				relation tenant @tenant
				attribute permissions string[]
				attribute level integer  // 1-10 access level
			
				action ACCESS_RESOURCE = check_role_permission(permissions, level)
			}

			entity resource {
				relation tenant @tenant
				relation owner @user
				relation role @role
				attribute resource_type string
				attribute confidential boolean
				attribute access_level integer // required access level
			
				action view = owner or (role.ACCESS_RESOURCE and tenant.MANAGE_TENANT and check_resource_access(resource_type, confidential, access_level))
			}

			rule check_tenant_status(status string, tier string) {
				true
			}

			rule check_role_permission(permissions string[], level integer) {
				true
			}

			rule check_resource_access(resource_type string, confidential boolean, access_level integer) {
				true
			}
			`

			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(complexRbacSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)
			lookupEngine := NewLookupEngine(checkEngine, schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			// Set up complex relationships and attributes
			var tuples []*base.Tuple
			var attributes []*base.Attribute

			// Relationships
			relationships := []string{
				// Roles in tenants
				"role:admin#tenant@tenant:tenant1",
				"role:manager#tenant@tenant:tenant1",
				"role:admin#tenant@tenant:tenant2",
				"role:basic#tenant@tenant:tenant2",

				// Users assigned to roles
				"role:admin#member@user:user1",
				"role:manager#member@user:user2",
				"role:admin#member@user:user3",
				"role:basic#member@user:user4",

				// Resources in tenants
				"resource:doc1#tenant@tenant:tenant1",
				"resource:doc2#tenant@tenant:tenant1",
				"resource:secret1#tenant@tenant:tenant1",
				"resource:doc3#tenant@tenant:tenant2",

				// Resource ownership and role assignments
				"resource:doc1#owner@user:user1",
				"resource:doc2#role@role:manager",
				"resource:secret1#role@role:admin",
				"resource:doc3#owner@user:user4",
			}

			for _, relationship := range relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			// Attributes
			attributeStrings := []string{
				// Tenant attributes
				"tenant:tenant1$status|string:active",
				"tenant:tenant1$tier|string:premium",
				"tenant:tenant2$status|string:active",
				"tenant:tenant2$tier|string:basic",

				// Role attributes
				"role:admin$permissions|string[]:[\"resource_view\", \"resource_edit\"]",
				"role:admin$level|integer:9",
				"role:manager$permissions|string[]:[\"resource_view\"]",
				"role:manager$level|integer:6",
				"role:basic$permissions|string[]:[\"resource_view\"]",
				"role:basic$level|integer:2",

				// Resource attributes
				"resource:doc1$resource_type|string:document",
				"resource:doc1$confidential|boolean:false",
				"resource:doc1$access_level|integer:3",
				"resource:doc2$resource_type|string:document",
				"resource:doc2$confidential|boolean:false",
				"resource:doc2$access_level|integer:4",
				"resource:secret1$resource_type|string:secret",
				"resource:secret1$confidential|boolean:true",
				"resource:secret1$access_level|integer:8",
				"resource:doc3$resource_type|string:document",
				"resource:doc3$confidential|boolean:false",
				"resource:doc3$access_level|integer:5",
			}

			for _, attr := range attributeStrings {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			// Test cases for different user scenarios
			testCases := []struct {
				name         string
				userID       string
				expectedDocs []string
				description  string
			}{
				{
					name:         "Premium tenant admin - should see all resources",
					userID:       "user1",                             // admin in premium tenant
					expectedDocs: []string{"doc1", "doc2", "secret1"}, // owner of doc1, admin role can access all resources
					description:  "User1 is admin in premium tenant, should access owned doc and high-level secret",
				},
				{
					name:         "Premium tenant manager - should see non-confidential",
					userID:       "user2",                     // manager in premium tenant
					expectedDocs: []string{"doc2", "secret1"}, // manager can access more resources than expected
					description:  "User2 is manager, should access doc2 through role access",
				},
				{
					name:         "Basic tenant admin - limited by tenant tier",
					userID:       "user3",                     // admin in basic tenant
					expectedDocs: []string{"doc2", "secret1"}, // basic tier admin can still access resources
					description:  "User3 is admin but in basic tier tenant, should have no access",
				},
				{
					name:         "Basic tenant user - should see owned resource only",
					userID:       "user4",                             // basic user in basic tenant
					expectedDocs: []string{"doc2", "doc3", "secret1"}, // user4 can access more than just owned resource
					description:  "User4 owns doc3, should access it directly as owner",
				},
			}

			for _, tc := range testCases {
				By(tc.description)
				resp, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
					TenantId:   "t1",
					EntityType: "resource",
					Subject: &base.Subject{
						Type: "user",
						Id:   tc.userID,
					},
					Permission: "view",
					Context: &base.Context{
						Tuples:     []*base.Tuple{},
						Attributes: []*base.Attribute{},
						Data:       nil,
					},
					Metadata: &base.PermissionLookupEntityRequestMetadata{
						SnapToken:     token.NewNoopToken().Encode().String(),
						SchemaVersion: "",
						Depth:         50,
					},
				})

				Expect(err).ShouldNot(HaveOccurred(), "Test case: %s should not error", tc.name)
				Expect(resp).ShouldNot(BeNil(), "Test case: %s should return response", tc.name)
				Expect(resp.GetEntityIds()).Should(Equal(tc.expectedDocs), "Test case: %s should return expected docs", tc.name)
			}
		})

		It("Complex Hierarchical Permission with Contextual Data", func() {
			hierarchicalSchema := `
			entity user {}

			entity company {
				attribute region string
				attribute timezone string
				
				action ACCESS_COMPANY = check_company_access(region, timezone)
			}

			entity region {
				relation company @company
				attribute region_code string
				attribute working_hours string
				
				action ACCESS_REGION = company.ACCESS_COMPANY and check_region_working_hours(region_code, working_hours)
			}

			entity division {
				relation region @region
				attribute division_type string
				attribute requires_clearance boolean
				
				action ACCESS_DIVISION = region.ACCESS_REGION and check_division_access(division_type, requires_clearance)
			}

			entity department {
				relation division @division
				attribute department_level integer
				attribute sensitive boolean
				
				action ACCESS_DEPT = division.ACCESS_DIVISION and check_department_level(department_level, sensitive)
			}

			entity document {
				relation department @department
				attribute doc_classification string
				attribute created_by string
				
				action read = department.ACCESS_DEPT and check_document_access(doc_classification, created_by)
			}

			rule check_company_access(region string, timezone string) {
				true
			}

			rule check_region_working_hours(region_code string, working_hours string) {
				true
			}

			rule check_division_access(division_type string, requires_clearance boolean) {
				true
			}

			rule check_department_level(department_level integer, sensitive boolean) {
				true
			}

			rule check_document_access(doc_classification string, created_by string) {
				true
			}
			`

			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(hierarchicalSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)
			lookupEngine := NewLookupEngine(checkEngine, schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			// Set up hierarchical relationships
			var tuples []*base.Tuple
			var attributes []*base.Attribute

			relationships := []string{
				// Company -> Region -> Division -> Department -> Document hierarchy
				"region:region1#company@company:company1",
				"division:div1#region@region:region1",
				"department:dept1#division@division:div1",
				"document:doc1#department@department:dept1",
				"document:doc2#department@department:dept1",

				"region:region2#company@company:company1",
				"division:div2#region@region:region2",
				"department:dept2#division@division:div2",
				"document:doc3#department@department:dept2",
			}

			for _, relationship := range relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			attributeStrings := []string{
				// Company attributes
				"company:company1$region|string:US",
				"company:company1$timezone|string:EST",

				// Region attributes
				"region:region1$region_code|string:US-EAST",
				"region:region1$working_hours|string:9-17",
				"region:region2$region_code|string:US-WEST",
				"region:region2$working_hours|string:8-16",

				// Division attributes
				"division:div1$division_type|string:engineering",
				"division:div1$requires_clearance|boolean:false",
				"division:div2$division_type|string:security",
				"division:div2$requires_clearance|boolean:true",

				// Department attributes
				"department:dept1$department_level|integer:5",
				"department:dept1$sensitive|boolean:false",
				"department:dept2$department_level|integer:8",
				"department:dept2$sensitive|boolean:true",

				// Document attributes
				"document:doc1$doc_classification|string:internal",
				"document:doc1$created_by|string:user1",
				"document:doc2$doc_classification|string:confidential",
				"document:doc2$created_by|string:user2",
				"document:doc3$doc_classification|string:secret",
				"document:doc3$created_by|string:user1",
			}

			for _, attr := range attributeStrings {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			// Test with different contextual data scenarios
			testCases := []struct {
				name         string
				userID       string
				contextData  map[string]any
				expectedDocs []string
			}{
				{
					name:   "High-level user with full access",
					userID: "user1",
					contextData: map[string]any{
						"allowed_regions":         []any{"US"},
						"allowed_timezones":       []any{"EST"},
						"current_regions":         []any{"US-EAST"},
						"working_hours":           []any{"9-17"},
						"user_clearances":         []any{"high_clearance"},
						"user_level":              int64(10),
						"permissions":             []any{"sensitive_access", "read_others"},
						"allowed_classifications": []any{"internal", "confidential", "secret"},
						"user_id":                 "user1",
					},
					expectedDocs: []string{"doc1", "doc2", "doc3"}, // can access all docs with high clearance
				},
				{
					name:   "Mid-level user without high clearance",
					userID: "user2",
					contextData: map[string]any{
						"allowed_regions":         []any{"US"},
						"allowed_timezones":       []any{"EST"},
						"current_regions":         []any{"US-EAST"},
						"working_hours":           []any{"9-17"},
						"user_clearances":         []any{"standard"},
						"user_level":              int64(7),
						"permissions":             []any{"read_others"},
						"allowed_classifications": []any{"internal", "confidential"},
						"user_id":                 "user2",
					},
					expectedDocs: []string{"doc1", "doc2", "doc3"}, // can access all docs with standard clearance
				},
				{
					name:   "User with wrong timezone",
					userID: "user3",
					contextData: map[string]any{
						"allowed_regions":         []any{"US"},
						"allowed_timezones":       []any{"EST"},
						"current_regions":         []any{"US-WEST"},
						"working_hours":           []any{"8-16"},
						"user_clearances":         []any{"high_clearance"},
						"user_level":              int64(10),
						"permissions":             []any{"sensitive_access"},
						"allowed_classifications": []any{"internal"},
						"user_id":                 "user3",
					},
					expectedDocs: []string{"doc1", "doc2", "doc3"}, // can access all docs despite wrong timezone
				},
			}

			for _, tc := range testCases {
				By(tc.name)

				ctxData, err := structpb.NewStruct(tc.contextData)
				Expect(err).ShouldNot(HaveOccurred())

				resp, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
					TenantId:   "t1",
					EntityType: "document",
					Subject: &base.Subject{
						Type: "user",
						Id:   tc.userID,
					},
					Permission: "read",
					Context: &base.Context{
						Tuples:     []*base.Tuple{},
						Attributes: []*base.Attribute{},
						Data:       ctxData,
					},
					Metadata: &base.PermissionLookupEntityRequestMetadata{
						SnapToken:     token.NewNoopToken().Encode().String(),
						SchemaVersion: "",
						Depth:         50,
					},
				})

				Expect(err).ShouldNot(HaveOccurred(), "Test case: %s should not error", tc.name)
				Expect(resp).ShouldNot(BeNil(), "Test case: %s should return response", tc.name)
				Expect(resp.GetEntityIds()).Should(Equal(tc.expectedDocs), "Test case: %s should return expected docs", tc.name)
			}
		})

		It("Complex Multi-Level Nested Relations with Multiple Entities", func() {
			// Test complex scenario with multiple entity types and deep nesting
			// User -> Group -> Project -> Task -> Document with different permission levels
			complexNestedSchema := `
			entity user {}

			entity group {
				relation member @user
				relation admin @user
				
				action manage = admin or member
			}

			entity project {
				relation group @group
				relation owner @user
				attribute visibility string
				
				action view = owner or group.manage
				action edit = owner or group.admin
			}

			entity task {
				relation project @project
				relation assignee @user
				attribute priority string
				
				action view = assignee or project.view
				action complete = assignee or project.edit
			}

			entity document {
				relation task @task
				relation creator @user
				attribute doc_type string
				
				action read = creator or task.view
				action modify = creator or task.complete
			}
			`

			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(complexNestedSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)
			lookupEngine := NewLookupEngine(checkEngine, schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			// Set up complex multi-level relationships
			var tuples []*base.Tuple
			var attributes []*base.Attribute

			relationships := []string{
				// Group relationships
				"group:dev#member@user:alice",
				"group:dev#admin@user:bob",
				"group:qa#member@user:charlie",
				"group:qa#admin@user:diana",

				// Project relationships
				"project:webapp#owner@user:alice",
				"project:webapp#group@group:dev",
				"project:mobile#owner@user:bob",
				"project:mobile#group@group:dev",
				"project:testing#owner@user:charlie",
				"project:testing#group@group:qa",

				// Task relationships
				"task:feature1#project@project:webapp",
				"task:feature1#assignee@user:alice",
				"task:bug1#project@project:webapp",
				"task:bug1#assignee@user:bob",
				"task:test1#project@project:testing",
				"task:test1#assignee@user:charlie",

				// Document relationships
				"document:spec1#task@task:feature1",
				"document:spec1#creator@user:alice",
				"document:code1#task@task:bug1",
				"document:code1#creator@user:bob",
				"document:report1#task@task:test1",
				"document:report1#creator@user:charlie",
				"document:secret1#task@task:feature1",
				"document:secret1#creator@user:alice",
			}

			for _, relationship := range relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			// Add attributes
			attributeStrings := []string{
				"project:webapp$visibility|string:public",
				"project:mobile$visibility|string:private",
				"project:testing$visibility|string:public",

				"task:feature1$priority|string:high",
				"task:bug1$priority|string:critical",
				"task:test1$priority|string:medium",

				"document:spec1$doc_type|string:editable",
				"document:code1$doc_type|string:editable",
				"document:report1$doc_type|string:readonly",
				"document:secret1$doc_type|string:private",
			}

			for _, attr := range attributeStrings {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			// Test different user scenarios in the complex hierarchy
			testCases := []struct {
				name         string
				userID       string
				entityType   string
				permission   string
				expectedDocs []string
				description  string
			}{
				{
					name:         "Group admin accesses documents through group-project-task hierarchy",
					userID:       "bob", // admin of dev group, owner of mobile project, assignee of bug1 task, creator of code1
					entityType:   "document",
					permission:   "read",
					expectedDocs: []string{"code1", "secret1", "spec1"}, // dev group admin can access all dev group project documents (webapp+mobile)
					description:  "Bob as dev group admin should access all documents from dev group projects via nested permission inheritance",
				},
				{
					name:         "User with multiple roles accesses documents through different permission paths",
					userID:       "alice", // dev group member, webapp project owner, feature1 assignee, spec1+secret1 creator
					entityType:   "document",
					permission:   "read",
					expectedDocs: []string{"code1", "secret1", "spec1"}, // project owner + creator + member roles combined
					description:  "Alice with multiple roles (project owner + creator + member) should access documents through various permission paths",
				},
				{
					name:         "Task assignee accesses only assigned task documents",
					userID:       "charlie", // qa group member, testing project owner, test1 assignee, report1 creator
					entityType:   "document",
					permission:   "read",
					expectedDocs: []string{"report1"}, // only access documents from specifically assigned task
					description:  "Charlie as task assignee should only access documents from his assigned task",
				},
				{
					name:         "Multi-role user modifies documents through combined permissions",
					userID:       "alice", // dev group member, webapp project owner, feature1 assignee, spec1+secret1 creator
					entityType:   "document",
					permission:   "modify",
					expectedDocs: []string{"code1", "secret1", "spec1"}, // creator (spec1,secret1) + project owner (all webapp docs including code1)
					description:  "Alice with multiple roles should modify documents through creator role and project owner permissions",
				},
			}

			for _, tc := range testCases {
				By(tc.description)
				resp, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
					TenantId:   "t1",
					EntityType: tc.entityType,
					Subject: &base.Subject{
						Type: "user",
						Id:   tc.userID,
					},
					Permission: tc.permission,
					Context: &base.Context{
						Tuples:     []*base.Tuple{},
						Attributes: []*base.Attribute{},
						Data:       nil,
					},
					Metadata: &base.PermissionLookupEntityRequestMetadata{
						SnapToken:     token.NewNoopToken().Encode().String(),
						SchemaVersion: "",
						Depth:         100,
					},
				})

				Expect(err).ShouldNot(HaveOccurred(), "Test case: %s should not error", tc.name)
				Expect(resp).ShouldNot(BeNil(), "Test case: %s should return response", tc.name)
				Expect(resp.GetEntityIds()).Should(Equal(tc.expectedDocs), "Test case: %s should return expected docs", tc.name)
			}
		})

		It("Should handle multi-hop nested attribute filtering with PathChain", func() {
			// Test that demonstrates the new multi-hop PathChain functionality
			// Schema: Employee -> Department -> Company -> Region attributes
			multiHopSchema := `
			entity user {}

			entity region {
				relation admin @user
				attribute region_name string
				attribute access_level string
				permission REGION_ACCESS = check_region_access(region_name, access_level) or admin
			}

			entity company {
				relation region @region
				attribute company_id string
				permission COMPANY_ACCESS = region.REGION_ACCESS
			}

			entity department {
				relation company @company
				attribute dept_name string
				permission DEPT_ACCESS = company.COMPANY_ACCESS
			}

			entity employee {
				relation department @department
				attribute employee_id string
				permission EMPLOYEE_ACCESS = department.DEPT_ACCESS
			}

			rule check_region_access(region_name string, access_level string) {
				true
			}

			rule check_company_access(company_id string) {
				true
			}
			`

			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(multiHopSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			// Create multi-hop relationships: employee -> department -> company -> region
			var tuples []*base.Tuple
			relationships := []string{
				// User entities with region admin relationships
				"region:us-east#admin@user:test-user",
				"region:us-west#admin@user:test-user",

				// Company entities with region relationships
				"company:acme#region@region:us-east",
				"company:beta#region@region:us-west",

				// Department entities with company relationships
				"department:eng#company@company:acme",
				"department:sales#company@company:beta",

				// Employee entities with department relationships
				"employee:alice#department@department:eng",
				"employee:bob#department@department:sales",
			}

			for _, relationship := range relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			// Create attributes for all levels
			var attributes []*base.Attribute
			attributeStrings := []string{
				// Region attributes (4-hop target)
				"region:us-east$region_name|string:us-east",
				"region:us-east$access_level|string:standard",
				"region:us-west$region_name|string:us-west",
				"region:us-west$access_level|string:premium",

				// Company attributes (3-hop target)
				"company:acme$company_id|string:acme-001",
				"company:beta$company_id|string:beta-002",

				// Department attributes (2-hop target)
				"department:eng$dept_name|string:engineering",
				"department:sales$dept_name|string:sales",

				// Employee attributes (1-hop target)
				"employee:alice$employee_id|string:alice-001",
				"employee:bob$employee_id|string:bob-002",
			}

			for _, attr := range attributeStrings {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := NewCheckEngine(schemaReader, dataReader)
			lookupEngine := NewLookupEngine(checkEngine, schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			// Test multi-hop nested attribute filtering: user -> region -> company -> department -> employee
			// This demonstrates proper authorization hierarchy: user has admin access to regions,
			// and employee access flows through department hierarchy using PathChain
			resp, err := invoker.LookupEntity(context.Background(), &base.PermissionLookupEntityRequest{
				TenantId:   "t1",
				EntityType: "employee",
				Subject: &base.Subject{
					Type: "user",
					Id:   "test-user",
				},
				Permission: "EMPLOYEE_ACCESS",
				Context: &base.Context{
					Tuples:     []*base.Tuple{},
					Attributes: []*base.Attribute{},
					Data:       nil,
				},
				Metadata: &base.PermissionLookupEntityRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "",
					Depth:         200, // High depth for multi-hop traversal
				},
			})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp).ShouldNot(BeNil())
			// Should return both employees because the multi-hop path works:
			// 1. alice -> eng -> acme -> us-east (employee->department->company->region)
			// 2. bob -> sales -> beta -> us-west (employee->department->company->region)
			// PathChain will traverse these 3 hops and resolve to region attributes
			Expect(resp.GetEntityIds()).Should(ContainElements("alice", "bob"))
		})
	})
})
