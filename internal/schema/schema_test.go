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

			Expect(NewSchemaFromEntityDefinitions(entities...)).To(Equal(&base.SchemaDefinition{
				EntityDefinitions: map[string]*base.EntityDefinition{
					"user": entities[0],
				},
			}))
		})

		It("Case 2", func() {
			entities := make([]*base.EntityDefinition, 0, 2)

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
						RelationReferences: []*base.RelationReference{
							{
								Name: "user",
							},
						},
						Option: map[string]string{},
					},
					"admin": {
						Name: "admin",
						RelationReferences: []*base.RelationReference{
							{
								Name: "user",
							},
						},
						Option: map[string]string{},
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

			Expect(NewSchemaFromEntityDefinitions(entities...)).To(Equal(&base.SchemaDefinition{
				EntityDefinitions: map[string]*base.EntityDefinition{
					"user":         entities[0],
					"organization": entities[1],
				},
			}))
		})

		It("Case 3", func() {
			entities := make([]*base.EntityDefinition, 0, 3)

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
						RelationReferences: []*base.RelationReference{
							{
								Name: "user",
							},
						},
						Option: map[string]string{},
					},
					"admin": {
						Name: "admin",
						RelationReferences: []*base.RelationReference{
							{
								Name: "user",
							},
						},
						Option: map[string]string{},
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
			}, &base.EntityDefinition{
				Name: "repository",
				Relations: map[string]*base.RelationDefinition{
					"parent": {
						Name: "parent",
						RelationReferences: []*base.RelationReference{
							{
								Name: "organization",
							},
						},
						Option: map[string]string{},
					},
					"maintainer": {
						Name: "maintainer",
						RelationReferences: []*base.RelationReference{
							{
								Name: "user",
							},
							{
								Name: "organization#member",
							},
						},
						Option: map[string]string{},
					},
					"owner": {
						Name: "owner",
						RelationReferences: []*base.RelationReference{
							{
								Name: "user",
							},
						},
						Option: map[string]string{},
					},
				},
				Actions: map[string]*base.ActionDefinition{
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
																			Relation: "maintainer",
																		},
																	},
																},
															},
														},
														{
															Type: &base.Child_Leaf{
																Leaf: &base.Leaf{
																	Exclusion: false,
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
									Exclusion: false,
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
				References: map[string]base.EntityDefinition_RelationalReference{
					"parent":     base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
					"maintainer": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
					"owner":      base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
					"update":     base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					"delete":     base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
				},
				Option: map[string]string{},
			})

			Expect(NewSchemaFromEntityDefinitions(entities...)).To(Equal(&base.SchemaDefinition{
				EntityDefinitions: map[string]*base.EntityDefinition{
					"user":         entities[0],
					"organization": entities[1],
					"repository":   entities[2],
				},
			}))
		})
	})
})
