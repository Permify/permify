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
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/telemetry"
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetEntityIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetEntityIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetEntityIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetEntityIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetEntityIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetEntityIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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
							"read": {"1", "2", "3", "4", "7", "10"},
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
							"read":   {"2", "3", "5", "6", "9", "10"},
							"update": {"5"},
						},
					},
					{
						entityType: "doc",
						subject:    "user:4",
						assertions: map[string][]string{
							"read":   {"7", "8"},
							"delete": {"7", "8"},
							"share":  {},
						},
					},
					{
						entityType: "doc",
						subject:    "user:5",
						assertions: map[string][]string{
							"read":   {"9", "10"},
							"delete": {"9", "10"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetEntityIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetEntityIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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
							"upload_file": {},
							"view_file":   {},
							"delete_file": {"1", "2", "3"},
						},
					},
					{
						entityType: "event",
						subject:    "user:6",
						assertions: map[string][]string{
							"create_event":  {},
							"view_event":    {},
							"edit_event":    {},
							"delete_event":  {},
							"RSVP_to_event": {},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetEntityIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetEntityIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetEntityIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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
						"view": {"2", "36", "22", "47", "1", "23", "16", "43", "24", "41", "13", "26", "10", "45", "11", "37", "20", "46", "99", "38", "12", "25", "40", "49", "17", "15", "4", "27", "83", "48", "5", "57", "75", "29", "39", "77", "98", "97", "81", "78", "76", "80", "3", "21", "65", "66", "53", "44", "34", "71", "58", "14", "74", "72", "70", "51", "33", "84", "67", "50", "63", "68", "8", "6", "69", "62", "73", "79", "91", "88", "35", "100", "64", "30", "7", "19", "18", "93", "32", "56", "55", "86", "59", "90", "60", "9", "87", "89", "31", "95", "52", "92", "61", "94", "28", "42", "85", "82", "54", "96"},
						"edit": {"57", "67", "64", "51", "94", "91", "77", "69", "81", "66", "84", "86", "70", "79", "59", "73", "76", "80", "78", "74", "93", "61", "92", "97", "96", "72", "71", "98", "99", "65", "95", "54", "89", "88", "63", "53", "82", "100", "85", "56", "87", "75", "90", "52", "58", "55", "68", "60", "62", "83"},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetEntityIds(), res)).Should(Equal(true))
				}
			}
		})
	})

	weekdaySchemaEntityFilter := `
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
			permission delete = is_weekday(request.day_of_week)
		}
		
		rule check_balance(balance integer) {
			balance > 5000
		}

		rule is_weekday(day_of_week string) {
			  day_of_week != 'saturday' && day_of_week != 'sunday'
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

			conf, err := newSchema(weekdaySchemaEntityFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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
					"repository:4#organization@organization:1",

					"organization:2#member@user:1",
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
						entityType: "repository",
						subject:    "user:1",
						assertions: map[string][]string{
							"view": {"1", "3"},
						},
					},
					{
						entityType: "organization",
						subject:    "user:1",
						assertions: map[string][]string{
							"view": {"2"},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetEntityIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetSubjectIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetSubjectIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetSubjectIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetSubjectIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetSubjectIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetSubjectIds(), res)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetSubjectIds(), res)).Should(Equal(true))
				}
			}
		})
	})

	weekdaySchemaSubjectFilter := `
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
			permission delete = is_weekday(request.day_of_week)
		}
		
		rule check_balance(balance integer) {
			balance > 5000
		}

		rule is_weekday(day_of_week string) {
			  day_of_week != 'saturday' && day_of_week != 'sunday'
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

			conf, err := newSchema(weekdaySchemaSubjectFilter)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			dataReader := factories.DataReaderFactory(db, logger.New("debug"))
			dataWriter := factories.DataWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			schemaBasedEntityFilter := NewSchemaBasedEntityFilter(dataReader)
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)

			massEntityFilter := NewMassEntityFilter(dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)

			lookupEngine := NewLookupEngine(
				checkEngine,
				schemaReader,
				schemaBasedEntityFilter,
				massEntityFilter,
				schemaBasedSubjectFilter,
				massSubjectFilter,
			)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				lookupEngine,
				nil,
				telemetry.NewNoopMeter(),
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
					Expect(isSameArray(response.GetSubjectIds(), res)).Should(Equal(true))
				}
			}
		})
	})
})
