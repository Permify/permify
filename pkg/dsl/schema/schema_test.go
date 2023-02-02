package schema

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TestSchema -
func TestSchema(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "schema-suite")
}

var _ = Describe("schema", func() {
	Context("NewSchema", func() {
		It("Case 1", func() {
			entities := make([]*base.EntityDefinition, 0, 1)

			entities = append(entities, &base.EntityDefinition{
				Name:       "user",
				Relations:  map[string]*base.RelationDefinition{},
				Actions:    map[string]*base.ActionDefinition{},
				References: map[string]base.EntityDefinition_RelationalReference{},
				Option:     map[string]string{},
			})

			Expect(NewSchema(entities...)).To(Equal(&base.IndexedSchema{
				EntityDefinitions: map[string]*base.EntityDefinition{
					"user": entities[0],
				},
				RelationDefinitions: map[string]*base.RelationDefinition{},
				ActionDefinitions:   map[string]*base.ActionDefinition{},
			}))
		})

		It("Case 2", func() {
			entities := make([]*base.EntityDefinition, 0, 1)

			entities = append(entities, &base.EntityDefinition{
				Name:       "user",
				Relations:  map[string]*base.RelationDefinition{},
				Actions:    map[string]*base.ActionDefinition{},
				References: map[string]base.EntityDefinition_RelationalReference{},
				Option:     map[string]string{},
			}, &base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
					"owner": {
						Name: "owner",
						EntityReference: &base.RelationReference{
							Name: "user",
						},
					},
					"admin": {
						Name: "admin",
						EntityReference: &base.RelationReference{
							Name: "user",
						},
					},
				},
				Actions: map[string]*base.ActionDefinition{
					"update": {
						Name: "update",
						Child: &base.Child{
							Type: &base.Child_Rewrite{
								Rewrite: &base.Rewrite{
									RewriteOperation: base.Rewrite_OPERATION_UNION,
									Children: []*base.Child{
										{
											Type: &base.Child_Leaf{
												Leaf: &base.Leaf{
													Exclusion: false,
													Type: &base.Leaf_ComputedUserSet{
														ComputedUserSet: &base.ComputedUserSet{
															Relation: "owner",
														},
													},
												},
											},
										},
										{
											Type: &base.Child_Leaf{
												Leaf: &base.Leaf{
													Exclusion: false,
													Type: &base.Leaf_ComputedUserSet{
														ComputedUserSet: &base.ComputedUserSet{
															Relation: "admin",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				References: map[string]base.EntityDefinition_RelationalReference{
					"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
					"update": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
				},
				Option: map[string]string{},
			})

			Expect(NewSchema(entities...)).To(Equal(&base.IndexedSchema{
				EntityDefinitions: map[string]*base.EntityDefinition{
					"user":         entities[0],
					"organization": entities[1],
				},
				RelationDefinitions: map[string]*base.RelationDefinition{
					"organization#owner": entities[1].Relations["owner"],
					"organization#admin": entities[1].Relations["admin"],
				},
				ActionDefinitions: map[string]*base.ActionDefinition{
					"organization#update": entities[1].Actions["update"],
				},
			}))
		})
	})
})
