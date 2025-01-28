package consistent

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This is the entry point for the test suite for the "consistent" package.
// It registers a failure handler and runs the specifications (specs) for this package.
func TestConsistent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "consistent-suite")
}

type TestMember string

func (m TestMember) String() string {
	return string(m)
}

var _ = Describe("Consistent", func() {
	var (
		config     Config
		consist    *Consistent
		hasher     Hasher
		testMember Member
	)

	BeforeEach(func() {
		hasher = func(data []byte) uint64 {
			var hash uint64
			for _, b := range data {
				hash = hash*31 + uint64(b)
			}
			return hash
		}

		config = Config{
			Hasher:            hasher,
			PartitionCount:    271,
			ReplicationFactor: 20,
			Load:              1.25,
			PickerWidth:       1,
		}

		consist = New(config)
	})

	Describe("Initialization", func() {
		It("should initialize with default values when config is incomplete", func() {
			incompleteConfig := Config{Hasher: hasher}
			instance := New(incompleteConfig)
			Expect(instance).NotTo(BeNil())
			Expect(instance.GetAverageLoad()).To(BeNumerically("==", float64(0)))
		})
	})

	Describe("Member Management", func() {
		BeforeEach(func() {
			testMember = TestMember("member1")
		})

		It("should add a member to the consistent hash ring", func() {
			consist.Add(testMember)
			members := consist.Members()
			Expect(members).To(HaveLen(1))
			Expect(members[0].String()).To(Equal(testMember.String()))
		})

		It("should not add the same member twice", func() {
			consist.Add(testMember)
			consist.Add(testMember)
			members := consist.Members()
			Expect(members).To(HaveLen(1))
		})

		It("should remove a member from the consistent hash ring", func() {
			consist.Add(testMember)
			consist.Remove(testMember.String())
			members := consist.Members()
			Expect(members).To(BeEmpty())
		})
	})

	Describe("Partition Management", func() {
		BeforeEach(func() {
			consist.Add(TestMember("member1"))
			consist.Add(TestMember("member2"))
			consist.Add(TestMember("member3"))
		})

		It("should distribute partitions across members without exceeding average load", func() {
			loads := consist.GetLoadDistribution()
			Expect(loads).To(HaveLen(3)) // There are 3 members.

			// Calculate the maximum allowed load per member.
			expectedAverageLoad := consist.GetAverageLoad()

			// Verify each member's load does not exceed the expected average load.
			for _, load := range loads {
				Expect(load).To(BeNumerically("<=", expectedAverageLoad))
			}
		})

		It("should locate the correct partition owner for a key", func() {
			key := []byte("test_key")
			member := consist.LocateKey(key)
			Expect(member).NotTo(BeNil())
		})

		It("should find the closest N members", func() {
			key := []byte("test_key")
			members, err := consist.ClosestN(key, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(2))
		})
	})

	Describe("Error Handling", func() {
		It("should return an error when requesting more members than available", func() {
			key := []byte("test_key")
			_, err := consist.ClosestN(key, 5)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not enough members to satisfy the request"))
		})
	})
})
