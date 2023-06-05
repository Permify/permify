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

var _ = Describe("check-engine", func() {
	// DRIVE SAMPLE

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

	Context("Drive Sample: Check", func() {
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

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.PermissionCheckResponse_Result
			}

			tests := struct {
				relationships []string
				checks        []check
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:1#folder@user:3",
					"folder:1#collaborator@user:1",
					"folder:1#collaborator@user:3",
					"organization:1#admin@user:1",
					"doc:1#org@organization:1#...",
				},
				checks: []check{
					{
						entity:  "doc:1",
						subject: "user:1",
						assertions: map[string]base.PermissionCheckResponse_Result{
							"read": base.PermissionCheckResponse_RESULT_ALLOWED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, check := range tests.checks {
				entity, err := tuple.E(check.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(check.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range check.assertions {
					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         20,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
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

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.PermissionCheckResponse_Result
			}

			tests := struct {
				relationships []string
				checks        []check
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:1#org@organization:1#...",
					"organization:1#admin@user:1",
				},
				checks: []check{
					{
						entity:  "doc:1",
						subject: "user:1",
						assertions: map[string]base.PermissionCheckResponse_Result{
							"update": base.PermissionCheckResponse_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, check := range tests.checks {
				entity, err := tuple.E(check.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(check.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range check.assertions {
					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         20,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
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

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.PermissionCheckResponse_Result
			}

			tests := struct {
				relationships []string
				checks        []check
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:1#parent@folder:1#...",
					"folder:1#collaborator@user:7",
					"folder:1#collaborator@user:3",
					"doc:1#org@organization:1#...",
					"organization:1#admin@user:7",
				},
				checks: []check{
					{
						entity:  "doc:1",
						subject: "user:1",
						assertions: map[string]base.PermissionCheckResponse_Result{
							"read": base.PermissionCheckResponse_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, check := range tests.checks {
				entity, err := tuple.E(check.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(check.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range check.assertions {
					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         20,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
				}
			}
		})
	})

	// GITHUB SAMPLE

	githubSchema := `
	entity user {}
	
	entity organization {
		relation admin @user
		relation member @user
	
		action create_repository = admin or member
		action delete = admin
	}
	
	entity repository {
		relation parent @organization
		relation owner @user
	
		action push   = owner
		action read   = owner and (parent.admin or parent.member)
		action delete = parent.member and (parent.admin or owner)
	}
	`

	Context("Github Sample: Check", func() {
		It("Github Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(githubSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.PermissionCheckResponse_Result
			}

			tests := struct {
				relationships []string
				checks        []check
			}{
				relationships: []string{
					"repository:1#owner@user:2",
				},
				checks: []check{
					{
						entity:  "repository:1",
						subject: "user:1",
						assertions: map[string]base.PermissionCheckResponse_Result{
							"push": base.PermissionCheckResponse_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, check := range tests.checks {
				entity, err := tuple.E(check.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(check.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range check.assertions {
					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         20,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
				}
			}
		})

		It("Github Sample: Case 2", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(githubSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.PermissionCheckResponse_Result
			}

			tests := struct {
				relationships []string
				checks        []check
			}{
				relationships: []string{
					"repository:1#owner@organization:2#admin",
					"organization:2#admin@organization:3#member",
					"organization:2#admin@user:3",
					"organization:2#admin@user:8",
					"organization:3#member@user:1",
				},
				checks: []check{
					{
						entity:  "repository:1",
						subject: "user:1",
						assertions: map[string]base.PermissionCheckResponse_Result{
							"push": base.PermissionCheckResponse_RESULT_ALLOWED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			for _, check := range tests.checks {
				entity, err := tuple.E(check.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(check.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range check.assertions {
					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         20,
						},
						ContextualTuples: tuples,
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
				}
			}
		})

		It("Github Sample: Case 3", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(githubSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.PermissionCheckResponse_Result
			}

			tests := struct {
				relationships []string
				checks        []check
			}{
				relationships: []string{
					"repository:1#parent@organization:8#...",
					"organization:8#member@user:1",
					"organization:8#admin@user:2",
					"repository:1#owner@user:7",
				},
				checks: []check{
					{
						entity:  "repository:1",
						subject: "user:1",
						assertions: map[string]base.PermissionCheckResponse_Result{
							"delete": base.PermissionCheckResponse_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, check := range tests.checks {
				entity, err := tuple.E(check.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(check.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range check.assertions {
					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         20,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
				}
			}
		})
	})

	// EXCLUSION SAMPLE

	exclusionSchema := `
entity user {}
	
entity organization {
	relation member @user
}
	
entity parent {
	relation member @user
	relation admin @user
}
	
entity repo {	
	relation org @organization
	relation parent @parent
	relation member @user
	
	permission push   = org.member not parent.member
	permission delete = push

	permission update = (org.member not parent.member) and member
	permission view = member not update
	permission manage = update
	permission admin = manage
}
`

	Context("Exclusion Sample: Check", func() {
		It("Exclusion Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(exclusionSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.PermissionCheckResponse_Result
			}

			tests := struct {
				relationships []string
				checks        []check
			}{
				relationships: []string{
					"organization:1#member@user:1",
					"organization:1#member@user:2",
					"parent:1#member@user:1",
					"repo:1#org@organization:1#...",
					"repo:1#parent@parent:1#...",
				},
				checks: []check{
					{
						entity:  "repo:1",
						subject: "user:2",
						assertions: map[string]base.PermissionCheckResponse_Result{
							"push": base.PermissionCheckResponse_RESULT_ALLOWED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			for _, check := range tests.checks {
				entity, err := tuple.E(check.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(check.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range check.assertions {
					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         20,
						},
						ContextualTuples: tuples,
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
				}
			}
		})

		It("Exclusion Sample: Case 2", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(exclusionSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.PermissionCheckResponse_Result
			}

			tests := struct {
				relationships []string
				checks        []check
			}{
				relationships: []string{
					"organization:1#member@user:1",
					"organization:1#member@user:2",
					"parent:1#admin@user:2",
					"parent:1#member@user:1",
					"parent:1#member@parent:1#admin",
					"repo:1#org@organization:1#...",
					"repo:1#parent@parent:1#...",
				},
				checks: []check{
					{
						entity:  "repo:1",
						subject: "user:2",
						assertions: map[string]base.PermissionCheckResponse_Result{
							"push": base.PermissionCheckResponse_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, check := range tests.checks {
				entity, err := tuple.E(check.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(check.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range check.assertions {
					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         20,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
				}
			}
		})

		It("Exclusion Sample: Case 3", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(exclusionSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.PermissionCheckResponse_Result
			}

			tests := struct {
				relationships []string
				checks        []check
			}{
				relationships: []string{
					"organization:1#member@user:1",
					"organization:1#member@user:2",
					"parent:1#admin@user:2",
					"parent:1#member@user:1",
					"parent:1#member@user:2",
					"repo:1#org@organization:1#...",
					"repo:1#parent@parent:1#...",
				},
				checks: []check{
					{
						entity:  "repo:1",
						subject: "user:2",
						assertions: map[string]base.PermissionCheckResponse_Result{
							"delete": base.PermissionCheckResponse_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, check := range tests.checks {
				entity, err := tuple.E(check.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(check.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range check.assertions {
					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         20,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
				}
			}
		})

		It("Exclusion Sample: Case 4", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(exclusionSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.PermissionCheckResponse_Result
			}

			tests := struct {
				relationships []string
				checks        []check
			}{
				relationships: []string{
					"organization:1#member@user:1",

					"parent:1#admin@user:2",
					"parent:1#member@user:1",
					"parent:1#member@parent:1#admin",

					"repo:1#org@organization:1#...",
					"repo:1#parent@parent:1#...",
					"repo:1#member@user:2",
				},
				checks: []check{
					{
						entity:  "repo:1",
						subject: "user:2",
						assertions: map[string]base.PermissionCheckResponse_Result{
							"update": base.PermissionCheckResponse_RESULT_DENIED,
						},
					},
					{
						entity:  "repo:1",
						subject: "user:2",
						assertions: map[string]base.PermissionCheckResponse_Result{
							"view": base.PermissionCheckResponse_RESULT_ALLOWED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, check := range tests.checks {
				entity, err := tuple.E(check.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(check.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range check.assertions {
					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         20,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
				}
			}
		})

		It("Exclusion Sample: Case 5", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(exclusionSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.PermissionCheckResponse_Result
			}

			tests := struct {
				relationships []string
				checks        []check
			}{
				relationships: []string{
					"organization:1#member@user:1",
					"parent:1#admin@user:2",
					"parent:1#member@user:1",
					"parent:1#member@parent:1#admin",
					"repo:1#org@organization:1#...",
					"repo:1#parent@parent:1#...",
				},
				checks: []check{
					{
						entity:  "repo:1",
						subject: "user:2",
						assertions: map[string]base.PermissionCheckResponse_Result{
							"view": base.PermissionCheckResponse_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, check := range tests.checks {
				entity, err := tuple.E(check.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(check.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range check.assertions {
					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         20,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
				}
			}
		})

		It("Exclusion Sample: Case 6", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(exclusionSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db, logger.New("debug"))
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.PermissionCheckResponse_Result
			}

			tests := struct {
				relationships []string
				checks        []check
			}{
				relationships: []string{
					"organization:1#member@user:1",
					"parent:1#admin@user:2",
					"parent:1#member@user:1",
					"parent:1#member@parent:1#admin",
					"repo:1#org@organization:1#...",
					"repo:1#parent@parent:1#...",
				},
				checks: []check{
					{
						entity:  "repo:1",
						subject: "user:2",
						assertions: map[string]base.PermissionCheckResponse_Result{
							"admin": base.PermissionCheckResponse_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db, logger.New("debug"))
			relationshipReader := factories.RelationshipReaderFactory(db, logger.New("debug"))
			relationshipWriter := factories.RelationshipWriterFactory(db, logger.New("debug"))

			checkEngine := NewCheckEngine(schemaReader, relationshipReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				relationshipReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = relationshipWriter.WriteRelationships(context.Background(), "t1", database.NewTupleCollection(tuples...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, check := range tests.checks {
				entity, err := tuple.E(check.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(check.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission, res := range check.assertions {
					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         20,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
				}
			}
		})
	})
})
