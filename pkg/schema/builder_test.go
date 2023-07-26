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
				Entity("user", Relations(), Attributes(), Permissions())...,
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
				"user": {
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_Reference{},
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
								ComputedUserSet("owner"),
								ComputedUserSet("admin"),
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
					References:  map[string]base.EntityDefinition_Reference{},
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
					References: map[string]base.EntityDefinition_Reference{
						"owner":  base.EntityDefinition_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_REFERENCE_RELATION,
						"update": base.EntityDefinition_REFERENCE_PERMISSION,
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
								ComputedUserSet("owner"),
								Intersection(
									ComputedUserSet("admin"),
									ComputedUserSet("owner"),
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
					References:  map[string]base.EntityDefinition_Reference{},
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
					References: map[string]base.EntityDefinition_Reference{
						"owner":  base.EntityDefinition_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_REFERENCE_RELATION,
						"update": base.EntityDefinition_REFERENCE_PERMISSION,
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
							ComputedUserSet("owner"),
						),
					),
				),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
				"user": {
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_Reference{},
				},
				"organization": {
					Name: "organization",
					Permissions: map[string]*base.PermissionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
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
					References: map[string]base.EntityDefinition_Reference{
						"owner":  base.EntityDefinition_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_REFERENCE_RELATION,
						"update": base.EntityDefinition_REFERENCE_PERMISSION,
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
							ComputedUserSet("owner"),
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
								ComputedUserSet("owner"),
								Exclusion(
									TupleToUserSet("parent", "update"),
									TupleToUserSet("parent", "owner"),
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
					References:  map[string]base.EntityDefinition_Reference{},
				},
				"organization": {
					Name: "organization",
					Permissions: map[string]*base.PermissionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
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
					References: map[string]base.EntityDefinition_Reference{
						"owner":  base.EntityDefinition_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_REFERENCE_RELATION,
						"update": base.EntityDefinition_REFERENCE_PERMISSION,
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
														RewriteOperation: base.Rewrite_OPERATION_EXCLUSION,
														Children: []*base.Child{
															{
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
					References: map[string]base.EntityDefinition_Reference{
						"parent": base.EntityDefinition_REFERENCE_RELATION,
						"owner":  base.EntityDefinition_REFERENCE_RELATION,
						"delete": base.EntityDefinition_REFERENCE_PERMISSION,
					},
				},
			}))
		})
	})
})
