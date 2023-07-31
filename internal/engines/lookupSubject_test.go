package engines

import (
	"context"

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

var _ = Describe("lookup-subject-engine", func() {
	driveSchema := `
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
}
`

	Context("Drive Sample: Subject Filter", func() {
		It("Drive Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchema)
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
			schemaBasedEntityFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)
			massEntityFilter := NewMassSubjectFilter(dataReader)
			lookupSubjectEngine := NewLookupSubjectEngine(checkEngine, schemaReader, schemaBasedEntityFilter, massEntityFilter)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				lookupSubjectEngine,
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

			conf, err := newSchema(driveSchema)
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
			schemaBasedEntityFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)
			massEntityFilter := NewMassSubjectFilter(dataReader)
			lookupSubjectEngine := NewLookupSubjectEngine(checkEngine, schemaReader, schemaBasedEntityFilter, massEntityFilter)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				lookupSubjectEngine,
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

			conf, err := newSchema(driveSchema)
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
			schemaBasedEntityFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)
			massEntityFilter := NewMassSubjectFilter(dataReader)
			lookupSubjectEngine := NewLookupSubjectEngine(checkEngine, schemaReader, schemaBasedEntityFilter, massEntityFilter)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				lookupSubjectEngine,
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

			conf, err := newSchema(driveSchema)
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
			schemaBasedEntityFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)
			massEntityFilter := NewMassSubjectFilter(dataReader)
			lookupSubjectEngine := NewLookupSubjectEngine(checkEngine, schemaReader, schemaBasedEntityFilter, massEntityFilter)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				lookupSubjectEngine,
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

			conf, err := newSchema(driveSchema)
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
			schemaBasedEntityFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)
			massEntityFilter := NewMassSubjectFilter(dataReader)
			lookupSubjectEngine := NewLookupSubjectEngine(checkEngine, schemaReader, schemaBasedEntityFilter, massEntityFilter)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				lookupSubjectEngine,
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

			conf, err := newSchema(driveSchema)
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
			schemaBasedEntityFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)
			massEntityFilter := NewMassSubjectFilter(dataReader)
			lookupSubjectEngine := NewLookupSubjectEngine(checkEngine, schemaReader, schemaBasedEntityFilter, massEntityFilter)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				lookupSubjectEngine,
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

			conf, err := newSchema(driveSchema)
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
			schemaBasedEntityFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)
			massEntityFilter := NewMassSubjectFilter(dataReader)
			lookupSubjectEngine := NewLookupSubjectEngine(checkEngine, schemaReader, schemaBasedEntityFilter, massEntityFilter)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				lookupSubjectEngine,
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

	weekdaySchema := `
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

			conf, err := newSchema(weekdaySchema)
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
					"repository:1#is_public@boolean:true",
					"repository:2#is_public@boolean:false",
					"repository:3#is_public@boolean:true",

					"organization:1#balance@integer:4000",
					"organization:2#balance@integer:6000",
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
			schemaBasedSubjectFilter := NewSchemaBasedSubjectFilter(schemaReader, dataReader)
			massSubjectFilter := NewMassSubjectFilter(dataReader)
			lookupSubjectEngine := NewLookupSubjectEngine(checkEngine, schemaReader, schemaBasedSubjectFilter, massSubjectFilter)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				lookupSubjectEngine,
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
