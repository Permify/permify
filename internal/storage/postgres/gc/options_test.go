package gc

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/testinstance"
)

var _ = Describe("GC Options", func() {
	var db *PQDatabase.Postgres

	BeforeEach(func() {
		version := "14"
		db = testinstance.PostgresDB(version).(*PQDatabase.Postgres)
	})

	AfterEach(func() {
		if db != nil {
			err := db.Close()
			Expect(err).ShouldNot(HaveOccurred())
		}
	})

	Context("Option Functions", func() {
		It("should set interval option correctly", func() {
			expectedInterval := 30 * time.Second

			gc := NewGC(db, Interval(expectedInterval))

			Expect(gc.interval).Should(Equal(expectedInterval))
		})

		It("should set window option correctly", func() {
			expectedWindow := 10 * time.Minute

			gc := NewGC(db, Window(expectedWindow))

			Expect(gc.window).Should(Equal(expectedWindow))
		})

		It("should set timeout option correctly", func() {
			expectedTimeout := 5 * time.Minute

			gc := NewGC(db, Timeout(expectedTimeout))

			Expect(gc.timeout).Should(Equal(expectedTimeout))
		})

		It("should apply multiple options correctly", func() {
			expectedInterval := 45 * time.Second
			expectedWindow := 15 * time.Minute
			expectedTimeout := 8 * time.Minute

			gc := NewGC(db,
				Interval(expectedInterval),
				Window(expectedWindow),
				Timeout(expectedTimeout),
			)

			Expect(gc.interval).Should(Equal(expectedInterval))
			Expect(gc.window).Should(Equal(expectedWindow))
			Expect(gc.timeout).Should(Equal(expectedTimeout))
		})

		It("should override options when multiple of same type are provided", func() {
			firstInterval := 30 * time.Second
			secondInterval := 60 * time.Second

			gc := NewGC(db,
				Interval(firstInterval),
				Interval(secondInterval),
			)

			Expect(gc.interval).Should(Equal(secondInterval))
		})

		It("should use default values when no options are provided", func() {
			gc := NewGC(db)

			Expect(gc.interval).Should(Equal(_defaultInterval))
			Expect(gc.window).Should(Equal(_defaultWindow))
			Expect(gc.timeout).Should(Equal(_defaultTimeout))
		})

		It("should handle zero duration values", func() {
			gc := NewGC(db,
				Interval(0),
				Window(0),
				Timeout(0),
			)

			Expect(gc.interval).Should(Equal(time.Duration(0)))
			Expect(gc.window).Should(Equal(time.Duration(0)))
			Expect(gc.timeout).Should(Equal(time.Duration(0)))
		})

		It("should handle negative duration values", func() {
			negativeDuration := -5 * time.Second

			gc := NewGC(db,
				Interval(negativeDuration),
				Window(negativeDuration),
				Timeout(negativeDuration),
			)

			Expect(gc.interval).Should(Equal(negativeDuration))
			Expect(gc.window).Should(Equal(negativeDuration))
			Expect(gc.timeout).Should(Equal(negativeDuration))
		})

		It("should handle very large duration values", func() {
			largeDuration := 24 * time.Hour

			gc := NewGC(db,
				Interval(largeDuration),
				Window(largeDuration),
				Timeout(largeDuration),
			)

			Expect(gc.interval).Should(Equal(largeDuration))
			Expect(gc.window).Should(Equal(largeDuration))
			Expect(gc.timeout).Should(Equal(largeDuration))
		})

		It("should maintain database reference when options are applied", func() {
			gc := NewGC(db,
				Interval(30*time.Second),
				Window(10*time.Minute),
				Timeout(5*time.Minute),
			)

			Expect(gc.database).Should(Equal(db))
		})

		It("should allow chaining of option functions", func() {
			interval := 30 * time.Second
			window := 10 * time.Minute
			timeout := 5 * time.Minute

			// Test that option functions can be chained
			intervalOption := Interval(interval)
			windowOption := Window(window)
			timeoutOption := Timeout(timeout)

			gc := NewGC(db, intervalOption, windowOption, timeoutOption)

			Expect(gc.interval).Should(Equal(interval))
			Expect(gc.window).Should(Equal(window))
			Expect(gc.timeout).Should(Equal(timeout))
		})

		It("should handle mixed option types in any order", func() {
			interval := 45 * time.Second
			window := 15 * time.Minute
			timeout := 8 * time.Minute

			// Test different orders of options
			gc1 := NewGC(db, Interval(interval), Window(window), Timeout(timeout))
			gc2 := NewGC(db, Timeout(timeout), Interval(interval), Window(window))
			gc3 := NewGC(db, Window(window), Timeout(timeout), Interval(interval))

			Expect(gc1.interval).Should(Equal(interval))
			Expect(gc1.window).Should(Equal(window))
			Expect(gc1.timeout).Should(Equal(timeout))

			Expect(gc2.interval).Should(Equal(interval))
			Expect(gc2.window).Should(Equal(window))
			Expect(gc2.timeout).Should(Equal(timeout))

			Expect(gc3.interval).Should(Equal(interval))
			Expect(gc3.window).Should(Equal(window))
			Expect(gc3.timeout).Should(Equal(timeout))
		})
	})

	Context("Option Function Behavior", func() {
		It("should return a function that modifies the GC struct", func() {
			interval := 30 * time.Second
			option := Interval(interval)

			// Verify that the option is a function
			Expect(option).ShouldNot(BeNil())

			// Create a GC instance and apply the option
			gc := NewGC(db)
			originalInterval := gc.interval

			// Apply the option
			option(gc)

			// Verify the interval was changed
			Expect(gc.interval).Should(Equal(interval))
			Expect(gc.interval).ShouldNot(Equal(originalInterval))
		})

		It("should not affect other GC instances when option is applied", func() {
			interval1 := 30 * time.Second
			interval2 := 60 * time.Second

			gc1 := NewGC(db, Interval(interval1))
			gc2 := NewGC(db, Interval(interval2))

			Expect(gc1.interval).Should(Equal(interval1))
			Expect(gc2.interval).Should(Equal(interval2))
			Expect(gc1.interval).ShouldNot(Equal(gc2.interval))
		})

		It("should allow reusing the same option function multiple times", func() {
			interval := 30 * time.Second
			option := Interval(interval)

			gc1 := NewGC(db)
			gc2 := NewGC(db)

			// Apply the same option to both instances
			option(gc1)
			option(gc2)

			Expect(gc1.interval).Should(Equal(interval))
			Expect(gc2.interval).Should(Equal(interval))
		})
	})
})
