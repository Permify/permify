package engines

import (
	"context"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("subject-permission-engine", func() {
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

	Context("Drive Sample: Subject Permission", func() {
		It("Drive Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type assertion struct {
				onlyPermission bool
				subject        string
				entity         string
				result         map[string]base.CheckResult
			}

			tests := struct {
				relationships []string
				assertions    []assertion
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:1#folder@user:3",
					"folder:1#collaborator@user:1",
					"folder:1#collaborator@user:3",
					"organization:1#admin@user:1",
					"doc:1#org@organization:1#...",
				},
				assertions: []assertion{
					{
						subject:        "user:1",
						entity:         "doc:1",
						onlyPermission: false,
						result: map[string]base.CheckResult{
							// relations
							"org":    base.CheckResult_CHECK_RESULT_DENIED,
							"parent": base.CheckResult_CHECK_RESULT_DENIED,
							"owner":  base.CheckResult_CHECK_RESULT_DENIED,

							// permissions
							"read":   base.CheckResult_CHECK_RESULT_ALLOWED,
							"update": base.CheckResult_CHECK_RESULT_DENIED,
							"delete": base.CheckResult_CHECK_RESULT_ALLOWED,
							"share":  base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
					{
						subject:        "user:1",
						entity:         "doc:1",
						onlyPermission: true,
						result: map[string]base.CheckResult{
							"read":   base.CheckResult_CHECK_RESULT_ALLOWED,
							"update": base.CheckResult_CHECK_RESULT_DENIED,
							"delete": base.CheckResult_CHECK_RESULT_ALLOWED,
							"share":  base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			subjectPermissionEngine := NewSubjectPermission(checkEngine, schemaReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				subjectPermissionEngine,
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

			for _, assertion := range tests.assertions {
				entity, err := tuple.E(assertion.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(assertion.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				response, err := invoker.SubjectPermission(context.Background(), &base.PermissionSubjectPermissionRequest{
					TenantId: "t1",
					Subject:  subject,
					Entity:   entity,
					Metadata: &base.PermissionSubjectPermissionRequestMetadata{
						SnapToken:      token.NewNoopToken().Encode().String(),
						SchemaVersion:  "",
						Depth:          100,
						OnlyPermission: assertion.onlyPermission,
					},
				})

				Expect(err).ShouldNot(HaveOccurred())
				Expect(reflect.DeepEqual(response.Results, assertion.result)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type assertion struct {
				onlyPermission bool
				subject        string
				entity         string
				result         map[string]base.CheckResult
			}

			tests := struct {
				relationships []string
				assertions    []assertion
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:1#org@organization:1#...",
					"organization:1#admin@user:1",
				},
				assertions: []assertion{
					{
						subject:        "user:1",
						entity:         "doc:1",
						onlyPermission: false,
						result: map[string]base.CheckResult{
							// relations
							"org":    base.CheckResult_CHECK_RESULT_DENIED,
							"parent": base.CheckResult_CHECK_RESULT_DENIED,
							"owner":  base.CheckResult_CHECK_RESULT_DENIED,

							// permissions
							"read":   base.CheckResult_CHECK_RESULT_ALLOWED,
							"update": base.CheckResult_CHECK_RESULT_DENIED,
							"delete": base.CheckResult_CHECK_RESULT_ALLOWED,
							"share":  base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
					{
						subject:        "user:1",
						entity:         "doc:1",
						onlyPermission: true,
						result: map[string]base.CheckResult{
							"read":   base.CheckResult_CHECK_RESULT_ALLOWED,
							"update": base.CheckResult_CHECK_RESULT_DENIED,
							"delete": base.CheckResult_CHECK_RESULT_ALLOWED,
							"share":  base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			subjectPermissionEngine := NewSubjectPermission(checkEngine, schemaReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				subjectPermissionEngine,
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

			for _, assertion := range tests.assertions {
				entity, err := tuple.E(assertion.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(assertion.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				response, err := invoker.SubjectPermission(context.Background(), &base.PermissionSubjectPermissionRequest{
					TenantId: "t1",
					Subject:  subject,
					Entity:   entity,
					Metadata: &base.PermissionSubjectPermissionRequestMetadata{
						SnapToken:      token.NewNoopToken().Encode().String(),
						SchemaVersion:  "",
						Depth:          100,
						OnlyPermission: assertion.onlyPermission,
					},
				})

				Expect(err).ShouldNot(HaveOccurred())
				Expect(reflect.DeepEqual(response.Results, assertion.result)).Should(Equal(true))
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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type assertion struct {
				onlyPermission bool
				subject        string
				entity         string
				result         map[string]base.CheckResult
			}

			tests := struct {
				relationships []string
				assertions    []assertion
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:1#parent@folder:1#...",
					"folder:1#collaborator@user:7",
					"folder:1#collaborator@user:3",
					"doc:1#org@organization:1#...",
					"organization:1#admin@user:7",
				},
				assertions: []assertion{
					{
						subject:        "user:1",
						entity:         "doc:1",
						onlyPermission: false,
						result: map[string]base.CheckResult{
							// relations
							"org":    base.CheckResult_CHECK_RESULT_DENIED,
							"parent": base.CheckResult_CHECK_RESULT_DENIED,
							"owner":  base.CheckResult_CHECK_RESULT_DENIED,

							// permissions
							"read":   base.CheckResult_CHECK_RESULT_DENIED,
							"update": base.CheckResult_CHECK_RESULT_DENIED,
							"delete": base.CheckResult_CHECK_RESULT_DENIED,
							"share":  base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
					{
						subject:        "user:1",
						entity:         "doc:1",
						onlyPermission: true,
						result: map[string]base.CheckResult{
							"read":   base.CheckResult_CHECK_RESULT_DENIED,
							"update": base.CheckResult_CHECK_RESULT_DENIED,
							"delete": base.CheckResult_CHECK_RESULT_DENIED,
							"share":  base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			subjectPermissionEngine := NewSubjectPermission(checkEngine, schemaReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				subjectPermissionEngine,
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

			for _, assertion := range tests.assertions {
				entity, err := tuple.E(assertion.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(assertion.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				response, err := invoker.SubjectPermission(context.Background(), &base.PermissionSubjectPermissionRequest{
					TenantId: "t1",
					Subject:  subject,
					Entity:   entity,
					Metadata: &base.PermissionSubjectPermissionRequestMetadata{
						SnapToken:      token.NewNoopToken().Encode().String(),
						SchemaVersion:  "",
						Depth:          100,
						OnlyPermission: assertion.onlyPermission,
					},
				})

				Expect(err).ShouldNot(HaveOccurred())
				Expect(reflect.DeepEqual(response.Results, assertion.result)).Should(Equal(true))
			}
		})

		It("Drive Sample: Case 4: empty references", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type assertion struct {
				onlyPermission bool
				subject        string
				entity         string
				result         map[string]base.CheckResult
			}

			tests := struct {
				relationships []string
				assertions    []assertion
			}{
				relationships: []string{},
				assertions: []assertion{
					{
						subject:        "user:12",
						entity:         "organization:1",
						onlyPermission: true,
						result:         map[string]base.CheckResult{},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			subjectPermissionEngine := NewSubjectPermission(checkEngine, schemaReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				subjectPermissionEngine,
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

			for _, assertion := range tests.assertions {
				entity, err := tuple.E(assertion.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(assertion.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				response, err := invoker.SubjectPermission(context.Background(), &base.PermissionSubjectPermissionRequest{
					TenantId: "t1",
					Subject:  subject,
					Entity:   entity,
					Metadata: &base.PermissionSubjectPermissionRequestMetadata{
						SnapToken:      token.NewNoopToken().Encode().String(),
						SchemaVersion:  "",
						Depth:          100,
						OnlyPermission: assertion.onlyPermission,
					},
				})

				Expect(err).ShouldNot(HaveOccurred())
				Expect(reflect.DeepEqual(response.Results, assertion.result)).Should(Equal(true))
			}
		})
	})
})
