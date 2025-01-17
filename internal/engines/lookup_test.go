package engines

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

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
							Depth:         100,
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
							Depth:         100,
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
})
