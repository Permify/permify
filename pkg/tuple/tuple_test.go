package tuple

import (
	`testing`

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
				target   EntityAndRelation
				expected string
			}{
				{EntityAndRelation{
					Entity: Entity{
						Type: "repository",
						ID:   "1",
					},
					Relation: "admin",
				}, "repository:1#admin"},
				{EntityAndRelation{
					Entity: Entity{
						Type: "doc",
						ID:   "1",
					},
					Relation: "viewer",
				}, "doc:1#viewer"},
			}

			for _, tt := range tests {
				Expect(tt.target.String()).Should(Equal(tt.expected))
			}
		})

	})

	Context("Relation", func() {

		It("Split", func() {

			tests := []struct {
				target   Relation
				expected []Relation
			}{
				{"parent.admin", []Relation{
					"parent", "admin",
				}},
				{"owner", []Relation{
					"owner", "",
				}},
				{"parent.parent.admin", []Relation{
					"parent", "parent", "admin",
				}},
			}

			for _, tt := range tests {
				Expect(tt.target.Split()).Should(Equal(tt.expected))
			}
		})

		It("IsComputed", func() {

			tests := []struct {
				target   Relation
				expected bool
			}{
				{"parent.admin", false},
				{"owner", true},
				{"parent.parent.admin", false},
				{"member", true},
			}

			for _, tt := range tests {
				Expect(tt.target.IsComputed()).Should(Equal(tt.expected))
			}
		})

	})

	Context("Subject", func() {

		It("Equals", func() {

			tests := []struct {
				target   Subject
				v        interface{}
				expected bool
			}{
				{target: Subject{
					Type: "user",
					ID:   "1",
				}, v: Subject{
					Type: "user",
					ID:   "1",
				}, expected: true},
				{target: Subject{
					Type: "organization",
					ID:   "1",
				}, v: Subject{
					Type:     "organization",
					ID:       "1",
					Relation: "admin",
				}, expected: false},
				{target: Subject{
					Type:     "organization",
					ID:       "1",
					Relation: "member",
				}, v: Subject{
					Type:     "organization",
					ID:       "1",
					Relation: "member",
				}, expected: true},
			}

			for _, tt := range tests {
				Expect(tt.target.Equals(tt.v)).Should(Equal(tt.expected))
			}
		})

		It("IsValid", func() {

			tests := []struct {
				target   Subject
				expected bool
			}{
				{
					target: Subject{
						Type:     "",
						ID:       "1",
						Relation: "",
					},
					expected: false,
				},
				{
					target: Subject{
						Type:     USER,
						ID:       "1",
						Relation: "",
					},
					expected: true,
				},
				{
					target: Subject{
						Type:     USER,
						ID:       "1",
						Relation: "admin",
					},
					expected: false,
				},
				{
					target: Subject{
						Type:     USER,
						ID:       "1",
						Relation: "admin",
					},
					expected: false,
				},
				{
					target: Subject{
						Type:     "organization",
						ID:       "1",
						Relation: "admin",
					},
					expected: true,
				},
			}

			for _, tt := range tests {
				Expect(tt.target.IsValid()).Should(Equal(tt.expected))
			}
		})

	})

})
