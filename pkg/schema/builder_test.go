package schema

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TestBuilder -
func TestBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "builder-suite")
}

var _ = Describe("compiler", func() {
	Context("Schema", func() {
		It("Case 1", func() {
			is := Schema(
				Entity("user", Relations(), Permissions()),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
				"user": {
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_RelationalReference{},
				},
			}))
		})

		It("Case 2", func() {
			is := Schema(
				Entity("user", Relations(), Permissions()),
				Entity("organization",
					Relations(
						Relation("owner", Reference("user")),
						Relation("admin", Reference("user")),
					),
					Permissions(
						Permission("update",
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
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_RelationalReference{},
				},
				"organization": {
					Name: "organization",
					Permissions: map[string]*base.PermissionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Exclusion: false,
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
												Exclusion: false,
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
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
			}))
		})

		It("Case 3", func() {
			is := Schema(
				Entity("user", Relations(), Permissions()),
				Entity("organization",
					Relations(
						Relation("owner", Reference("user")),
						Relation("admin", Reference("user")),
					),
					Permissions(
						Permission("update",
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
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_RelationalReference{},
				},
				"organization": {
					Name: "organization",
					Permissions: map[string]*base.PermissionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Exclusion: false,
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
														RewriteOperation: base.Rewrite_OPERATION_INTERSECTION,
														Children: []*base.Child{
															{
																Exclusion: false,
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
																Exclusion: false,
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
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
			}))
		})

		It("Case 4", func() {
			is := Schema(
				Entity("user", Relations(), Permissions()),
				Entity("organization",
					Relations(
						Relation("owner", Reference("user")),
						Relation("admin", Reference("user")),
					),
					Permissions(
						Permission("update",
							ComputedUserSet("owner", false),
						),
					),
				),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
				"user": {
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_RelationalReference{},
				},
				"organization": {
					Name: "organization",
					Permissions: map[string]*base.PermissionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Exclusion: false,
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
						},
					},
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
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
			}))
		})

		It("Case 5", func() {
			is := Schema(
				Entity("user", Relations(), Permissions()),
				Entity("organization",
					Relations(
						Relation("owner", Reference("user")),
						Relation("admin", Reference("user")),
					),
					Permissions(
						Permission("update",
							ComputedUserSet("owner", false),
						),
					),
				),
				Entity("repository",
					Relations(
						Relation("parent", Reference("organization")),
						Relation("owner", Reference("user"), Reference("organization#admin")),
					),
					Permissions(
						Permission("delete",
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
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_RelationalReference{},
				},
				"organization": {
					Name: "organization",
					Permissions: map[string]*base.PermissionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Exclusion: false,
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
						},
					},
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
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
				"repository": {
					Name: "repository",
					Permissions: map[string]*base.PermissionDefinition{
						"delete": {
							Name: "delete",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Exclusion: false,
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
																Exclusion: false,
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
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
																Exclusion: true,
																Type: &base.Child_Leaf{
																	Leaf: &base.Leaf{
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
									Type:     "organization",
									Relation: "",
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
								{
									Type:     "organization",
									Relation: "admin",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"parent": base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"delete": base.EntityDefinition_RELATIONAL_REFERENCE_PERMISSION,
					},
				},
			}))
		})
	})
})
