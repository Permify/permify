package balancer

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Queue", func() {
	Describe("Newly initialized queue", func() {
		var q *Queue
		BeforeEach(func() {
			q = NewQueue()
		})

		It("should not be nil", func() {
			Expect(q).ShouldNot(BeNil())
		})

		It("should be empty", func() {
			Expect(q.IsEmpty()).Should(BeTrue())
		})

		It("should have a length of 0", func() {
			Expect(q.Len()).Should(Equal(0))
		})
	})

	Describe("Queue operations", func() {
		var q *Queue
		BeforeEach(func() {
			q = NewQueue()
		})

		Context("EnQueue", func() {
			It("should add items to the queue", func() {
				q.EnQueue(1)
				Expect(q.IsEmpty()).Should(BeFalse())
				Expect(q.Len()).Should(Equal(1))
			})
		})

		Context("DeQueue", func() {
			It("should remove and return the first item in the queue", func() {
				q.EnQueue(1)
				q.EnQueue(2)

				val, ok := q.DeQueue()
				Expect(ok).Should(BeTrue())
				Expect(val).Should(Equal(1))

				val, ok = q.DeQueue()
				Expect(ok).Should(BeTrue())
				Expect(val).Should(Equal(2))

				val, ok = q.DeQueue()
				Expect(ok).Should(BeFalse())
				Expect(val).Should(BeNil())
			})
		})
	})
})
