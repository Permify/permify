package validation

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"

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
	})
})
