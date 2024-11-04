package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/rs/xid"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/engines"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/pkg/attribute"
	pkgcache "github.com/Permify/permify/pkg/cache"
	"github.com/Permify/permify/pkg/cache/ristretto"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/dsl/compiler"
	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

func TestEngines(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cache-suite")
}

var _ = Describe("cache", func() {
	Context("Set Cache Key", func() {
		It("Success", func() {
			// Initialize a new Ristretto cache with a capacity of 10 cache
			cache, err := ristretto.New()
			Expect(err).ShouldNot(HaveOccurred())

			// Initialize a new EngineKeys struct with a new cache.Cache instance
			engineKeys := CheckEngineWithCache{nil, nil, cache, nil}

			// Create a new PermissionCheckRequest and PermissionCheckResponse
			checkReq := &base.PermissionCheckRequest{
				TenantId: "t1",
				Metadata: &base.PermissionCheckRequestMetadata{
					SchemaVersion: "test_version",
					SnapToken:     "test_snap_token",
					Depth:         20,
				},
				Entity: &base.Entity{
					Type: "test-entity",
					Id:   "e1",
				},
				Permission: "test-permission",
				Subject: &base.Subject{
					Type: "user",
					Id:   "u1",
				},
			}

			checkResp := &base.PermissionCheckResponse{
				Can: base.CheckResult_CHECK_RESULT_ALLOWED,
				Metadata: &base.PermissionCheckResponseMetadata{
					CheckCount: 0,
				},
			}

			// Set the value for the given key in the cache
			success := engineKeys.setCheckKey(checkReq, checkResp, true)

			cache.Wait()

			// Check that the operation was successful
			Expect(success).Should(BeTrue())

			// Retrieve the value for the given key from the cache
			resp, found := engineKeys.getCheckKey(checkReq, true)

			// Check that the key was found and the retrieved value is the same as the original value
			Expect(found).Should(BeTrue())
			Expect(checkResp).Should(Equal(resp))
		})

		It("With Hash Error", func() {
			// Initialize a new Ristretto cache with a capacity of 10 cache
			cache, err := ristretto.New()
			Expect(err).ShouldNot(HaveOccurred())

			// Initialize a new EngineKeys struct with a new cache.Cache instance
			engineKeys := CheckEngineWithCache{nil, nil, cache, nil}

			// Create a new PermissionCheckRequest and PermissionCheckResponse
			checkReq := &base.PermissionCheckRequest{
				TenantId: "t1",
				Metadata: &base.PermissionCheckRequestMetadata{
					SchemaVersion: "test_version",
					SnapToken:     "test_snap_token",
					Depth:         20,
				},
				Entity: &base.Entity{
					Type: "test-entity",
					Id:   "e1",
				},
				Permission: "test-permission",
				Subject: &base.Subject{
					Type: "user",
					Id:   "u1",
				},
			}

			checkResp := &base.PermissionCheckResponse{
				Can: base.CheckResult_CHECK_RESULT_ALLOWED,
				Metadata: &base.PermissionCheckResponseMetadata{
					CheckCount: 0,
				},
			}

			// Force an error while writing the key to the hash object by passing a nil key
			success := engineKeys.setCheckKey(nil, checkResp, true)

			cache.Wait()

			// Check that the operation was unsuccessful
			Expect(success).Should(BeFalse())

			// Retrieve the value for the given key from the cache
			resp, found := engineKeys.getCheckKey(checkReq, true)

			// Check that the key was not found
			Expect(found).Should(BeFalse())
			Expect(resp).Should(BeNil())
		})

		It("With Arguments", func() {
			// Initialize a new Ristretto cache with a capacity of 10 cache
			cache, err := ristretto.New()
			Expect(err).ShouldNot(HaveOccurred())

			// Initialize a new EngineKeys struct with a new cache.Cache instance
			engineKeys := CheckEngineWithCache{nil, nil, cache, nil}

			// Create a new PermissionCheckRequest and PermissionCheckResponse
			checkReq := &base.PermissionCheckRequest{
				TenantId: "t1",
				Metadata: &base.PermissionCheckRequestMetadata{
					SchemaVersion: "test_version",
					SnapToken:     "test_snap_token",
					Depth:         20,
				},
				Arguments: []*base.Argument{
					{
						Type: &base.Argument_ComputedAttribute{
							ComputedAttribute: &base.ComputedAttribute{
								Name: "test_argument_1",
							},
						},
					},
					{
						Type: &base.Argument_ComputedAttribute{
							ComputedAttribute: &base.ComputedAttribute{
								Name: "test_argument_2",
							},
						},
					},
				},
				Entity: &base.Entity{
					Type: "test-entity",
					Id:   "e1",
				},
				Permission: "test-rule",
				Subject: &base.Subject{
					Type: "user",
					Id:   "u1",
				},
			}

			checkResp := &base.PermissionCheckResponse{
				Can: base.CheckResult_CHECK_RESULT_ALLOWED,
				Metadata: &base.PermissionCheckResponseMetadata{
					CheckCount: 0,
				},
			}

			// Set the value for the given key in the cache
			success := engineKeys.setCheckKey(checkReq, checkResp, true)

			cache.Wait()

			// Check that the operation was successful
			Expect(success).Should(BeTrue())

			// Retrieve the value for the given key from the cache
			resp, found := engineKeys.getCheckKey(checkReq, true)

			// Check that the key was found and the retrieved value is the same as the original value
			Expect(found).Should(BeTrue())
			Expect(checkResp).Should(Equal(resp))
		})

		It("With Context", func() {
			value, err := anypb.New(&base.BooleanValue{Data: true})
			if err != nil {
			}

			data, err := structpb.NewStruct(map[string]interface{}{
				"day_of_a_week": "saturday",
				"day_of_a_year": 356,
			})
			if err != nil {
			}

			// Create a new PermissionCheckRequest and PermissionCheckResponse
			checkReq := &base.PermissionCheckRequest{
				TenantId: "t1",
				Metadata: &base.PermissionCheckRequestMetadata{
					SchemaVersion: "test_version",
					SnapToken:     "test_snap_token",
					Depth:         20,
				},
				Arguments: []*base.Argument{
					{
						Type: &base.Argument_ComputedAttribute{
							ComputedAttribute: &base.ComputedAttribute{
								Name: "test_argument_1",
							},
						},
					},
					{
						Type: &base.Argument_ComputedAttribute{
							ComputedAttribute: &base.ComputedAttribute{
								Name: "test_argument_2",
							},
						},
					},
				},
				Context: &base.Context{
					Tuples: []*base.Tuple{
						{
							Entity: &base.Entity{
								Type: "entity_type",
								Id:   "entity_id",
							},
							Relation: "relation",
							Subject: &base.Subject{
								Type: "subject_type",
								Id:   "subject_id",
							},
						},
					},
					Attributes: []*base.Attribute{
						{
							Entity: &base.Entity{
								Type: "entity_type",
								Id:   "entity_id",
							},
							Attribute: "is_public",
							Value:     value,
						},
					},
					Data: data,
				},
				Entity: &base.Entity{
					Type: "test-entity",
					Id:   "e1",
				},
				Permission: "test-rule",
				Subject: &base.Subject{
					Type: "user",
					Id:   "u1",
				},
			}

			Expect(engines.GenerateKey(checkReq, false)).Should(Equal("check|t1|test_version|test_snap_token|entity_type:entity_id#relation@subject_type:subject_id,entity_type:entity_id$is_public|boolean:true,day_of_a_week:saturday,day_of_a_year:356|test-entity:e1$test-rule(test_argument_1,test_argument_2)"))
		})
	})

	Context("Get Cache Key", func() {
		It("Key Not Found", func() {
			// Initialize a new Ristretto cache with a capacity of 10 cache
			cache, err := ristretto.New()
			Expect(err).ShouldNot(HaveOccurred())

			// Initialize a new EngineKeys struct with a new cache.Cache instance
			engineKeys := CheckEngineWithCache{nil, nil, cache, nil}

			// Create a new PermissionCheckRequest
			checkReq := &base.PermissionCheckRequest{
				TenantId: "t1",
				Metadata: &base.PermissionCheckRequestMetadata{
					SchemaVersion: "test_version",
					SnapToken:     "test_snap_token",
					Depth:         20,
				},
				Entity: &base.Entity{
					Type: "test-entity",
					Id:   "e1",
				},
				Permission: "test-permission",
				Subject: &base.Subject{
					Type: "user",
					Id:   "u1",
				},
			}

			// Retrieve the value for a non-existent key from the cache
			resp, found := engineKeys.getCheckKey(checkReq, true)

			// Check that the key was not found
			Expect(found).Should(BeFalse())
			Expect(resp).Should(BeNil())
		})

		It("Set and Get Multiple Keys", func() {
			// Initialize a new Ristretto cache with a capacity of 10 cache
			cache, err := ristretto.New()
			Expect(err).ShouldNot(HaveOccurred())

			// Initialize a new EngineKeys struct with a new cache.Cache instance
			engineKeys := CheckEngineWithCache{nil, nil, cache, nil}

			// Create some new PermissionCheckRequests and PermissionCheckResponses
			checkReq1 := &base.PermissionCheckRequest{
				TenantId: "t1",
				Metadata: &base.PermissionCheckRequestMetadata{
					SchemaVersion: "test_version",
					SnapToken:     "test_snap_token",
					Depth:         20,
				},
				Entity: &base.Entity{
					Type: "test-entity",
					Id:   "e1",
				},
				Permission: "test-permission",
				Subject: &base.Subject{
					Type: "user",
					Id:   "u1",
				},
			}
			checkResp1 := &base.PermissionCheckResponse{
				Can: base.CheckResult_CHECK_RESULT_ALLOWED,
				Metadata: &base.PermissionCheckResponseMetadata{
					CheckCount: 0,
				},
			}

			checkReq2 := &base.PermissionCheckRequest{
				TenantId: "t1",
				Metadata: &base.PermissionCheckRequestMetadata{
					SchemaVersion: "test_version",
					SnapToken:     "test_snap_token",
					Depth:         20,
				},
				Entity: &base.Entity{
					Type: "test-entity",
					Id:   "e2",
				},
				Permission: "test-permission",
				Subject: &base.Subject{
					Type: "user",
					Id:   "u1",
				},
			}
			checkResp2 := &base.PermissionCheckResponse{
				Can: base.CheckResult_CHECK_RESULT_DENIED,
				Metadata: &base.PermissionCheckResponseMetadata{
					CheckCount: 0,
				},
			}

			checkReq3 := &base.PermissionCheckRequest{
				TenantId: "t2",
				Metadata: &base.PermissionCheckRequestMetadata{
					SchemaVersion: "test_version",
					SnapToken:     "test_snap_token",
					Depth:         20,
				},
				Entity: &base.Entity{
					Type: "test-entity",
					Id:   "e1",
				},
				Permission: "test-permission",
				Subject: &base.Subject{
					Type: "user",
					Id:   "u2",
				},
			}
			checkResp3 := &base.PermissionCheckResponse{
				Can: base.CheckResult_CHECK_RESULT_DENIED,
				Metadata: &base.PermissionCheckResponseMetadata{
					CheckCount: 0,
				},
			}

			// Set the values for the given cache in the cache
			success1 := engineKeys.setCheckKey(checkReq1, checkResp1, true)
			success2 := engineKeys.setCheckKey(checkReq2, checkResp2, true)
			success3 := engineKeys.setCheckKey(checkReq3, checkResp3, true)

			cache.Wait()

			// Check that all the operations were successful
			Expect(success1).Should(BeTrue())
			Expect(success2).Should(BeTrue())
			Expect(success3).Should(BeTrue())

			// Retrieve the value for the given key from the cache
			resp1, found1 := engineKeys.getCheckKey(checkReq1, true)
			resp2, found2 := engineKeys.getCheckKey(checkReq2, true)
			resp3, found3 := engineKeys.getCheckKey(checkReq3, true)

			// Check that the key was not found
			Expect(found1).Should(BeTrue())
			Expect(checkResp1).Should(Equal(resp1))

			Expect(found2).Should(BeTrue())
			Expect(checkResp2).Should(Equal(resp2))

			Expect(found3).Should(BeTrue())
			Expect(checkResp3).Should(Equal(resp3))
		})
	})

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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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
					{
						entity:  "repository:1",
						subject: "user:1",
						assertions: map[string]base.CheckResult{
							"push": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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
					{
						entity:  "repo:1",
						subject: "user:2",
						assertions: map[string]base.CheckResult{
							"push": base.CheckResult_CHECK_RESULT_ALLOWED,
						},
					},
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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
					{
						entity:  "repo:1",
						subject: "user:2",
						assertions: map[string]base.CheckResult{
							"delete": base.CheckResult_CHECK_RESULT_DENIED,
						},
					},
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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
						entity:  "repo:1",
						subject: "facebookuser:3",
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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
	workkdaySchema := `
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

			conf, err := newSchema(workkdaySchema)
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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

			conf, err := newSchema(workkdaySchema)
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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

			conf, err := newSchema(workkdaySchema)
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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

			// engines cache cache
			var engineKeyCache pkgcache.Cache
			engineKeyCache, err = ristretto.New(ristretto.NumberOfCounters(1_000), ristretto.MaxCost("10MiB"))
			Expect(err).ShouldNot(HaveOccurred())

			checkEngine := engines.NewCheckEngine(schemaReader, dataReader)
			checkEngineWithCache := NewCheckEngineWithCache(checkEngine, schemaReader, engineKeyCache)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				checkEngineWithCache,
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
})

// newSchema -
func newSchema(model string) ([]storage.SchemaDefinition, error) {
	sch, err := parser.NewParser(model).Parse()
	if err != nil {
		return nil, err
	}

	_, _, err = compiler.NewCompiler(false, sch).Compile()
	if err != nil {
		return nil, err
	}

	version := xid.New().String()

	cnf := make([]storage.SchemaDefinition, 0, len(sch.Statements))
	for _, st := range sch.Statements {
		cnf = append(cnf, storage.SchemaDefinition{
			TenantID:             "t1",
			Version:              version,
			Name:                 st.GetName(),
			SerializedDefinition: []byte(st.String()),
		})
	}

	return cnf, err
}
