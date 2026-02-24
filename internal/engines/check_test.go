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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
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
						assertions: map[string]base.CheckResult{
							"read": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
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

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
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
						assertions: map[string]base.CheckResult{
							"update": base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
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

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
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
						assertions: map[string]base.CheckResult{
							"read": base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
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

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
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
						assertions: map[string]base.CheckResult{
							"push": base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
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

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
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
						assertions: map[string]base.CheckResult{
							"push": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
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
						Context: &base.Context{
							Tuples: tuples,
						},
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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
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
						assertions: map[string]base.CheckResult{
							"delete": base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
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

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
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
						assertions: map[string]base.CheckResult{
							"push": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
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
						Context: &base.Context{
							Tuples: tuples,
						},
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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
			}

			tests := struct {
				relationships []string
				contextual    []string
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
				contextual: []string{
					"parent:1#member@parent:1#admin",
					"repo:1#org@organization:1#...",
					"repo:1#parent@parent:1#...",
				},
				checks: []check{
					{
						entity:  "repo:1",
						subject: "user:2",
						assertions: map[string]base.CheckResult{
							"push": base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
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

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			var contextual []*base.Tuple

			for _, relationship := range tests.contextual {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				contextual = append(contextual, t)
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
						Context: &base.Context{
							Tuples: contextual,
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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
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
						assertions: map[string]base.CheckResult{
							"delete": base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
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

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
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
						assertions: map[string]base.CheckResult{
							"update": base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
					{
						entity:  "repo:1",
						subject: "user:2",
						assertions: map[string]base.CheckResult{
							"view": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
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

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
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
						assertions: map[string]base.CheckResult{
							"view": base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
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

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
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

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
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
						assertions: map[string]base.CheckResult{
							"admin": base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
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

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
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

	// POLYMORPHIC RELATIONS SAMPLE

	polymorphicRelationsSchema := `
	entity googleuser {}
	
	entity facebookuser {}
	
	entity company {
		relation member @googleuser @facebookuser
	}
	
	entity organization {
		relation member @googleuser @facebookuser
	
		action edit = member
	}
	
	entity repo {
		relation parent @company @organization
	
		permission push   = parent.member
		permission delete = push
	}
	`

	Context("Polymorphic Relations Sample: Check", func() {
		It("Polymorphic Relations Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(polymorphicRelationsSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
			}

			tests := struct {
				relationships []string
				checks        []check
			}{
				relationships: []string{
					"repo:1#parent@organization:1",
					"repo:1#parent@company:1",
					"company:1#member@googleuser:2",
					"organization:1#member@facebookuser:3",
				},
				checks: []check{
					{
						entity:  "repo:1",
						subject: "googleuser:2",
						assertions: map[string]base.CheckResult{
							"push": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
					{
						entity:  "repo:1",
						subject: "facebookuser:3",
						assertions: map[string]base.CheckResult{
							"push": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
					{
						entity:  "organization:1",
						subject: "facebookuser:3",
						assertions: map[string]base.CheckResult{
							"edit": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
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
						Context: &base.Context{
							Tuples: tuples,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
				}
			}
		})
	})

	// WORKDAY SAMPLE
	workdaySchema := `
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
			  is_public == true && (context.data.day_of_week != 'saturday' && context.data.day_of_week != 'sunday')
		}
		`

	Context("Weekday Sample: Check", func() {
		It("Weekday Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(workdaySchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
			}

			tests := struct {
				relationships []string
				attributes    []string
				checks        []check
			}{
				relationships: []string{},
				attributes: []string{
					"repository:1$is_public|boolean:true",
				},
				checks: []check{
					{
						entity:  "repository:1",
						subject: "user:1",
						assertions: map[string]base.CheckResult{
							"view": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)
			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple
			var attributes []*base.Attribute

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			for _, attr := range tests.attributes {
				t, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
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

		It("Weekday Sample: Case 2", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(workdaySchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				context    map[string]interface{}
				assertions map[string]base.CheckResult
			}

			tests := struct {
				relationships []string
				attributes    []string
				checks        []check
			}{
				relationships: []string{
					"organization:1#member@user:1",
					"repository:1#organization@organization:1",
				},
				attributes: []string{
					"organization:1$balance|integer:7000",
				},
				checks: []check{
					{
						entity:  "organization:1",
						subject: "user:1",
						assertions: map[string]base.CheckResult{
							"view": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
					{
						entity:  "repository:1",
						subject: "user:1",
						assertions: map[string]base.CheckResult{
							"edit": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
					{
						entity:  "repository:1",
						subject: "user:1",
						context: map[string]interface{}{
							"day_of_week": "saturday",
						},
						assertions: map[string]base.CheckResult{
							"delete": base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)
			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple
			var attributes []*base.Attribute

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			for _, attr := range tests.attributes {
				t, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
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

					ctx := &base.Context{
						Tuples:     []*base.Tuple{},
						Attributes: []*base.Attribute{},
						Data:       &structpb.Struct{},
					}

					if check.context != nil {
						value, err := structpb.NewStruct(check.context)
						if err != nil {
							fmt.Printf("Error creating struct: %v", err)
						}
						ctx.Data = value
					}

					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Context:    ctx,
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

		It("Weekday Sample: Case 3", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(workdaySchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				assertions map[string]base.CheckResult
			}

			tests := struct {
				relationships []string
				attributes    []string
				checks        []check
			}{
				relationships: []string{},
				attributes: []string{
					"repository:1$is_public|boolean:true",
				},
				checks: []check{
					{
						entity:  "repository:1",
						subject: "user:1",
						assertions: map[string]base.CheckResult{
							"view": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple
			var attributes []*base.Attribute

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			for _, attr := range tests.attributes {
				t, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, t)
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
						Context: &base.Context{
							Attributes: attributes,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
				}
			}
		})
	})

	// IP RANGE SAMPLE
	IpRangeSchema := `
		entity user {}
	
		entity organization {
	
			relation admin @user
	
			attribute ip_range string[]
	
			permission view = check_ip_range(ip_range) or admin
		}
	
		rule check_ip_range(ip_range string[]) {
			 context.data.ip_address in ip_range
		}
		`

	Context("Ip Range Sample: Check", func() {
		It("Ip Range Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(IpRangeSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				context    map[string]interface{}
				assertions map[string]base.CheckResult
			}

			tests := struct {
				relationships []string
				attributes    []string
				checks        []check
			}{
				relationships: []string{},
				attributes: []string{
					"organization:1$ip_range|string[]:18.216.238.147,94.176.248.171,61.49.24.70",
				},
				checks: []check{
					{
						entity:  "organization:1",
						subject: "user:1",
						context: map[string]interface{}{
							"ip_address": "18.216.238.147",
						},
						assertions: map[string]base.CheckResult{
							"view": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)
			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple
			var attributes []*base.Attribute

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			for _, attr := range tests.attributes {
				t, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
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

					ctx := &base.Context{
						Tuples:     []*base.Tuple{},
						Attributes: []*base.Attribute{},
						Data:       &structpb.Struct{},
					}

					if check.context != nil {
						value, err := structpb.NewStruct(check.context)
						if err != nil {
							fmt.Printf("Error creating struct: %v", err)
						}
						ctx.Data = value
					}

					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Context:    ctx,
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

	// DEPTH CHECK SAMPLE (3-level deep check)
	depthCheckSchema := `
	entity user {}

	entity bottom {
		relation member @user
		permission check = member
	}

	entity middle {
		relation parent @bottom
		permission check = parent.check
	}

	entity top {
		relation parent @middle
		permission check = parent.check
	}
	`

	Context("Depth Check Sample: Check", func() {
		It("Depth Check Sample: Case 1 - Depth 3 should pass for 3-level deep check", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(depthCheckSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				depth      int32
				assertions map[string]base.CheckResult
			}

			relationships := []string{
				"top:1#parent@middle:1#...",
				"middle:1#parent@bottom:1#...",
				"bottom:1#member@user:1",

				"top:2#parent@middle:2#...",
				"middle:2#parent@bottom:2#...",
				"bottom:2#member@user:2",

				"top:3#parent@middle:3#...",
				"middle:3#parent@bottom:3#...",
				"bottom:3#member@user:3",

				"top:4#parent@middle:4#...",
				"middle:4#parent@bottom:4#...",
				"bottom:4#member@user:4",

				"top:5#parent@middle:5#...",
				"middle:5#parent@bottom:5#...",
				"bottom:5#member@user:5",

				"top:6#parent@middle:6#...",
				"middle:6#parent@bottom:6#...",
				"bottom:6#member@user:6",

				"top:7#parent@middle:7#...",
				"middle:7#parent@bottom:7#...",
				"bottom:7#member@user:7",

				"top:8#parent@middle:8#...",
				"middle:8#parent@bottom:8#...",
				"bottom:8#member@user:8",

				"top:9#parent@middle:9#...",
				"middle:9#parent@bottom:9#...",
				"bottom:9#member@user:9",

				"top:10#parent@middle:10#...",
				"middle:10#parent@bottom:10#...",
				"bottom:10#member@user:10",
			}

			var checks []check
			for i := 1; i <= 10; i++ {
				entity := fmt.Sprintf("top:%d", i)
				subject := fmt.Sprintf("user:%d", i)
				checks = append(checks, check{
					entity:  entity,
					subject: subject,
					depth:   3,
					assertions: map[string]base.CheckResult{
						"check": base.CheckResult_CHECK_RESULT_ALLOWED,
					},
				})
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, check := range checks {
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
							Depth:         check.depth,
						},
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(res).Should(Equal(response.GetCan()))
				}
			}
		})

		It("Depth Check Sample: Case 2 - Depth 2 should fail for 3-level deep check", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(depthCheckSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)

			Expect(err).ShouldNot(HaveOccurred())

			type check struct {
				entity     string
				subject    string
				depth      int32
				assertions map[string]base.CheckResult
			}

			relationships := []string{
				"top:1#parent@middle:1#...",
				"middle:1#parent@bottom:1#...",
				"bottom:1#member@user:1",

				"top:2#parent@middle:2#...",
				"middle:2#parent@bottom:2#...",
				"bottom:2#member@user:2",

				"top:3#parent@middle:3#...",
				"middle:3#parent@bottom:3#...",
				"bottom:3#member@user:3",

				"top:4#parent@middle:4#...",
				"middle:4#parent@bottom:4#...",
				"bottom:4#member@user:4",

				"top:5#parent@middle:5#...",
				"middle:5#parent@bottom:5#...",
				"bottom:5#member@user:5",

				"top:6#parent@middle:6#...",
				"middle:6#parent@bottom:6#...",
				"bottom:6#member@user:6",

				"top:7#parent@middle:7#...",
				"middle:7#parent@bottom:7#...",
				"bottom:7#member@user:7",

				"top:8#parent@middle:8#...",
				"middle:8#parent@bottom:8#...",
				"bottom:8#member@user:8",

				"top:9#parent@middle:9#...",
				"middle:9#parent@bottom:9#...",
				"bottom:9#member@user:9",

				"top:10#parent@middle:10#...",
				"middle:10#parent@bottom:10#...",
				"bottom:10#member@user:10",
			}

			var checks []check
			for i := 1; i <= 10; i++ {
				entity := fmt.Sprintf("top:%d", i)
				subject := fmt.Sprintf("user:%d", i)
				checks = append(checks, check{
					entity:  entity,
					subject: subject,
					depth:   2,
					assertions: map[string]base.CheckResult{
						"check": base.CheckResult_CHECK_RESULT_DENIED,
					},
				})
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngine,
				nil,
				nil,
				nil,
			)

			checkEngine.SetInvoker(invoker)

			var tuples []*base.Tuple

			for _, relationship := range relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			for _, check := range checks {
				entity, err := tuple.E(check.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ear, err := tuple.EAR(check.subject)
				Expect(err).ShouldNot(HaveOccurred())

				subject := &base.Subject{
					Type:     ear.GetEntity().GetType(),
					Id:       ear.GetEntity().GetId(),
					Relation: ear.GetRelation(),
				}

				for permission := range check.assertions {
					response, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: permission,
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         check.depth,
						},
					})

					Expect(err).Should(HaveOccurred())
					Expect(err.Error()).Should(ContainSubstring("ERROR_CODE_DEPTH_NOT_ENOUGH"))
					Expect(response.GetCan()).Should(Equal(base.CheckResult_CHECK_RESULT_DENIED))
				}
			}
		})
	})

	Context("Recursive Attribute Permissions", func() {
		It("should allow same-type recursive attribute permissions", func() {
			schema := `
			entity user {}

			entity resource {
				relation parent @resource
				attribute is_public boolean
				permission view = is_public or parent.view
			}
			`

			db, err := factories.DatabaseFactory(config.Database{Engine: "memory"})
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(schema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)
			lookupEngine := NewLookupEngine(checkEngine, schemaReader, dataReader)
			invoker := invoke.NewDirectInvoker(schemaReader, dataReader, checkEngine, nil, lookupEngine, nil)
			checkEngine.SetInvoker(invoker)

			relationships := []string{
				"resource:r1#parent@resource:default",
			}

			var tuples []*base.Tuple
			for _, relationship := range relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			publicAttr, err := attribute.Attribute("resource:default$is_public|boolean:true")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = dataWriter.Write(
				context.Background(),
				"t1",
				database.NewTupleCollection(tuples...),
				database.NewAttributeCollection(publicAttr),
			)
			Expect(err).ShouldNot(HaveOccurred())

			resp, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
				TenantId:   "t1",
				Entity:     &base.Entity{Type: "resource", Id: "r1"},
				Permission: "view",
				Subject:    &base.Subject{Type: "user", Id: "u1"},
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "",
					Depth:         20,
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.GetCan()).To(Equal(base.CheckResult_CHECK_RESULT_ALLOWED))
		})

		It("should allow cross-type recursive attribute permissions", func() {
			schema := `
			entity user {}

			entity org {
				attribute is_public boolean
				permission view = is_public
			}

			entity folder {
				relation parent @org
				attribute is_public boolean
				permission view = is_public or parent.view
			}

			entity resource {
				relation parent @folder
				permission view = parent.view
			}
			`

			db, err := factories.DatabaseFactory(config.Database{Engine: "memory"})
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(schema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)
			lookupEngine := NewLookupEngine(checkEngine, schemaReader, dataReader)
			invoker := invoke.NewDirectInvoker(schemaReader, dataReader, checkEngine, nil, lookupEngine, nil)
			checkEngine.SetInvoker(invoker)

			relationships := []string{
				"folder:f1#parent@org:o1",
				"resource:r1#parent@folder:f1",
			}

			var tuples []*base.Tuple
			for _, relationship := range relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			publicAttr, err := attribute.Attribute("org:o1$is_public|boolean:true")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = dataWriter.Write(
				context.Background(),
				"t1",
				database.NewTupleCollection(tuples...),
				database.NewAttributeCollection(publicAttr),
			)
			Expect(err).ShouldNot(HaveOccurred())

			resp, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
				TenantId:   "t1",
				Entity:     &base.Entity{Type: "resource", Id: "r1"},
				Permission: "view",
				Subject:    &base.Subject{Type: "user", Id: "u1"},
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "",
					Depth:         20,
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.GetCan()).To(Equal(base.CheckResult_CHECK_RESULT_ALLOWED))
		})

		It("should allow mixed-entrance recursive attribute permissions", func() {
			schema := `
			entity user {}

			entity resource {
				relation viewer @user
				relation parent @resource
				attribute is_public boolean
				permission view = viewer or is_public or parent.view
			}
			`

			db, err := factories.DatabaseFactory(config.Database{Engine: "memory"})
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(schema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)
			lookupEngine := NewLookupEngine(checkEngine, schemaReader, dataReader)
			invoker := invoke.NewDirectInvoker(schemaReader, dataReader, checkEngine, nil, lookupEngine, nil)
			checkEngine.SetInvoker(invoker)

			relationships := []string{
				"resource:za#viewer@user:u1",
				"resource:zb#parent@resource:za",
				"resource:zc#parent@resource:zb",
			}

			var tuples []*base.Tuple
			for _, relationship := range relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			publicAttr, err := attribute.Attribute("resource:za$is_public|boolean:true")
			Expect(err).ShouldNot(HaveOccurred())

			_, err = dataWriter.Write(
				context.Background(),
				"t1",
				database.NewTupleCollection(tuples...),
				database.NewAttributeCollection(publicAttr),
			)
			Expect(err).ShouldNot(HaveOccurred())

			resp, err := invoker.Check(context.Background(), &base.PermissionCheckRequest{
				TenantId:   "t1",
				Entity:     &base.Entity{Type: "resource", Id: "zc"},
				Permission: "view",
				Subject:    &base.Subject{Type: "user", Id: "u1"},
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "",
					Depth:         20,
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.GetCan()).To(Equal(base.CheckResult_CHECK_RESULT_ALLOWED))
		})
	})
})
