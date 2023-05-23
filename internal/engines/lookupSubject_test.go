package engines

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
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

	relation owner @user
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

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			lookupSubjectEngine := NewLookupSubjectEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				nil,
				nil,
				nil,
				lookupSubjectEngine,
			)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
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
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			lookupSubjectEngine := NewLookupSubjectEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				nil,
				nil,
				nil,
				lookupSubjectEngine,
			)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
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
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			lookupSubjectEngine := NewLookupSubjectEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				nil,
				nil,
				nil,
				lookupSubjectEngine,
			)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
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
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			lookupSubjectEngine := NewLookupSubjectEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				nil,
				nil,
				nil,
				lookupSubjectEngine,
			)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
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
})
