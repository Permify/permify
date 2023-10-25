package validation

import (
	"testing"

	"google.golang.org/protobuf/types/known/anypb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TestValidation - Test suite for validation package
func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "validation-suite")
}

var _ = Describe("validation", func() {
	Context("Statement", func() {
		It("Case 1", func() {
			// Create a test entity definition
			entityDef := &base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
					"admin": {
						Name: "admin",
						RelationReferences: []*base.RelationReference{
							{
								Type: "user",
							},
						},
					},
				},
			}

			// Create a valid test tuple
			validTuple := &base.Tuple{
				Subject: &base.Subject{
					Type:     "user",
					Id:       "y",
					Relation: "",
				},
				Relation: "admin",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "x",
				},
			}

			// Create an invalid test tuple with subject of wrong type
			invalidTuple1 := &base.Tuple{
				Subject: &base.Subject{
					Type:     "WrongType",
					Id:       "x",
					Relation: "",
				},
				Relation: "TestRelation",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "y",
				},
			}

			// Create an invalid test tuple with relation not defined in entity definition
			invalidTuple2 := &base.Tuple{
				Subject: &base.Subject{
					Type:     "organization",
					Id:       "x",
					Relation: "member",
				},
				Relation: "admin",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "y",
				},
			}

			// Test the function with a valid tuple
			err := ValidateTuple(entityDef, validTuple)
			Expect(err).Should(BeNil())

			// Test the function with an invalid tuple with wrong subject type
			err = ValidateTuple(entityDef, invalidTuple1)
			Expect(err).ShouldNot(BeNil())

			// Test the function with an invalid tuple with relation not defined in entity definition
			err = ValidateTuple(entityDef, invalidTuple2)
			Expect(err).ShouldNot(BeNil())
		})

		It("Case 2", func() {
			// Create a test entity definition
			entityDef := &base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
					"admin": {
						Name: "admin",
						RelationReferences: []*base.RelationReference{
							{
								Type: "user",
							},
						},
					},
					"member": {
						Name: "member",
						RelationReferences: []*base.RelationReference{
							{
								Type: "user",
							},
						},
					},
				},
			}

			// Create a valid test tuple
			validTuple := &base.Tuple{
				Subject: &base.Subject{
					Type: "user",
					Id:   "y",
				},
				Relation: "admin",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "x",
				},
			}

			// Create an invalid test tuple with subject of wrong type
			invalidTuple1 := &base.Tuple{
				Subject: &base.Subject{
					Type: "WrongType",
					Id:   "x",
				},
				Relation: "TestRelation",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "y",
				},
			}

			// Create an invalid test tuple with relation not defined in entity definition
			invalidTuple2 := &base.Tuple{
				Subject: &base.Subject{
					Type:     "organization",
					Id:       "x",
					Relation: "member",
				},
				Relation: "admin",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "y",
				},
			}

			// Test the function with a valid tuple
			err := ValidateTuple(entityDef, validTuple)
			Expect(err).Should(BeNil())

			// Test the function with an invalid tuple with wrong subject type
			err = ValidateTuple(entityDef, invalidTuple1)
			Expect(err).ShouldNot(BeNil())

			// Test the function with an invalid tuple with relation not defined in entity definition
			err = ValidateTuple(entityDef, invalidTuple2)
			Expect(err).ShouldNot(BeNil())
		})

		It("Case 3", func() {
			// Create a test entity definition
			entityDef := &base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
					"admin": {
						Name: "admin",
						RelationReferences: []*base.RelationReference{
							{
								Type:     "team",
								Relation: "admin",
							},
							{
								Type:     "organization",
								Relation: "member",
							},
						},
					},
					"member": {
						Name: "member",
						RelationReferences: []*base.RelationReference{
							{
								Type: "user",
							},
							{
								Type:     "team",
								Relation: "member",
							},
						},
					},
				},
			}

			// Create a valid test tuple
			validTuple1 := &base.Tuple{
				Subject: &base.Subject{
					Type:     "team",
					Id:       "y",
					Relation: "admin",
				},
				Relation: "admin",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "x",
				},
			}

			// Create a valid test tuple
			validTuple2 := &base.Tuple{
				Subject: &base.Subject{
					Type:     "organization",
					Id:       "y",
					Relation: "member",
				},
				Relation: "admin",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "x",
				},
			}

			// Create a valid test tuple
			validTuple3 := &base.Tuple{
				Subject: &base.Subject{
					Type:     "team",
					Id:       "y",
					Relation: "member",
				},
				Relation: "member",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "x",
				},
			}

			// Create a valid test tuple
			validTuple4 := &base.Tuple{
				Subject: &base.Subject{
					Type: "user",
					Id:   "y",
				},
				Relation: "member",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "x",
				},
			}

			// Create an invalid test tuple with subject of wrong type
			invalidTuple1 := &base.Tuple{
				Subject: &base.Subject{
					Type: "user",
					Id:   "x",
				},
				Relation: "admin",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "x",
				},
			}

			// Create an invalid test tuple with relation not defined in entity definition
			invalidTuple2 := &base.Tuple{
				Subject: &base.Subject{
					Type:     "team",
					Id:       "x",
					Relation: "member",
				},
				Relation: "admin",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "y",
				},
			}

			// Test the function with a valid tuple
			err := ValidateTuple(entityDef, validTuple1)
			Expect(err).Should(BeNil())

			// Test the function with a valid tuple
			err = ValidateTuple(entityDef, validTuple2)
			Expect(err).Should(BeNil())

			// Test the function with a valid tuple
			err = ValidateTuple(entityDef, validTuple3)
			Expect(err).Should(BeNil())

			// Test the function with a valid tuple
			err = ValidateTuple(entityDef, validTuple4)
			Expect(err).Should(BeNil())

			// Test the function with an invalid tuple with wrong subject type
			err = ValidateTuple(entityDef, invalidTuple1)
			Expect(err).ShouldNot(BeNil())

			// Test the function with an invalid tuple with relation not defined in entity definition
			err = ValidateTuple(entityDef, invalidTuple2)
			Expect(err).ShouldNot(BeNil())
		})

		It("Case 4", func() {
			// Create a test entity definition
			entityDef := &base.EntityDefinition{
				Name: "organization",
				Attributes: map[string]*base.AttributeDefinition{
					"public": {
						Name: "public",
						Type: base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN,
					},
					"balance": {
						Name: "balance",
						Type: base.AttributeType_ATTRIBUTE_TYPE_DOUBLE,
					},
					"ips": {
						Name: "ips",
						Type: base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY,
					},
				},
			}

			va1v, err := anypb.New(&base.BooleanValue{Data: true})
			Expect(err).ShouldNot(HaveOccurred())

			// Create a valid test attribute
			validAttribute1 := &base.Attribute{
				Value:     va1v,
				Attribute: "public",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "x",
				},
			}

			va2v, err := anypb.New(&base.DoubleValue{Data: 145.34})
			Expect(err).ShouldNot(HaveOccurred())

			// Create a valid test attribute
			validAttribute2 := &base.Attribute{
				Value:     va2v,
				Attribute: "balance",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "x",
				},
			}

			va3v, err := anypb.New(&base.StringArrayValue{Data: []string{"127.0.0.1", "127.0.0.2"}})
			Expect(err).ShouldNot(HaveOccurred())

			// Create a valid test attribute
			validAttribute3 := &base.Attribute{
				Value:     va3v,
				Attribute: "ips",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "x",
				},
			}

			// Create an invalid test tuple with subject of wrong type
			invalidAttribute1 := &base.Attribute{
				Value:     va3v,
				Attribute: "reference",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "x",
				},
			}

			// Create an invalid test tuple with relation not defined in entity definition
			invalidAttribute2 := &base.Attribute{
				Value:     va3v,
				Attribute: "public",
				Entity: &base.Entity{
					Type: "organization",
					Id:   "y",
				},
			}

			// Test the function with a valid tuple
			err = ValidateAttribute(entityDef, validAttribute1)
			Expect(err).Should(BeNil())

			// Test the function with a valid tuple
			err = ValidateAttribute(entityDef, validAttribute2)
			Expect(err).Should(BeNil())

			// Test the function with a valid tuple
			err = ValidateAttribute(entityDef, validAttribute3)
			Expect(err).Should(BeNil())

			// Test the function with an invalid tuple with wrong subject type
			err = ValidateAttribute(entityDef, invalidAttribute1)
			Expect(err).ShouldNot(BeNil())

			// Test the function with an invalid tuple with relation not defined in entity definition
			err = ValidateAttribute(entityDef, invalidAttribute2)
			Expect(err).ShouldNot(BeNil())
		})

		It("Case 5", func() {

			err := ValidateTupleFilter(&base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "organization",
					Ids:  []string{"1"},
				},
				Relation: "admin",
				Subject:  &base.SubjectFilter{},
			})
			Expect(err).ShouldNot(HaveOccurred())

			err = ValidateTupleFilter(&base.TupleFilter{
				Entity: &base.EntityFilter{
					Type: "",
					Ids:  []string{"1"},
				},
				Relation: "admin",
				Subject:  &base.SubjectFilter{},
			})
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_ENTITY_TYPE_REQUIRED.String()))
		})
	})
})
