package tuple

import (
	"errors"
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
	Context("TupleToString", func() {
		It("EntityAndRelationToString", func() {
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
				Expect(EntityAndRelationToString(tt.target.GetEntity(), tt.target.GetRelation())).Should(Equal(tt.expected))
			}
		})
	})

	Context("StringToTuple", func() {
		It("Tuple", func() {
			tests := []struct {
				target   string
				err      error
				expected *base.Tuple
			}{
				{
					target: "repository:1#admin@user:1",
					expected: &base.Tuple{
						Entity: &base.Entity{
							Type: "repository",
							Id:   "1",
						},
						Relation: "admin",
						Subject: &base.Subject{
							Type: "user",
							Id:   "1",
						},
					},
				},
				{
					target: "repository:1#parent@organization:1#...",
					expected: &base.Tuple{
						Entity: &base.Entity{
							Type: "repository",
							Id:   "1",
						},
						Relation: "parent",
						Subject: &base.Subject{
							Type:     "organization",
							Id:       "1",
							Relation: ELLIPSIS,
						},
					},
				},
				{
					target: "repository:1#admin@organization:1#member",
					expected: &base.Tuple{
						Entity: &base.Entity{
							Type: "repository",
							Id:   "1",
						},
						Relation: "admin",
						Subject: &base.Subject{
							Type:     "organization",
							Id:       "1",
							Relation: "member",
						},
					},
				},
				{
					target: "repository:1#wrong:1#member",
					err:    ErrInvalidTuple,
				},
			}

			for _, tt := range tests {
				e, err := Tuple(tt.target)
				if err != nil {
					Expect(err).Should(Equal(tt.err))
				} else {
					Expect(e).Should(Equal(tt.expected))
				}
			}
		})

		It("EAR", func() {
			tests := []struct {
				target   string
				err      error
				expected *base.EntityAndRelation
			}{
				{
					target: "repository:1#admin",
					expected: &base.EntityAndRelation{
						Entity: &base.Entity{
							Type: "repository",
							Id:   "1",
						},
						Relation: "admin",
					},
				},
				{
					target: "test:1#x",
					expected: &base.EntityAndRelation{
						Entity: &base.Entity{
							Type: "test",
							Id:   "1",
						},
						Relation: "x",
					},
				},
				{
					target: "test:5",
					expected: &base.EntityAndRelation{
						Entity: &base.Entity{
							Type: "test",
							Id:   "5",
						},
						Relation: "",
					},
				},
				{
					target:   "wrong",
					expected: nil,
					err:      ErrInvalidEntity,
				},
			}

			for _, tt := range tests {
				e, err := EAR(tt.target)
				if err != nil {
					Expect(err).Should(Equal(tt.err))
				} else {
					Expect(e).Should(Equal(tt.expected))
				}
			}
		})

		It("E", func() {
			tests := []struct {
				target   string
				err      error
				expected *base.Entity
			}{
				{
					target: "repository:1",
					expected: &base.Entity{
						Type: "repository",
						Id:   "1",
					},
				},
				{
					target: "test:4",
					expected: &base.Entity{
						Type: "test",
						Id:   "4",
					},
				},
				{
					target:   "wrong",
					expected: nil,
					err:      ErrInvalidEntity,
				},
				{
					target:   "wrong:wrong:wrong:wrong",
					expected: nil,
					err:      ErrInvalidEntity,
				},
			}

			for _, tt := range tests {
				e, err := E(tt.target)
				if err != nil {
					Expect(err).Should(Equal(tt.err))
				} else {
					Expect(e).Should(Equal(tt.expected))
				}
			}
		})
	})

	Context("Relation", func() {
		It("SplitRelation", func() {
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

		It("IsRelationComputed", func() {
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
						Type:     "user",
						Id:       "1",
						Relation: "",
					},
					expected: true,
				},
				{
					target: &base.Subject{
						Type:     "user",
						Id:       "1",
						Relation: "admin",
					},
					expected: true,
				},
				{
					target: &base.Subject{
						Type:     "user",
						Id:       "1",
						Relation: "admin",
					},
					expected: true,
				},
				{
					target: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: "admin",
					},
					expected: true,
				},
				{
					target: &base.Subject{
						Type:     "organization",
						Id:       "",
						Relation: "admin",
					},
					expected: false,
				},
			}

			for _, tt := range tests {
				Expect(IsSubjectValid(tt.target)).Should(Equal(tt.expected))
			}
		})

		It("ValidateSubjectType", func() {
			tests := []struct {
				target        *base.Subject
				relationTypes []string
				expected      error
			}{
				{
					target: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: "member",
					},
					relationTypes: []string{
						"organization#member",
						"user",
					},
					expected: nil,
				},
				{
					target: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: "",
					},
					relationTypes: []string{
						"organization",
					},
					expected: nil,
				},
				{
					target: &base.Subject{
						Type:     "user",
						Id:       "u82",
						Relation: "",
					},
					relationTypes: []string{
						"user",
					},
					expected: nil,
				},
				{
					target: &base.Subject{
						Type:     "testrel",
						Id:       "u82",
						Relation: "",
					},
					relationTypes: []string{
						"test",
						"user",
					},
					expected: errors.New(base.ErrorCode_ERROR_CODE_SUBJECT_TYPE_NOT_FOUND.String()),
				},
				{
					target: &base.Subject{
						Type:     "test",
						Id:       "u3",
						Relation: "mem",
					},
					relationTypes: []string{
						"test#member",
						"user",
					},
					expected: errors.New(base.ErrorCode_ERROR_CODE_SUBJECT_TYPE_NOT_FOUND.String()),
				},
				{
					target: &base.Subject{
						Type:     "test",
						Id:       "u3",
						Relation: "mem",
					},
					relationTypes: []string{},
					expected:      errors.New(base.ErrorCode_ERROR_CODE_SUBJECT_TYPE_NOT_FOUND.String()),
				},
			}

			for _, tt := range tests {
				if tt.expected == nil {
					Expect(ValidateSubjectType(tt.target, tt.relationTypes)).Should(BeNil())
				} else {
					Expect(ValidateSubjectType(tt.target, tt.relationTypes)).Should(Equal(tt.expected))
				}
			}
		})

		It("ToString", func() {
			tests := []struct {
				target *base.Tuple
				str    string
			}{
				{
					target: &base.Tuple{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Relation: "member",
						Subject: &base.Subject{
							Type: "user",
							Id:   "1",
						},
					},
					str: "organization:1#member@user:1",
				},
				{
					target: &base.Tuple{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Relation: "member",
						Subject: &base.Subject{
							Type:     "organization",
							Id:       "2",
							Relation: "admin",
						},
					},
					str: "organization:1#member@organization:2#admin",
				},
			}

			for _, tt := range tests {
				Expect(ToString(tt.target)).Should(Equal(tt.str))
			}
		})

		It("AreRelationReferencesEqual", func() {
			tests := []struct {
				target1 *base.RelationReference
				target2 *base.RelationReference
				result  bool
			}{
				{
					target1: &base.RelationReference{
						Type:     "organization",
						Relation: "member",
					},
					target2: &base.RelationReference{
						Type:     "organization",
						Relation: "member",
					},
					result: true,
				},
				{
					target1: &base.RelationReference{
						Type:     "organization",
						Relation: "member",
					},
					target2: &base.RelationReference{
						Type:     "organization",
						Relation: "member",
					},
					result: true,
				},
				{
					target1: &base.RelationReference{
						Type:     "organization",
						Relation: "member",
					},
					target2: &base.RelationReference{
						Type:     "organization",
						Relation: "admin",
					},
					result: false,
				},
			}

			for _, tt := range tests {
				Expect(AreRelationReferencesEqual(tt.target1, tt.target2)).Should(Equal(tt.result))
			}
		})

		It("RelationReference", func() {
			tests := []struct {
				target string
				result *base.RelationReference
			}{
				{
					target: "organization#member",
					result: &base.RelationReference{
						Type:     "organization",
						Relation: "member",
					},
				},
				{
					target: "repository",
					result: &base.RelationReference{
						Type:     "repository",
						Relation: "",
					},
				},
			}

			for _, tt := range tests {
				Expect(RelationReference(tt.target)).Should(Equal(tt.result))
			}
		})

		It("ReferenceToString", func() {
			tests := []struct {
				target *base.RelationReference
				result string
			}{
				{
					target: &base.RelationReference{
						Type:     "organization",
						Relation: "member",
					},
					result: "organization#member",
				},
				{
					target: &base.RelationReference{
						Type:     "repository",
						Relation: "",
					},
					result: "repository",
				},
			}

			for _, tt := range tests {
				Expect(ReferenceToString(tt.target)).Should(Equal(tt.result))
			}
		})

		It("ReferenceToString", func() {
			tests := []struct {
				target *base.RelationReference
				result string
			}{
				{
					target: &base.RelationReference{
						Type:     "organization",
						Relation: "member",
					},
					result: "organization#member",
				},
				{
					target: &base.RelationReference{
						Type:     "repository",
						Relation: "",
					},
					result: "repository",
				},
			}

			for _, tt := range tests {
				Expect(ReferenceToString(tt.target)).Should(Equal(tt.result))
			}
		})

		It("IsEntityAndSubjectEquals", func() {
			tests := []struct {
				target *base.Tuple
				result bool
			}{
				{
					target: &base.Tuple{
						Entity: &base.Entity{
							Type: "account",
							Id:   "1",
						},
						Relation: "member",
						Subject: &base.Subject{
							Type:     "account",
							Id:       "1",
							Relation: "member",
						},
					},
					result: true,
				},
				{
					target: &base.Tuple{
						Entity: &base.Entity{
							Type: "account",
							Id:   "1",
						},
						Relation: "member",
						Subject: &base.Subject{
							Type:     "account",
							Id:       "1",
							Relation: "admin",
						},
					},
					result: false,
				},
			}

			for _, tt := range tests {
				Expect(IsEntityAndSubjectEquals(tt.target)).Should(Equal(tt.result))
			}
		})

		It("NormalizeRelation", func() {
			tests := []struct {
				target string
				result string
			}{
				{
					target: "",
					result: "",
				},
				{
					target: "...",
					result: "",
				},
				{
					target: ELLIPSIS,
					result: "",
				},
			}

			for _, tt := range tests {
				Expect(NormalizeRelation(tt.target)).Should(Equal(tt.result))
			}
		})

		It("AreSubjectsEqual", func() {
			tests := []struct {
				target1 *base.Subject
				target2 *base.Subject
				result  bool
			}{
				{
					target1: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: "member",
					},
					target2: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: "member",
					},
					result: true,
				},
				{
					target1: &base.Subject{
						Type:     "repository",
						Id:       "3",
						Relation: "...",
					},
					target2: &base.Subject{
						Type:     "repository",
						Id:       "3",
						Relation: "",
					},
					result: true,
				},
				{
					target1: &base.Subject{
						Type:     "repository",
						Id:       "3",
						Relation: "...",
					},
					target2: &base.Subject{
						Type:     "repository",
						Id:       "4",
						Relation: "",
					},
					result: false,
				},
			}

			for _, tt := range tests {
				Expect(AreSubjectsEqual(tt.target1, tt.target2)).Should(Equal(tt.result))
			}
		})

		It("EAREqual", func() {
			tests := []struct {
				target1 *base.EntityAndRelation
				target2 *base.EntityAndRelation
				result  bool
			}{
				{
					target1: &base.EntityAndRelation{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Relation: "member",
					},
					target2: &base.EntityAndRelation{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Relation: "member",
					},
					result: true,
				},
				{
					target1: &base.EntityAndRelation{
						Entity: &base.Entity{
							Type: "repository",
							Id:   "3",
						},
						Relation: "admin",
					},
					target2: &base.EntityAndRelation{
						Entity: &base.Entity{
							Type: "repository",
							Id:   "3",
						},
						Relation: "member",
					},
					result: false,
				},
			}

			for _, tt := range tests {
				Expect(EAREqual(tt.target1, tt.target2)).Should(Equal(tt.result))
			}
		})

		It("SubjectToEAR", func() {
			tests := []struct {
				subject *base.Subject
				result  *base.EntityAndRelation
			}{
				{
					subject: &base.Subject{
						Type:     "organization",
						Id:       "1",
						Relation: "member",
					},
					result: &base.EntityAndRelation{
						Entity: &base.Entity{
							Type: "organization",
							Id:   "1",
						},
						Relation: "member",
					},
				},
			}

			for _, tt := range tests {
				Expect(SubjectToEAR(tt.subject)).Should(Equal(tt.result))
			}
		})
	})
})
