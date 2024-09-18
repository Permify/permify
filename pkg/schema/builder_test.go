package schema

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

var _ = Describe("schema", func() {
	Context("Schema", func() {
		It("Case 1", func() {
			is := Schema(
				Entities(Entity("user", Relations(), Attributes(), Permissions())),
				Rules(),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
				"user": {
					Name:        "user",
					Relations:   map[string]*base.RelationDefinition{},
					Attributes:  map[string]*base.AttributeDefinition{},
					Permissions: map[string]*base.PermissionDefinition{},
					References:  map[string]base.EntityDefinition_Reference{},
				},
			}))
		})

		It("Case 2", func() {
			is := Schema(
				Entities(
					Entity("user", Relations(), Attributes(), Permissions()),
					Entity("organization",
						Relations(
							Relation("owner", Reference("user")),
							Relation("admin", Reference("user")),
						),
						Attributes(),
						Permissions(
							Permission("update",
								Union(
									ComputedUserSet("owner"),
									ComputedUserSet("admin"),
								),
							),
						),
					)),
				Rules(),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
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
					Attributes: map[string]*base.AttributeDefinition{},
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
						"admin":  base.EntityDefinition_REFERENCE_RELATION,
						"update": base.EntityDefinition_REFERENCE_PERMISSION,
					},
				},
			}))
		})

		It("Case 3", func() {
			is := Schema(
				Entities(
					Entity("user", Relations(), Attributes(), Permissions()),
					Entity("organization",
						Relations(
							Relation("owner", Reference("user")),
							Relation("admin", Reference("user")),
						),
						Attributes(),
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
					)),
				Rules(),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
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
					Attributes: map[string]*base.AttributeDefinition{},
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
				Entities(
					Entity("user", Relations(), Attributes(), Permissions()),
					Entity("organization",
						Relations(
							Relation("owner", Reference("user")),
							Relation("admin", Reference("user")),
						),
						Attributes(),
						Permissions(
							Permission("update",
								ComputedUserSet("owner"),
							),
						),
					)),
				Rules(),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
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
					Attributes: map[string]*base.AttributeDefinition{},
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
				Entities(
					Entity("user", Relations(), Attributes(), Permissions()),
					Entity("organization",
						Relations(
							Relation("owner", Reference("user")),
							Relation("admin", Reference("user")),
						),
						Attributes(),
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
						Attributes(),
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
				),
				Rules(),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
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
					Attributes: map[string]*base.AttributeDefinition{},
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
					References: map[string]base.EntityDefinition_Reference{
						"owner":  base.EntityDefinition_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_REFERENCE_RELATION,
						"update": base.EntityDefinition_REFERENCE_PERMISSION,
					},
				},
				"repository": {
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
					Attributes: map[string]*base.AttributeDefinition{},
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
					References: map[string]base.EntityDefinition_Reference{
						"parent": base.EntityDefinition_REFERENCE_RELATION,
						"owner":  base.EntityDefinition_REFERENCE_RELATION,
						"delete": base.EntityDefinition_REFERENCE_PERMISSION,
					},
				},
			}))
		})

		It("Case 6", func() {
			is := Schema(
				Entities(
					Entity("user", Relations(), Attributes(), Permissions()),
					Entity("organization",
						Relations(
							Relation("owner", Reference("user")),
							Relation("admin", Reference("user")),
						),
						Attributes(
							Attribute("is_public", base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN),
						),
						Permissions(
							Permission("update",
								Union(
									ComputedAttribute("is_public"),
									ComputedUserSet("owner"),
								),
							),
						),
					),
					Entity("repository",
						Relations(
							Relation("parent", Reference("organization")),
							Relation("owner", Reference("user"), Reference("organization#admin")),
						),
						Attributes(
							Attribute("is_public", base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN),
						),
						Permissions(
							Permission("edit",
								Call("is_workday", &base.Argument{
									Type: &base.Argument_ComputedAttribute{
										ComputedAttribute: &base.ComputedAttribute{
											Name: "is_public",
										},
									},
								}),
							),
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
				),
				Rules(
					Rule("is_workday",
						map[string]base.AttributeType{
							"is_public": base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN,
						},
						"is_public == true && (context.data.day_of_week != 'saturday' && context.data.day_of_week != 'sunday')",
					),
				),
			)

			Expect(is.EntityDefinitions).Should(Equal(map[string]*base.EntityDefinition{
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
					Attributes: map[string]*base.AttributeDefinition{
						"is_public": {
							Name: "is_public",
							Type: base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN,
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
														Type: &base.Leaf_ComputedAttribute{
															ComputedAttribute: &base.ComputedAttribute{
																Name: "is_public",
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
					References: map[string]base.EntityDefinition_Reference{
						"owner":     base.EntityDefinition_REFERENCE_RELATION,
						"admin":     base.EntityDefinition_REFERENCE_RELATION,
						"is_public": base.EntityDefinition_REFERENCE_ATTRIBUTE,
						"update":    base.EntityDefinition_REFERENCE_PERMISSION,
					},
				},
				"repository": {
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
					Attributes: map[string]*base.AttributeDefinition{
						"is_public": {
							Name: "is_public",
							Type: base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN,
						},
					},
					Permissions: map[string]*base.PermissionDefinition{
						"edit": {
							Name: "edit",
							Child: &base.Child{
								Type: &base.Child_Leaf{
									Leaf: &base.Leaf{
										Type: &base.Leaf_Call{
											Call: &base.Call{
												RuleName: "is_workday",
												Arguments: []*base.Argument{
													{
														Type: &base.Argument_ComputedAttribute{
															ComputedAttribute: &base.ComputedAttribute{
																Name: "is_public",
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
					References: map[string]base.EntityDefinition_Reference{
						"parent":    base.EntityDefinition_REFERENCE_RELATION,
						"owner":     base.EntityDefinition_REFERENCE_RELATION,
						"is_public": base.EntityDefinition_REFERENCE_ATTRIBUTE,
						"edit":      base.EntityDefinition_REFERENCE_PERMISSION,
						"delete":    base.EntityDefinition_REFERENCE_PERMISSION,
					},
				},
			}))
		})
	})
})
