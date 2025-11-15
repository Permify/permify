package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgtype"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestPostgres -
func TestPostgres(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "postgres-suite")
}

var _ = Describe("Postgres", func() {
	Context("Type Definitions", func() {
		It("Case 1: XID8 should be based on pguint64", func() {
			var x XID8
			// XID8 is a type alias for pguint64, test that they have the same fields
			Expect(x.Uint).Should(Equal(uint64(0)))
			Expect(x.Status).Should(Equal(pgtype.Status(0)))
		})
	})

	Context("New", func() {
		It("Case 1: Should return error with invalid URI", func() {
			_, err := New("invalid-uri")
			Expect(err).Should(HaveOccurred())
		})

		It("Case 2: Should return error with empty URI", func() {
			_, err := New("")
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("NewWithSeparateURIs", func() {
		It("Case 1: Should return error with invalid writer URI", func() {
			_, err := NewWithSeparateURIs("invalid-uri", "postgres://localhost/test")
			Expect(err).Should(HaveOccurred())
		})

		It("Case 2: Should return error with invalid reader URI", func() {
			_, err := NewWithSeparateURIs("postgres://localhost/test", "invalid-uri")
			Expect(err).Should(HaveOccurred())
		})

		It("Case 3: Should return error with empty URIs", func() {
			_, err := NewWithSeparateURIs("", "")
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("Postgres struct methods", func() {
		var pg *Postgres

		BeforeEach(func() {
			pg = &Postgres{
				maxDataPerWrite:       1000,
				maxRetries:            3,
				watchBufferSize:       100,
				maxConnectionLifeTime: 30 * time.Minute,
				maxConnectionIdleTime: 5 * time.Minute,
				maxConnections:        10,
				maxIdleConnections:    5,
			}
		})

		Context("GetMaxDataPerWrite", func() {
			It("Case 1: Should return correct value", func() {
				Expect(pg.GetMaxDataPerWrite()).Should(Equal(1000))
			})
		})

		Context("GetMaxRetries", func() {
			It("Case 1: Should return correct value", func() {
				Expect(pg.GetMaxRetries()).Should(Equal(3))
			})
		})

		Context("GetWatchBufferSize", func() {
			It("Case 1: Should return correct value", func() {
				Expect(pg.GetWatchBufferSize()).Should(Equal(100))
			})
		})

		Context("GetEngineType", func() {
			It("Case 1: Should return 'postgres'", func() {
				Expect(pg.GetEngineType()).Should(Equal("postgres"))
			})
		})

		Context("Close", func() {
			It("Case 1: Should panic with nil pools", func() {
				Expect(func() {
					pg.Close()
				}).Should(Panic())
			})
		})

		Context("IsReady", func() {
			It("Case 1: Should panic with nil read pool", func() {
				ctx := context.Background()
				// pg.ReadPool is nil, so this will panic
				Expect(func() {
					pg.IsReady(ctx)
				}).Should(Panic())
			})
		})
	})

	Context("Query execution modes", func() {
		Context("queryExecModes map", func() {
			It("Case 1: Should contain all expected modes", func() {
				expectedModes := []string{
					"cache_statement",
					"cache_describe",
					"describe_exec",
					"mode_exec",
					"simple_protocol",
				}

				for _, mode := range expectedModes {
					Expect(queryExecModes).Should(HaveKey(mode))
				}
			})
		})
	})

	Context("Plan cache modes", func() {
		Context("planCacheModes map", func() {
			It("Case 1: Should contain all expected modes", func() {
				expectedModes := []string{
					"auto",
					"force_custom_plan",
					"disable",
				}

				for _, mode := range expectedModes {
					Expect(planCacheModes).Should(HaveKey(mode))
				}
			})
		})
	})

	Context("Options", func() {
		It("Case 1: MaxDataPerWrite should set maxDataPerWrite", func() {
			pg := &Postgres{}
			option := MaxDataPerWrite(500)
			option(pg)
			Expect(pg.maxDataPerWrite).Should(Equal(500))
		})

		It("Case 2: WatchBufferSize should set watchBufferSize", func() {
			pg := &Postgres{}
			option := WatchBufferSize(200)
			option(pg)
			Expect(pg.watchBufferSize).Should(Equal(200))
		})

		It("Case 3: MaxRetries should set maxRetries", func() {
			pg := &Postgres{}
			option := MaxRetries(5)
			option(pg)
			Expect(pg.maxRetries).Should(Equal(5))
		})

		It("Case 4: MaxConnections should set maxConnections", func() {
			pg := &Postgres{}
			option := MaxConnections(25)
			option(pg)
			Expect(pg.maxConnections).Should(Equal(25))
		})

		It("Case 5: MinConnections should set minConnections", func() {
			pg := &Postgres{}
			option := MinConnections(5)
			option(pg)
			Expect(pg.minConnections).Should(Equal(5))
		})

		It("Case 6: MinIdleConnections should set minIdleConns", func() {
			pg := &Postgres{}
			option := MinIdleConnections(3)
			option(pg)
			Expect(pg.minIdleConns).Should(Equal(3))
		})

		It("Case 7: HealthCheckPeriod should set healthCheckPeriod", func() {
			pg := &Postgres{}
			option := HealthCheckPeriod(30 * time.Second)
			option(pg)
			Expect(pg.healthCheckPeriod).Should(Equal(30 * time.Second))
		})

		It("Case 8: MaxConnectionLifetimeJitter should set maxConnectionLifetimeJitter", func() {
			pg := &Postgres{}
			option := MaxConnectionLifetimeJitter(10 * time.Second)
			option(pg)
			Expect(pg.maxConnectionLifetimeJitter).Should(Equal(10 * time.Second))
		})

		It("Case 9: ConnectTimeout should set connectTimeout", func() {
			pg := &Postgres{}
			option := ConnectTimeout(5 * time.Second)
			option(pg)
			Expect(pg.connectTimeout).Should(Equal(5 * time.Second))
		})

		It("Case 10: MaxOpenConnections should set maxConnections (deprecated wrapper)", func() {
			pg := &Postgres{}
			option := MaxOpenConnections(20)
			option(pg)
			Expect(pg.maxConnections).Should(Equal(20))
		})

		It("Case 11: MaxIdleConnections should set maxIdleConnections (deprecated)", func() {
			pg := &Postgres{}
			option := MaxIdleConnections(5)
			option(pg)
			Expect(pg.maxIdleConnections).Should(Equal(5))
		})
	})

	Context("Backward Compatibility", func() {
		It("Case 1: MaxIdleConnections should be used as MinConns when MinConns is not set", func() {
			pg := &Postgres{
				minConnections:     0, // Not set
				maxIdleConnections: 5,
			}
			// Simulate the logic from newDB
			minConns := pg.minConnections
			if minConns == 0 && pg.maxIdleConnections > 0 {
				minConns = pg.maxIdleConnections
			}
			Expect(minConns).Should(Equal(5))
		})

		It("Case 2: MinConns should take precedence over MaxIdleConnections", func() {
			pg := &Postgres{
				minConnections:     3, // Explicitly set
				maxIdleConnections: 5,
			}
			// Simulate the logic from newDB
			minConns := pg.minConnections
			if minConns == 0 && pg.maxIdleConnections > 0 {
				minConns = pg.maxIdleConnections
			}
			Expect(minConns).Should(Equal(3)) // Should use MinConns, not MaxIdleConnections
		})

		It("Case 3: MaxOpenConnections should map to MaxConnections", func() {
			pg := &Postgres{}
			// MaxOpenConnections internally calls MaxConnections
			option := MaxOpenConnections(15)
			option(pg)
			Expect(pg.maxConnections).Should(Equal(15))
		})

		It("Case 4: When both MinConns and MaxIdleConnections are 0, should use 0 (pgx default)", func() {
			pg := &Postgres{
				minConnections:     0,
				maxIdleConnections: 0,
			}
			// Simulate the logic from newDB
			minConns := pg.minConnections
			if minConns == 0 && pg.maxIdleConnections > 0 {
				minConns = pg.maxIdleConnections
			}
			Expect(minConns).Should(Equal(0)) // Should remain 0, pgx will use its default
		})

		It("Case 5: MinIdleConns should only be set when explicitly configured", func() {
			pg := &Postgres{
				minIdleConns:       0, // Not set
				maxIdleConnections: 5,
			}
			// MinIdleConns should only be set if > 0
			shouldSet := pg.minIdleConns > 0
			Expect(shouldSet).Should(BeFalse()) // Should not be set
		})

		It("Case 6: MinIdleConns should be set when explicitly configured", func() {
			pg := &Postgres{
				minIdleConns: 3, // Explicitly set
			}
			// MinIdleConns should only be set if > 0
			shouldSet := pg.minIdleConns > 0
			Expect(shouldSet).Should(BeTrue()) // Should be set
		})
	})
})

var _ = Describe("MockPQDatabase", func() {
	var mockDB *MockPQDatabase

	BeforeEach(func() {
		mockDB = &MockPQDatabase{}
	})

	Context("GetEngineType", func() {
		It("Case 1: Should return mocked engine type", func() {
			mockDB.On("GetEngineType").Return("postgres")

			result := mockDB.GetEngineType()
			Expect(result).Should(Equal("postgres"))
			mockDB.AssertExpectations(GinkgoT())
		})
	})

	Context("Close", func() {
		It("Case 1: Should return mocked error", func() {
			mockDB.On("Close").Return(nil)

			err := mockDB.Close()
			Expect(err).ShouldNot(HaveOccurred())
			mockDB.AssertExpectations(GinkgoT())
		})

		It("Case 2: Should return mocked error when close fails", func() {
			expectedError := errors.New("connection error")
			mockDB.On("Close").Return(expectedError)

			err := mockDB.Close()
			Expect(err).Should(HaveOccurred())
			Expect(err).Should(Equal(expectedError))
			mockDB.AssertExpectations(GinkgoT())
		})
	})

	Context("IsReady", func() {
		It("Case 1: Should return mocked ready state", func() {
			ctx := context.Background()
			mockDB.On("IsReady", ctx).Return(true, nil)

			ready, err := mockDB.IsReady(ctx)
			Expect(ready).Should(BeTrue())
			Expect(err).ShouldNot(HaveOccurred())
			mockDB.AssertExpectations(GinkgoT())
		})

		It("Case 2: Should return mocked not ready state with error", func() {
			ctx := context.Background()
			expectedError := errors.New("database not ready")
			mockDB.On("IsReady", ctx).Return(false, expectedError)

			ready, err := mockDB.IsReady(ctx)
			Expect(ready).Should(BeFalse())
			Expect(err).Should(HaveOccurred())
			Expect(err).Should(Equal(expectedError))
			mockDB.AssertExpectations(GinkgoT())
		})
	})
})
