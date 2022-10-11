package tuple

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TestTuple -
func TestTuple(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "tuple-suite")
}

var _ = Describe("tuple", func() {
	Context("EntityAndRelation", func() {
		It("String", func() {
			tests := []struct {
				target   *base.EntityAndRelation
				expected string
			}{
				{&base.EntityAndRelation{
					Entity: &base.Entity{
						Type: "repository",
						Id:   "1",
					},
					Relation: "admin",
				}, "repository:1#admin"},
				{&base.EntityAndRelation{
					Entity: &base.Entity{
						Type: "doc",
						Id:   "1",
					},
					Relation: "viewer",
				}, "doc:1#viewer"},
			}

			for _, tt := range tests {
				Expect(EntityAndRelationToString(tt.target)).Should(Equal(tt.expected))
			}
		})
	})

	Context("Relation", func() {
		It("Split", func() {
			tests := []struct {
				target   string
				expected []string
			}{
				{"parent.admin", []string{
					"parent", "admin",
				}},
				{"owner", []string{
					"owner", "",
				}},
				{"parent.parent.admin", []string{
					"parent", "parent", "admin",
				}},
			}

			for _, tt := range tests {
				Expect(SplitRelation(tt.target)).Should(Equal(tt.expected))
			}
		})

		It("IsComputed", func() {
			tests := []struct {
				target   string
				expected bool
			}{
				{"parent.admin", false},
				{"owner", true},
				{"parent.parent.admin", false},
				{"member", true},
			}

			for _, tt := range tests {
				Expect(IsRelationComputed(tt.target)).Should(Equal(tt.expected))
			}
		})
	})

	Context("Subject", func() {
		It("Equals", func() {
			tests := []struct {
				target   *base.Subject
				v        *base.Subject
				expected bool
			}{
				{target: &base.Subject{
					Type: "user",
					Id:   "1",
				}, v: &base.Subject{
					Type: "user",
					Id:   "1",
				}, expected: true},
				{target: &base.Subject{
					Type: "organization",
					Id:   "1",
				}, v: &base.Subject{
					Type:     "organization",
					Id:       "1",
					Relation: "admin",
				}, expected: false},
				{target: &base.Subject{
					Type:     "organization",
					Id:       "1",
					Relation: "member",
				}, v: &base.Subject{
					Type:     "organization",
					Id:       "1",
					Relation: "member",
				}, expected: true},
			}

			for _, tt := range tests {
				Expect(AreSubjectsEqual(tt.target, tt.v)).Should(Equal(tt.expected))
			}
		})

		It("IsValid", func() {
			tests := []struct {
				target   *base.Subject
				expected bool
			}{
				{
					target: &base.Subject{
						Type:     "",
						Id:       "1",
						Relation: "",
					},
					expected: false,
				},
				{
					target: &base.Subject{
						Type:     USER,
						Id:       "1",
						Relation: "",
					},
					expected: true,
				},
				{
					target: &base.Subject{
						Type:     USER,
						Id:       "1",
						Relation: "admin",
					},
					expected: false,
				},
				{
					target: &base.Subject{
						Type:     USER,
						Id:       "1",
						Relation: "admin",
					},
					expected: false,
				},
				{
					target: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: "admin",
					},
					expected: true,
				},
			}

			for _, tt := range tests {
				Expect(IsSubjectValid(tt.target)).Should(Equal(tt.expected))
			}
		})
	})
})
