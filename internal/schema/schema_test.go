package schema

import (
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"

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
	Context("NewSchemaFromEntityAndRuleDefinitions", func() {
		It("Case 1", func() {
			entities := make([]*base.EntityDefinition, 0, 1)

			entities = append(entities, &base.EntityDefinition{
				Name:        "user",
				Relations:   map[string]*base.RelationDefinition{},
				Permissions: map[string]*base.PermissionDefinition{},
				References:  map[string]base.EntityDefinition_Reference{},
			})

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

	Context("NewSchemaFromStringDefinitions", func() {
		It("Case 1", func() {
			Expect(NewSchemaFromStringDefinitions(true, `
entity user {}

entity organization {
	relation admin @user
}
`)).To(Equal(&base.SchemaDefinition{
				EntityDefinitions: map[string]*base.EntityDefinition{
					"user": {
						Name:        "user",
						Relations:   map[string]*base.RelationDefinition{},
						Attributes:  map[string]*base.AttributeDefinition{},
						Permissions: map[string]*base.PermissionDefinition{},
						References:  map[string]base.EntityDefinition_Reference{},
					},
					"organization": {
						Name: "organization",
						Relations: map[string]*base.RelationDefinition{
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
						Attributes:  map[string]*base.AttributeDefinition{},
						Permissions: map[string]*base.PermissionDefinition{},
						References: map[string]base.EntityDefinition_Reference{
							"admin": base.EntityDefinition_REFERENCE_RELATION,
						},
					},
				},
				RuleDefinitions: map[string]*base.RuleDefinition{},
				References: map[string]base.SchemaDefinition_Reference{
					"user":         base.SchemaDefinition_REFERENCE_ENTITY,
					"organization": base.SchemaDefinition_REFERENCE_ENTITY,
				},
			}))
		})
	})

	Context("NewEntityAndRuleDefinitionsFromStringDefinitions", func() {
		It("Case 1", func() {
			enDef, ruDef, err := NewEntityAndRuleDefinitionsFromStringDefinitions(true, `
entity user {}

entity organization {
	relation admin @user
}
`)

			Expect(err).ShouldNot(HaveOccurred())

			Expect(enDef).To(Equal([]*base.EntityDefinition{
				{
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Attributes:  map[string]*base.AttributeDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_Reference{},
				},
				{
					Name: "organization",
					Relations: map[string]*base.RelationDefinition{
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
					Attributes:  map[string]*base.AttributeDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References: map[string]base.EntityDefinition_Reference{
						"admin": base.EntityDefinition_REFERENCE_RELATION,
					},
				},
			}))

			Expect(ruDef).Should(Equal([]*base.RuleDefinition{}))

			_, _, err = NewEntityAndRuleDefinitionsFromStringDefinitions(true, `
entity user {

entity organization {
	relation admin @user
}
`)

			Expect(err.Error()).Should(Equal("4:21:expected token to be RELATION, PERMISSION, ATTRIBUTE, got ENTITY instead"))
		})
	})

	Context("GetEntityByName", func() {
		It("Case 1", func() {
			sch, err := NewSchemaFromStringDefinitions(true, `
entity user {}

entity organization {
	relation admin @user
}
`)
			Expect(err).ShouldNot(HaveOccurred())

			enDef, err := GetEntityByName(sch, "organization")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(enDef).To(Equal(&base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
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
				Attributes:  map[string]*base.AttributeDefinition{},
				Permissions: map[string]*base.PermissionDefinition{},
				References: map[string]base.EntityDefinition_Reference{
					"admin": base.EntityDefinition_REFERENCE_RELATION,
				},
			}))

			_, err = GetEntityByName(sch, "team")
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_ENTITY_DEFINITION_NOT_FOUND.String()))
		})
	})

	Context("GetRuleByName", func() {
		It("Case 1", func() {
			sch, err := NewSchemaFromStringDefinitions(true, `
entity user {}

entity organization {
	relation admin @user
}

rule check_region(region string, regions string[]) {
	region in regions
}
`)
			Expect(err).ShouldNot(HaveOccurred())

			ruDef, err := GetRuleByName(sch, "check_region")
			Expect(err).ShouldNot(HaveOccurred())

			envOptions := []cel.EnvOption{
				cel.Variable("region", types.StringType),
				cel.Variable("regions", cel.ListType(cel.StringType)),
			}

			env, err := cel.NewEnv(envOptions...)
			Expect(err).ShouldNot(HaveOccurred())

			compiledExp, issues := env.Compile("region in regions")
			Expect(issues.Err()).ShouldNot(HaveOccurred())

			exp, err := cel.AstToCheckedExpr(compiledExp)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(ruDef.Name).Should(Equal("check_region"))
			Expect(ruDef.Arguments).Should(Equal(map[string]base.AttributeType{
				"region":  base.AttributeType_ATTRIBUTE_TYPE_STRING,
				"regions": base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY,
			}))

			Expect(ruDef.Expression.Expr.String()).Should(Equal(exp.Expr.String()))

			_, err = GetRuleByName(sch, "check_category")
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_RULE_DEFINITION_NOT_FOUND.String()))
		})
	})

	Context("GetTypeOfReferenceByNameInEntityDefinition", func() {
		It("Case 1", func() {
			rdt, err := GetTypeOfReferenceByNameInEntityDefinition(&base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
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
				Attributes:  map[string]*base.AttributeDefinition{},
				Permissions: map[string]*base.PermissionDefinition{},
				References: map[string]base.EntityDefinition_Reference{
					"admin": base.EntityDefinition_REFERENCE_RELATION,
				},
			}, "admin")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rdt).Should(Equal(base.EntityDefinition_REFERENCE_RELATION))

			_, err = GetTypeOfReferenceByNameInEntityDefinition(&base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
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
				Attributes:  map[string]*base.AttributeDefinition{},
				Permissions: map[string]*base.PermissionDefinition{},
				References: map[string]base.EntityDefinition_Reference{
					"admin": base.EntityDefinition_REFERENCE_RELATION,
				},
			}, "member")
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String()))
		})
	})

	Context("GetPermissionByNameInEntityDefinition", func() {
		It("Case 1", func() {
			perDef, err := GetPermissionByNameInEntityDefinition(&base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
					"admin": {
						Name: "admin",
						RelationReferences: []*base.RelationReference{
							{
								Type:     "user",
								Relation: "",
							},
						},
					},
					"member": {
						Name: "member",
						RelationReferences: []*base.RelationReference{
							{
								Type:     "user",
								Relation: "",
							},
						},
					},
				},
				Attributes: map[string]*base.AttributeDefinition{},
				Permissions: map[string]*base.PermissionDefinition{
					"view": {
						Name: "view",
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
															Relation: "admin",
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
															Relation: "member",
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
					"admin":  base.EntityDefinition_REFERENCE_RELATION,
					"member": base.EntityDefinition_REFERENCE_RELATION,
					"view":   base.EntityDefinition_REFERENCE_PERMISSION,
				},
			}, "view")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(perDef).Should(Equal(&base.PermissionDefinition{
				Name: "view",
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
													Relation: "admin",
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
													Relation: "member",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}))

			_, err = GetPermissionByNameInEntityDefinition(&base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
					"admin": {
						Name: "admin",
						RelationReferences: []*base.RelationReference{
							{
								Type:     "user",
								Relation: "",
							},
						},
					},
					"member": {
						Name: "member",
						RelationReferences: []*base.RelationReference{
							{
								Type:     "user",
								Relation: "",
							},
						},
					},
				},
				Attributes: map[string]*base.AttributeDefinition{},
				Permissions: map[string]*base.PermissionDefinition{
					"view": {
						Name: "view",
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
															Relation: "admin",
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
															Relation: "member",
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
					"admin":  base.EntityDefinition_REFERENCE_RELATION,
					"member": base.EntityDefinition_REFERENCE_RELATION,
					"view":   base.EntityDefinition_REFERENCE_PERMISSION,
				},
			}, "edit")
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_PERMISSION_DEFINITION_NOT_FOUND.String()))
		})
	})

	Context("GetRelationByNameInEntityDefinition", func() {
		It("Case 1", func() {
			relDef, err := GetRelationByNameInEntityDefinition(&base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
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
				Attributes:  map[string]*base.AttributeDefinition{},
				Permissions: map[string]*base.PermissionDefinition{},
				References: map[string]base.EntityDefinition_Reference{
					"admin": base.EntityDefinition_REFERENCE_RELATION,
				},
			}, "admin")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(relDef).Should(Equal(&base.RelationDefinition{
				Name: "admin",
				RelationReferences: []*base.RelationReference{
					{
						Type:     "user",
						Relation: "",
					},
				},
			}))

			_, err = GetRelationByNameInEntityDefinition(&base.EntityDefinition{
				Name: "organization",
				Relations: map[string]*base.RelationDefinition{
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
				Attributes:  map[string]*base.AttributeDefinition{},
				Permissions: map[string]*base.PermissionDefinition{},
				References: map[string]base.EntityDefinition_Reference{
					"admin": base.EntityDefinition_REFERENCE_RELATION,
				},
			}, "member")
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_RELATION_DEFINITION_NOT_FOUND.String()))
		})
	})

	Context("GetAttributeByNameInEntityDefinition", func() {
		It("Case 1", func() {
			attrDef, err := GetAttributeByNameInEntityDefinition(&base.EntityDefinition{
				Name:      "organization",
				Relations: map[string]*base.RelationDefinition{},
				Attributes: map[string]*base.AttributeDefinition{
					"public": {
						Name: "public",
						Type: base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN,
					},
				},
				Permissions: map[string]*base.PermissionDefinition{},
				References: map[string]base.EntityDefinition_Reference{
					"public": base.EntityDefinition_REFERENCE_ATTRIBUTE,
				},
			}, "public")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(attrDef).Should(Equal(&base.AttributeDefinition{
				Name: "public",
				Type: base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN,
			}))

			_, err = GetAttributeByNameInEntityDefinition(&base.EntityDefinition{
				Name:      "organization",
				Relations: map[string]*base.RelationDefinition{},
				Attributes: map[string]*base.AttributeDefinition{
					"public": {
						Name: "public",
						Type: base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN,
					},
				},
				Permissions: map[string]*base.PermissionDefinition{},
				References: map[string]base.EntityDefinition_Reference{
					"public": base.EntityDefinition_REFERENCE_ATTRIBUTE,
				},
			}, "private")
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_ATTRIBUTE_DEFINITION_NOT_FOUND.String()))
		})
	})
})
