package schema

import (
	`testing`

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	base `github.com/Permify/permify/pkg/pb/base/v1`
)

// TestBuilder -
func TestBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "builder-suite")
}

var _ = Describe("compiler", func() {

	Context("SchemaDefinition", func() {

		It("Case 1", func() {

			is := Schema(
				Entity("user", Relations(), Actions()),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
				"user": {
					Name:       "user",
					Relations:  map[string]*base.RelationDefinition{},
					Actions:    map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{},
				},
			}))
		})

		It("Case 2", func() {
			is := Schema(
				Entity("user", Relations(), Actions()),
				Entity("organization",
					Relations(
						Relation("owner", Reference("user")),
						Relation("admin", Reference("user")),
					),
					Actions(
						Action("update",
							Union(
								ComputedUserSet("owner", false),
								ComputedUserSet("admin", false),
							),
						),
					),
				),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
				"user": {
					Name:       "user",
					Relations:  map[string]*base.RelationDefinition{},
					Actions:    map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{},
				},
				"organization": {
					Name: "organization",
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
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Name: "user",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Name: "user",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
			}))
		})

		It("Case 3", func() {
			is := Schema(
				Entity("user", Relations(), Actions()),
				Entity("organization",
					Relations(
						Relation("owner", Reference("user")),
						Relation("admin", Reference("user")),
					),
					Actions(
						Action("update",
							Union(
								ComputedUserSet("owner", false),
								Intersection(
									ComputedUserSet("admin", false),
									ComputedUserSet("owner", false),
								),
							),
						),
					),
				),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
				"user": {
					Name:       "user",
					Relations:  map[string]*base.RelationDefinition{},
					Actions:    map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{},
				},
				"organization": {
					Name: "organization",
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
																				Relation: "admin",
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
																				Relation: "owner",
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
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Name: "user",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Name: "user",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
			}))
		})

		It("Case 4", func() {
			is := Schema(
				Entity("user", Relations(), Actions()),
				Entity("organization",
					Relations(
						Relation("owner", Reference("user")),
						Relation("admin", Reference("user")),
					),
					Actions(
						Action("update",
							ComputedUserSet("owner", false),
						),
					),
				),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
				"user": {
					Name:       "user",
					Relations:  map[string]*base.RelationDefinition{},
					Actions:    map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{},
				},
				"organization": {
					Name: "organization",
					Actions: map[string]*base.ActionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
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
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Name: "user",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Name: "user",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
			}))
		})

		It("Case 5", func() {
			is := Schema(
				Entity("user", Relations(), Actions()),
				Entity("organization",
					Relations(
						Relation("owner", Reference("user")),
						Relation("admin", Reference("user")),
					),
					Actions(
						Action("update",
							ComputedUserSet("owner", false),
						),
					),
				),
				Entity("repository",
					Relations(
						Relation("parent", Reference("organization")),
						Relation("owner", Reference("user")),
					),
					Actions(
						Action("delete",
							Union(
								ComputedUserSet("owner", false),
								Union(
									TupleToUserSet("parent", "update", false),
									TupleToUserSet("parent", "owner", true),
								),
							),
						),
					),
				),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
				"user": {
					Name:       "user",
					Relations:  map[string]*base.RelationDefinition{},
					Actions:    map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{},
				},
				"organization": {
					Name: "organization",
					Actions: map[string]*base.ActionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
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
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Name: "user",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Name: "user",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
				"repository": {
					Name: "repository",
					Actions: map[string]*base.ActionDefinition{
						"delete": {
							Name: "delete",
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
												Type: &base.Child_Rewrite{
													Rewrite: &base.Rewrite{
														RewriteOperation: base.Rewrite_OPERATION_UNION,
														Children: []*base.Child{
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
																					Relation: "update",
																				},
																			},
																		},
																	},
																},
															},
															{
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
																		Exclusion: true,
																		Type: &base.Leaf_TupleToUserSet{
																			TupleToUserSet: &base.TupleToUserSet{
																				TupleSet: &base.TupleSet{
																					Relation: "parent",
																				},
																				Computed: &base.ComputedUserSet{
																					Relation: "owner",
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
					},
					Relations: map[string]*base.RelationDefinition{
						"parent": {
							Name: "parent",
							RelationReferences: []*base.RelationReference{
								{
									Name: "organization",
								},
							},
						},
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Name: "user",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"parent": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"delete": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
			}))
		})
	})
})
