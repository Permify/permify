package engines

import (
	"context"
	"fmt"
	"runtime"
	"time"

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


var _ = Describe("goroutine-leak-tests", func() {
	Context("Goroutine Leak Tests", func() {
		It("Should not leak goroutines during concurrent checks", func() {
			testSchema := `
				entity user {}
				
				entity organization {
					relation admin @user
					relation member @user
				}
				
				entity repository {
					relation owner @user
					relation contributor @user
					relation parent @organization
					
					permission read = (owner or contributor) and parent.member
					permission write = owner and parent.admin
					permission delete = owner and parent.admin
				}
			`

			db, err := factories.DatabaseFactory(config.Database{Engine: "memory"})
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(testSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)
			invoker := invoke.NewDirectInvoker(schemaReader, dataReader, checkEngine, nil, nil, nil)
			checkEngine.SetInvoker(invoker)

			// Create test data that will trigger multiple goroutines through complex permission evaluation
			relationships := []string{
				"organization:1#admin@user:1",
				"organization:1#member@user:1",
				"organization:1#member@user:2",
				"organization:1#member@user:3",
				"organization:2#admin@user:2",
				"organization:2#member@user:2",
				"organization:2#member@user:3",
				"repository:1#owner@user:1",
				"repository:1#contributor@user:2",
				"repository:1#contributor@user:3",
				"repository:1#parent@organization:1#...",
				"repository:2#owner@user:2",
				"repository:2#contributor@user:1",
				"repository:2#parent@organization:2#...",
			}

			var tuples []*base.Tuple
			for _, relationship := range relationships {
				tpl, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, tpl)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			// Record initial goroutine count after stabilization
			runtime.GC()
			time.Sleep(200 * time.Millisecond)
			runtime.GC()
			time.Sleep(100 * time.Millisecond)
			initialCount := runtime.NumGoroutine()

			// Run multiple concurrent checks to stress test goroutine management
			// Test various permission types to trigger different code paths
			numChecks := 100
			done := make(chan struct{}, numChecks)
			permissions := []string{"read", "write", "delete"}

			for i := 0; i < numChecks; i++ {
				go func(idx int) {
					defer func() { done <- struct{}{} }()
					
					// Alternate between different repositories and users
					repoId := (idx % 2) + 1
					userId := (idx % 3) + 1
					permission := permissions[idx%3]
					
					entity, err := tuple.E(fmt.Sprintf("repository:%d", repoId))
					Expect(err).ShouldNot(HaveOccurred())

					ear, err := tuple.EAR(fmt.Sprintf("user:%d", userId))
					Expect(err).ShouldNot(HaveOccurred())

					subject := &base.Subject{
						Type:     ear.GetEntity().GetType(),
						Id:       ear.GetEntity().GetId(),
						Relation: ear.GetRelation(),
					}

					// Create a context with timeout to prevent hanging
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					_, err = invoker.Check(ctx, &base.PermissionCheckRequest{
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
				}(i)
			}

			// Wait for all checks to complete
			for i := 0; i < numChecks; i++ {
				select {
				case <-done:
					// Success
				case <-time.After(30 * time.Second):
					Fail("Test timed out waiting for goroutines to complete")
				}
			}

			// Allow sufficient time for goroutines to cleanup
			time.Sleep(1 * time.Second)
			runtime.GC()
			time.Sleep(200 * time.Millisecond)
			runtime.GC()
			time.Sleep(100 * time.Millisecond)

			finalCount := runtime.NumGoroutine()

			// Check for significant goroutine leaks
			// Allow for minimal variance in goroutine count for background goroutines
			Expect(finalCount).Should(BeNumerically("<=", initialCount+5), 
				"Potential goroutine leak detected: initial=%d, final=%d, increase=%d",
				initialCount, finalCount, finalCount-initialCount)
		})

		It("Should handle context cancellation properly", func() {
			testSchema := `
				entity user {}
				
				entity organization {
					relation admin @user
					relation member @user
				}
				
				entity repository {
					relation owner @user
					relation contributor @user
					relation parent @organization
					
					permission read = (owner or contributor) and parent.member
				}
			`

			db, err := factories.DatabaseFactory(config.Database{Engine: "memory"})
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(testSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			checkEngine := NewCheckEngine(schemaReader, dataReader)
			invoker := invoke.NewDirectInvoker(schemaReader, dataReader, checkEngine, nil, nil, nil)
			checkEngine.SetInvoker(invoker)

			// Create test data
			relationships := []string{
				"organization:1#admin@user:1",
				"organization:1#member@user:1",
				"repository:1#owner@user:1",
				"repository:1#parent@organization:1#...",
			}

			var tuples []*base.Tuple
			for _, relationship := range relationships {
				tpl, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, tpl)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			entity, err := tuple.E("repository:1")
			Expect(err).ShouldNot(HaveOccurred())

			ear, err := tuple.EAR("user:1")
			Expect(err).ShouldNot(HaveOccurred())

			subject := &base.Subject{
				Type:     ear.GetEntity().GetType(),
				Id:       ear.GetEntity().GetId(),
				Relation: ear.GetRelation(),
			}

			// Create a context that will be cancelled quickly
			ctx, cancel := context.WithCancel(context.Background())
			
			// Cancel the context immediately
			cancel()

			// This should return a cancellation error
			_, err = invoker.Check(ctx, &base.PermissionCheckRequest{
				TenantId:   "t1",
				Entity:     entity,
				Subject:    subject,
				Permission: "read",
				Metadata: &base.PermissionCheckRequestMetadata{
					SnapToken:     token.NewNoopToken().Encode().String(),
					SchemaVersion: "",
					Depth:         20,
				},
			})

			// We expect an error due to context cancellation
			Expect(err).Should(HaveOccurred())
		})

		It("Should handle high concurrency properly", func() {
			testSchema := `
				entity user {}
				
				entity organization {
					relation admin @user
					relation member @user
				}
				
				entity repository {
					relation owner @user
					relation parent @organization
					
					permission read = owner and parent.member
					permission write = owner and parent.admin
				}
			`

			db, err := factories.DatabaseFactory(config.Database{Engine: "memory"})
			Expect(err).ShouldNot(HaveOccurred())

			conf, err := newSchema(testSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			// Create CheckEngine with limited concurrency
			checkEngine := NewCheckEngine(schemaReader, dataReader, CheckConcurrencyLimit(5))
			invoker := invoke.NewDirectInvoker(schemaReader, dataReader, checkEngine, nil, nil, nil)
			checkEngine.SetInvoker(invoker)

			// Create test data
			relationships := []string{
				"organization:1#admin@user:1",
				"organization:1#member@user:1",
				"repository:1#owner@user:1",
				"repository:1#parent@organization:1#...",
			}

			var tuples []*base.Tuple
			for _, relationship := range relationships {
				tpl, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, tpl)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			entity, err := tuple.E("repository:1")
			Expect(err).ShouldNot(HaveOccurred())

			ear, err := tuple.EAR("user:1")
			Expect(err).ShouldNot(HaveOccurred())

			subject := &base.Subject{
				Type:     ear.GetEntity().GetType(),
				Id:       ear.GetEntity().GetId(),
				Relation: ear.GetRelation(),
			}

			// Record initial goroutine count
			runtime.GC()
			time.Sleep(100 * time.Millisecond)
			initialCount := runtime.NumGoroutine()

			// Test with high number of concurrent requests
			numRequests := 30
			errors := make(chan error, numRequests)
			results := make(chan *base.PermissionCheckResponse, numRequests)

			for i := 0; i < numRequests; i++ {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					resp, err := invoker.Check(ctx, &base.PermissionCheckRequest{
						TenantId:   "t1",
						Entity:     entity,
						Subject:    subject,
						Permission: "read",
						Metadata: &base.PermissionCheckRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
							Depth:         20,
						},
					})
					
					if err != nil {
						errors <- err
					} else {
						results <- resp
					}
				}()
			}

			// Collect all results
			successCount := 0
			errorCount := 0
			for i := 0; i < numRequests; i++ {
				select {
				case <-results:
					successCount++
				case <-errors:
					errorCount++
				case <-time.After(15 * time.Second):
					Fail("Test timed out waiting for results")
				}
			}

			Expect(successCount).Should(BeNumerically(">", 0))

			// Allow cleanup time
			time.Sleep(500 * time.Millisecond)
			runtime.GC()
			time.Sleep(100 * time.Millisecond)

			finalCount := runtime.NumGoroutine()

			// Check for goroutine leaks
			Expect(finalCount).Should(BeNumerically("<=", initialCount+10))
		})
	})
})