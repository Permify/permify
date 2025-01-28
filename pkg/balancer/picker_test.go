package balancer

import (
	"context"
	"crypto/sha256"
	"encoding/binary"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/balancer"

	"github.com/Permify/permify/pkg/consistent"
)

type mockConnection struct {
	id string
}

type testMember struct {
	name string
	conn *mockConnection
}

func (m testMember) String() string {
	return m.name
}

func (m testMember) Connection() *mockConnection {
	return m.conn
}

var _ = Describe("Picker and Consistent Hashing", func() {
	var (
		c           *consistent.Consistent
		testMembers []testMember
		hasher      func(data []byte) uint64
	)

	// Custom hasher using SHA-256 for consistent hashing
	hasher = func(data []byte) uint64 {
		hash := sha256.Sum256(data)
		return binary.BigEndian.Uint64(hash[:8]) // Use the first 8 bytes as the hash
	}

	BeforeEach(func() {
		// Initialize consistent hashing with a valid hasher
		c = consistent.New(consistent.Config{
			Hasher:            hasher,
			PartitionCount:    100,
			ReplicationFactor: 2,
			Load:              1.5,
		})

		// Add test members to the consistent hash ring
		testMembers = []testMember{
			{name: "member1", conn: &mockConnection{id: "conn1"}},
			{name: "member2", conn: &mockConnection{id: "conn2"}},
			{name: "member3", conn: &mockConnection{id: "conn3"}},
		}
		for _, m := range testMembers {
			c.Add(m)
		}
	})

	Describe("Picker Logic", func() {
		var (
			p       *picker
			testCtx context.Context
		)

		BeforeEach(func() {
			// Initialize picker with consistent hashing and a width of 2
			p = &picker{
				consistent: c,
				width:      2,
			}
			// Set up context with a valid key
			testCtx = context.WithValue(context.Background(), Key, []byte("test-key"))
		})

		It("should pick a member successfully", func() {
			// Mock picker behavior
			members, err := c.ClosestN([]byte("test-key"), 2)
			Expect(err).To(BeNil())
			Expect(len(members)).To(BeNumerically(">", 0))
			Expect(members[0].String()).To(Equal("member1"))
		})

		It("should return an error if the context key is missing", func() {
			result, err := p.Pick(balancer.PickInfo{Ctx: context.Background()})
			Expect(err).To(MatchError("context key missing"))
			Expect(result.SubConn).To(BeNil())
		})

		It("should return an error if no members are available", func() {
			// Remove all members
			for _, m := range testMembers {
				c.Remove(m.String())
			}
			result, err := p.Pick(balancer.PickInfo{Ctx: testCtx})
			Expect(err).To(MatchError("failed to get closest members: not enough members to satisfy the request"))
			Expect(result.SubConn).To(BeNil())
		})
	})
})
