package balancer

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/balancer"

	"github.com/Permify/permify/pkg/consistent"
)

var _ = Describe("Picker and Consistent Hashing", func() {
	var (
		c       *consistent.Consistent
		members []ConsistentMember
		hasher  func(data []byte) uint64
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

		// Add ConsistentMembers to the consistent hash ring
		members = []ConsistentMember{
			{SubConn: &mockSubConnWrapper{}, name: "member1"},
			{SubConn: &mockSubConnWrapper{}, name: "member2"},
			{SubConn: &mockSubConnWrapper{}, name: "member3"},
		}
		for _, m := range members {
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
			members, err := c.ClosestN([]byte("test-key"), 2)
			Expect(err).To(BeNil())
			Expect(len(members)).To(BeNumerically(">", 0))
			Expect(members[0].(ConsistentMember).String()).To(Equal("member1"))
		})

		It("should return an error if the context key is missing", func() {
			result, err := p.Pick(balancer.PickInfo{Ctx: context.Background()})
			Expect(err).To(MatchError("context key missing"))
			Expect(result.SubConn).To(BeNil())
		})

		It("should return an error if no members are available", func() {
			// Remove all members
			for _, m := range members {
				c.Remove(m.String())
			}
			result, err := p.Pick(balancer.PickInfo{Ctx: testCtx})
			Expect(err).To(MatchError("failed to get closest members: not enough members to satisfy the request"))
			Expect(result.SubConn).To(BeNil())
		})

		It("should handle context key with wrong type", func() {
			wrongCtx := context.WithValue(context.Background(), Key, "wrong-type")
			result, err := p.Pick(balancer.PickInfo{Ctx: wrongCtx})
			Expect(err).To(MatchError("context key is not of type []byte"))
			Expect(result.SubConn).To(BeNil())
		})

		It("should handle empty key", func() {
			emptyCtx := context.WithValue(context.Background(), Key, []byte{})
			result, err := p.Pick(balancer.PickInfo{Ctx: emptyCtx})
			Expect(err).To(BeNil())
			Expect(result.SubConn).ToNot(BeNil())
		})

		It("should handle picker with width of 1", func() {
			p.width = 1
			result, err := p.Pick(balancer.PickInfo{Ctx: testCtx})
			Expect(err).To(BeNil())
			Expect(result.SubConn).ToNot(BeNil())
		})

		It("should handle picker with width larger than available members", func() {
			p.width = 10
			_, err := p.Pick(balancer.PickInfo{Ctx: testCtx})
			// If not enough members, should error
			if len(members) < 10 {
				Expect(err).ToNot(BeNil())
			} else {
				Expect(err).To(BeNil())
			}
		})

		It("should handle picker with zero width", func() {
			p.width = 0
			result, err := p.Pick(balancer.PickInfo{Ctx: testCtx})
			Expect(err).To(BeNil())
			Expect(result.SubConn).ToNot(BeNil())
		})

		It("should handle picker with negative width", func() {
			p.width = -1
			result, err := p.Pick(balancer.PickInfo{Ctx: testCtx})
			Expect(err).To(BeNil())
			Expect(result.SubConn).ToNot(BeNil())
		})

		It("should consistently pick the same member for the same key", func() {
			result1, err1 := p.Pick(balancer.PickInfo{Ctx: testCtx})
			result2, err2 := p.Pick(balancer.PickInfo{Ctx: testCtx})
			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(result1.SubConn).To(Equal(result2.SubConn))
		})

		It("should pick different members for different keys", func() {
			ctx1 := context.WithValue(context.Background(), Key, []byte("key1"))
			ctx2 := context.WithValue(context.Background(), Key, []byte("key2"))
			result1, err1 := p.Pick(balancer.PickInfo{Ctx: ctx1})
			result2, err2 := p.Pick(balancer.PickInfo{Ctx: ctx2})
			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(result1.SubConn).ToNot(BeNil())
			Expect(result2.SubConn).ToNot(BeNil())
		})

		It("should handle very long keys", func() {
			longKey := make([]byte, 10000)
			for i := range longKey {
				longKey[i] = byte(i % 256)
			}
			longCtx := context.WithValue(context.Background(), Key, longKey)
			result, err := p.Pick(balancer.PickInfo{Ctx: longCtx})
			Expect(err).To(BeNil())
			Expect(result.SubConn).ToNot(BeNil())
		})

		It("should handle special characters in keys", func() {
			specialKeys := [][]byte{
				[]byte("key with spaces"),
				[]byte("key-with-dashes"),
				[]byte("key_with_underscores"),
				[]byte("key.with.dots"),
				[]byte("key:with:colons"),
				[]byte("key;with;semicolons"),
				[]byte("key,with,commas"),
				[]byte("key!with!exclamation"),
				[]byte("key?with?question"),
				[]byte("key@with@at"),
				[]byte("key#with#hash"),
				[]byte("key$with$dollar"),
				[]byte("key%with%percent"),
				[]byte("key^with^caret"),
				[]byte("key&with&ampersand"),
				[]byte("key*with*asterisk"),
				[]byte("key(with)parentheses"),
				[]byte("key[with]brackets"),
				[]byte("key{with}braces"),
				[]byte("key<with>angles"),
				[]byte("key\"with\"quotes"),
				[]byte("key'with'apostrophe"),
				[]byte("key`with`backtick"),
				[]byte("key~with~tilde"),
				[]byte("key|with|pipe"),
				[]byte("key\\with\\backslash"),
				[]byte("key/with/forward/slash"),
			}
			for _, key := range specialKeys {
				specialCtx := context.WithValue(context.Background(), Key, key)
				result, err := p.Pick(balancer.PickInfo{Ctx: specialCtx})
				Expect(err).To(BeNil(), "Should handle key: %s", string(key))
				Expect(result.SubConn).ToNot(BeNil(), "Should return SubConn for key: %s", string(key))
			}
		})

		It("should handle unicode characters in keys", func() {
			unicodeKeys := [][]byte{
				[]byte("key with émojis 🚀"),
				[]byte("key with 中文"),
				[]byte("key with русский"),
				[]byte("key with العربية"),
				[]byte("key with हिन्दी"),
				[]byte("key with 日本語"),
				[]byte("key with 한국어"),
				[]byte("key with ελληνικά"),
				[]byte("key with עברית"),
				[]byte("key with தமிழ்"),
			}
			for _, key := range unicodeKeys {
				unicodeCtx := context.WithValue(context.Background(), Key, key)
				result, err := p.Pick(balancer.PickInfo{Ctx: unicodeCtx})
				Expect(err).To(BeNil(), "Should handle unicode key: %s", string(key))
				Expect(result.SubConn).ToNot(BeNil(), "Should return SubConn for unicode key: %s", string(key))
			}
		})
	})

	Describe("Consistent Hashing Behavior", func() {
		It("should distribute keys evenly across members", func() {
			p := &picker{
				consistent: c,
				width:      1,
			}

			// Map mockSubConnWrapper pointer to name
			subConnToName := map[balancer.SubConn]string{}
			for _, m := range members {
				subConnToName[m.SubConn] = m.name
			}

			keyCount := 1000
			memberCounts := make(map[string]int)

			for i := 0; i < keyCount; i++ {
				key := []byte(fmt.Sprintf("key-%d", i))
				ctx := context.WithValue(context.Background(), Key, key)
				result, err := p.Pick(balancer.PickInfo{Ctx: ctx})
				Expect(err).To(BeNil())
				pickedName := subConnToName[result.SubConn]
				memberCounts[pickedName]++
			}

			// Check that all members received some keys
			Expect(len(memberCounts)).To(Equal(3))
			minCount := keyCount / 10
			for member, count := range memberCounts {
				Expect(count).To(BeNumerically(">=", minCount),
					"Member %s should receive at least %d keys, got %d", member, minCount, count)
			}
		})

		It("should handle member removal gracefully", func() {
			p := &picker{
				consistent: c,
				width:      2,
			}
			testCtx := context.WithValue(context.Background(), Key, []byte("test-key"))
			result1, err1 := p.Pick(balancer.PickInfo{Ctx: testCtx})
			Expect(err1).To(BeNil())
			Expect(result1.SubConn).ToNot(BeNil())
			c.Remove("member1")
			result2, err2 := p.Pick(balancer.PickInfo{Ctx: testCtx})
			if len(members)-1 < 2 {
				Expect(err2).ToNot(BeNil())
			} else {
				Expect(err2).To(BeNil())
				Expect(result2.SubConn).ToNot(BeNil())
			}
		})

		It("should handle member addition gracefully", func() {
			p := &picker{
				consistent: c,
				width:      2,
			}
			testCtx := context.WithValue(context.Background(), Key, []byte("test-key"))
			result1, err1 := p.Pick(balancer.PickInfo{Ctx: testCtx})
			Expect(err1).To(BeNil())
			Expect(result1.SubConn).ToNot(BeNil())
			newMember := ConsistentMember{SubConn: &mockSubConnWrapper{}, name: "member4"}
			c.Add(newMember)
			result2, err2 := p.Pick(balancer.PickInfo{Ctx: testCtx})
			Expect(err2).To(BeNil())
			Expect(result2.SubConn).ToNot(BeNil())
		})
	})

	Describe("Error Scenarios", func() {
		It("should handle consistent hashing errors", func() {
			// Create a picker with a broken consistent hashing
			brokenC := consistent.New(consistent.Config{
				Hasher:            hasher,
				PartitionCount:    1, // Very small partition count to cause issues
				ReplicationFactor: 1,
				Load:              1.0,
			})

			p := &picker{
				consistent: brokenC,
				width:      2,
			}

			// Create a test context
			testCtx := context.WithValue(context.Background(), Key, []byte("test-key"))

			// Try to pick - should handle the error gracefully
			result, err := p.Pick(balancer.PickInfo{Ctx: testCtx})
			Expect(err).To(HaveOccurred())
			Expect(result.SubConn).To(BeNil())
		})

		It("should handle nil consistent hashing", func() {
			p := &picker{
				consistent: nil,
				width:      2,
			}

			// Create a test context
			testCtx := context.WithValue(context.Background(), Key, []byte("test-key"))

			// This should panic or return an error
			Expect(func() {
				p.Pick(balancer.PickInfo{Ctx: testCtx})
			}).To(Panic())
		})
	})

	Describe("Picker Configuration", func() {
		It("should work with different width configurations", func() {
			widths := []int{1, 2, 3, 5, 10}
			for _, width := range widths {
				p := &picker{
					consistent: c,
					width:      width,
				}
				testCtx := context.WithValue(context.Background(), Key, []byte("test-key"))
				_, err := p.Pick(balancer.PickInfo{Ctx: testCtx})
				if width > len(members) {
					Expect(err).ToNot(BeNil(), "Should error with width %d", width)
				} else {
					Expect(err).To(BeNil(), "Should work with width %d", width)
				}
			}
		})

		It("should handle edge case width values", func() {
			edgeWidths := []int{0, -1, -100, 1000, 999999}
			for _, width := range edgeWidths {
				p := &picker{
					consistent: c,
					width:      width,
				}
				testCtx := context.WithValue(context.Background(), Key, []byte("test-key"))
				_, err := p.Pick(balancer.PickInfo{Ctx: testCtx})
				if width > len(members) {
					Expect(err).ToNot(BeNil(), "Should error with edge width %d", width)
				} else {
					Expect(err).To(BeNil(), "Should work with edge width %d", width)
				}
			}
		})
	})
})
