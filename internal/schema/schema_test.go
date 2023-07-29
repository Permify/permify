package schema

import (
	"testing"

	"github.com/davecgh/go-spew/spew"

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
				Name:        "user",
				Relations:   map[string]*base.RelationDefinition{},
				Permissions: map[string]*base.PermissionDefinition{},
				References:  map[string]base.EntityDefinition_Reference{},
			})

			spew.Dump(NewSchemaFromEntityAndRuleDefinitions(entities, []*base.RuleDefinition{}))

			Expect(NewSchemaFromEntityAndRuleDefinitions(entities, []*base.RuleDefinition{})).To(Equal(&base.SchemaDefinition{
				EntityDefinitions: map[string]*base.EntityDefinition{
					"user": entities[0],
				},
				RuleDefinitions: map[string]*base.RuleDefinition{},
				References: map[string]base.SchemaDefinition_Reference{
					"user": base.SchemaDefinition_REFERENCE_ENTITY,
				},
			}))
		})

		It("Case 2", func() {
			entities := make([]*base.EntityDefinition, 0, 2)

			entities = append(entities, &base.EntityDefinition{
				Name:        "user",
				Relations:   map[string]*base.RelationDefinition{},
				Permissions: map[string]*base.PermissionDefinition{},
				References:  map[string]base.EntityDefinition_Reference{},
			}, &base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
					"owner": {
						Name: "owner",
						RelationReferences: []*base.RelationReference{
							{
								Type:     "user",
								Relation: "",
							},
						},
					},
					"admin": {
						Name: "admin",
						RelationReferences: []*base.RelationReference{
							{
								Type:     "user",
								Relation: "",
							},
						},
					},
				},
				Permissions: map[string]*base.PermissionDefinition{
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
				References: map[string]base.EntityDefinition_Reference{
					"owner":  base.EntityDefinition_REFERENCE_RELATION,
					"update": base.EntityDefinition_REFERENCE_PERMISSION,
				},
			})

			Expect(NewSchemaFromEntityAndRuleDefinitions(entities, []*base.RuleDefinition{})).To(Equal(&base.SchemaDefinition{
				EntityDefinitions: map[string]*base.EntityDefinition{
					"user":         entities[0],
					"organization": entities[1],
				},
				RuleDefinitions: map[string]*base.RuleDefinition{},
				References: map[string]base.SchemaDefinition_Reference{
					"user":         base.SchemaDefinition_REFERENCE_ENTITY,
					"organization": base.SchemaDefinition_REFERENCE_ENTITY,
				},
			}))
		})

		It("Case 3", func() {
			entities := make([]*base.EntityDefinition, 0, 3)

			entities = append(entities, &base.EntityDefinition{
				Name:        "user",
				Relations:   map[string]*base.RelationDefinition{},
				Permissions: map[string]*base.PermissionDefinition{},
				References:  map[string]base.EntityDefinition_Reference{},
			}, &base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
					"owner": {
						Name: "owner",
						RelationReferences: []*base.RelationReference{
							{
								Type:     "user",
								Relation: "",
							},
						},
					},
					"admin": {
						Name: "admin",
						RelationReferences: []*base.RelationReference{
							{
								Type:     "user",
								Relation: "",
							},
						},
					},
				},
				Permissions: map[string]*base.PermissionDefinition{
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
				References: map[string]base.EntityDefinition_Reference{
					"owner":  base.EntityDefinition_REFERENCE_RELATION,
					"update": base.EntityDefinition_REFERENCE_PERMISSION,
				},
			}, &base.EntityDefinition{
				Name: "repository",
				Relations: map[string]*base.RelationDefinition{
					"parent": {
						Name: "parent",
						RelationReferences: []*base.RelationReference{
							{
								Type:     "organization",
								Relation: "",
							},
						},
					},
					"maintainer": {
						Name: "maintainer",
						RelationReferences: []*base.RelationReference{
							{
								Type:     "user",
								Relation: "",
							},
							{
								Type:     "organization",
								Relation: "member",
							},
						},
					},
					"owner": {
						Name: "owner",
						RelationReferences: []*base.RelationReference{
							{
								Type:     "user",
								Relation: "",
							},
						},
					},
				},
				Permissions: map[string]*base.PermissionDefinition{
					"update": {
						Name: "update",
						Child: &base.Child{
							Type: &base.Child_Rewrite{
								Rewrite: &base.Rewrite{
									RewriteOperation: base.Rewrite_OPERATION_INTERSECTION,
									Children: []*base.Child{
										{
											Type: &base.Child_Leaf{
												Leaf: &base.Leaf{
													Type: &base.Leaf_ComputedUserSet{
														ComputedUserSet: &base.ComputedUserSet{
															Relation: "owner",
														},
													},
												},
											},
										},
										{
											Type: &base.Child_Rewrite{
												Rewrite: &base.Rewrite{
													RewriteOperation: base.Rewrite_OPERATION_UNION,
													Children: []*base.Child{
														{
															Type: &base.Child_Leaf{
																Leaf: &base.Leaf{
																	Type: &base.Leaf_ComputedUserSet{
																		ComputedUserSet: &base.ComputedUserSet{
																			Relation: "maintainer",
																		},
																	},
																},
															},
														},
														{
															Type: &base.Child_Leaf{
																Leaf: &base.Leaf{
																	Type: &base.Leaf_TupleToUserSet{
																		TupleToUserSet: &base.TupleToUserSet{
																			TupleSet: &base.TupleSet{
																				Relation: "parent",
																			},
																			Computed: &base.ComputedUserSet{
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
								},
							},
						},
					},
					"delete": {
						Name: "delete",
						Child: &base.Child{
							Type: &base.Child_Leaf{
								Leaf: &base.Leaf{
									Type: &base.Leaf_TupleToUserSet{
										TupleToUserSet: &base.TupleToUserSet{
											TupleSet: &base.TupleSet{
												Relation: "parent",
											},
											Computed: &base.ComputedUserSet{
												Relation: "admin",
											},
										},
									},
								},
							},
						},
					},
				},
				References: map[string]base.EntityDefinition_Reference{
					"parent":     base.EntityDefinition_REFERENCE_RELATION,
					"maintainer": base.EntityDefinition_REFERENCE_RELATION,
					"owner":      base.EntityDefinition_REFERENCE_RELATION,
					"update":     base.EntityDefinition_REFERENCE_PERMISSION,
					"delete":     base.EntityDefinition_REFERENCE_PERMISSION,
				},
			})

			Expect(NewSchemaFromEntityAndRuleDefinitions(entities, []*base.RuleDefinition{})).To(Equal(&base.SchemaDefinition{
				EntityDefinitions: map[string]*base.EntityDefinition{
					"user":         entities[0],
					"organization": entities[1],
					"repository":   entities[2],
				},
				RuleDefinitions: map[string]*base.RuleDefinition{},
				References: map[string]base.SchemaDefinition_Reference{
					"user":         base.SchemaDefinition_REFERENCE_ENTITY,
					"organization": base.SchemaDefinition_REFERENCE_ENTITY,
					"repository":   base.SchemaDefinition_REFERENCE_ENTITY,
				},
			}))
		})
	})
})
