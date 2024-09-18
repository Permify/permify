package engines

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/invoke"
	"github.com/Permify/permify/pkg/attribute"
	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
	"github.com/Permify/permify/pkg/tuple"
)

var _ = Describe("expand-engine", func() {
	// DRIVE SAMPLE
	driveSchema := `
	entity user {}

	entity organization {
		relation admin @user
	}

	entity folder {
		relation org @organization
		relation creator @user
		relation collaborator @user

		permission read = collaborator
		permission update = collaborator
		permission delete = creator or org.admin
	}

	entity doc {
		relation org @organization
		relation parent @folder
		relation owner @user @folder#creator

		permission read = (owner or parent.collaborator) or org.admin
		permission update = owner not org.admin
		permission delete = owner not update
		permission view = owner not read
		permission admin = view
	}
	`

	Context("Drive Sample: Expand", func() {
		It("Drive Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			// SCHEMA

			conf, err := newSchema(driveSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			// RELATIONSHIPS

			type expand struct {
				entity     string
				assertions map[string]*base.Expand
			}

			tests := struct {
				relationships []string
				contextual    []string
				expands       []expand
			}{
				relationships: []string{
					"doc:1#owner@user:2",
					"doc:1#parent@folder:1#...",
					"folder:1#collaborator@user:1",
					"folder:1#collaborator@user:3",
				},
				contextual: []string{
					"doc:1#org@organization:1#...",
					"organization:1#admin@user:1",
					"folder:2#creator@user:89",
					"doc:1#owner@folder:2#creator",
				},
				expands: []expand{
					{
						entity: "doc:1",
						assertions: map[string]*base.Expand{
							"read": {
								Entity: &base.Entity{
									Type: "doc",
									Id:   "1",
								},
								Permission: "read",
								Node: &base.Expand_Expand{
									Expand: &base.ExpandTreeNode{
										Operation: base.ExpandTreeNode_OPERATION_UNION,
										Children: []*base.Expand{
											{
												Entity: &base.Entity{
													Type: "doc",
													Id:   "1",
												},
												Permission: "read",
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Entity: &base.Entity{
																	Type: "doc",
																	Id:   "1",
																},
																Permission: "owner",
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_UNION,
																		Children: []*base.Expand{
																			{
																				Entity: &base.Entity{
																					Type: "folder",
																					Id:   "2",
																				},
																				Permission: "creator",
																				Node: &base.Expand_Leaf{
																					Leaf: &base.ExpandLeaf{
																						Type: &base.ExpandLeaf_Subjects{
																							Subjects: &base.Subjects{
																								Subjects: []*base.Subject{
																									{
																										Type: "user",
																										Id:   "89",
																									},
																								},
																							},
																						},
																					},
																				},
																			},
																			{
																				Entity: &base.Entity{
																					Type: "doc",
																					Id:   "1",
																				},
																				Permission: "owner",
																				Node: &base.Expand_Leaf{
																					Leaf: &base.ExpandLeaf{
																						Type: &base.ExpandLeaf_Subjects{
																							Subjects: &base.Subjects{
																								Subjects: []*base.Subject{
																									{
																										Type: "user",
																										Id:   "2",
																									},
																									{
																										Type:     "folder",
																										Id:       "2",
																										Relation: "creator",
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
															{
																Entity: &base.Entity{
																	Type: "doc",
																	Id:   "1",
																},
																Permission: "parent",
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_UNION,
																		Children: []*base.Expand{
																			{
																				Entity: &base.Entity{
																					Type: "folder",
																					Id:   "1",
																				},
																				Permission: "collaborator",
																				Node: &base.Expand_Leaf{
																					Leaf: &base.ExpandLeaf{
																						Type: &base.ExpandLeaf_Subjects{
																							Subjects: &base.Subjects{
																								Subjects: []*base.Subject{
																									{
																										Type: "user",
																										Id:   "1",
																									},
																									{
																										Type: "user",
																										Id:   "3",
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
											{
												Entity: &base.Entity{
													Type: "doc",
													Id:   "1",
												},
												Permission: "org",
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Entity: &base.Entity{
																	Type: "organization",
																	Id:   "1",
																},
																Permission: "admin",
																Node: &base.Expand_Leaf{
																	Leaf: &base.ExpandLeaf{
																		Type: &base.ExpandLeaf_Subjects{
																			Subjects: &base.Subjects{
																				Subjects: []*base.Subject{
																					{
																						Type: "user",
																						Id:   "1",
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
					},
					{
						entity: "doc:1",
						assertions: map[string]*base.Expand{
							"delete": {
								Entity: &base.Entity{
									Type: "doc",
									Id:   "1",
								},
								Permission: "delete",
								Node: &base.Expand_Expand{
									Expand: &base.ExpandTreeNode{
										Operation: base.ExpandTreeNode_OPERATION_EXCLUSION,
										Children: []*base.Expand{
											{
												Entity: &base.Entity{
													Type: "doc",
													Id:   "1",
												},
												Permission: "owner",
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Entity: &base.Entity{
																	Type: "folder",
																	Id:   "2",
																},
																Permission: "creator",
																Node: &base.Expand_Leaf{
																	Leaf: &base.ExpandLeaf{
																		Type: &base.ExpandLeaf_Subjects{
																			Subjects: &base.Subjects{
																				Subjects: []*base.Subject{
																					{
																						Type: "user",
																						Id:   "89",
																					},
																				},
																			},
																		},
																	},
																},
															},
															{
																Entity: &base.Entity{
																	Type: "doc",
																	Id:   "1",
																},
																Permission: "owner",
																Node: &base.Expand_Leaf{
																	Leaf: &base.ExpandLeaf{
																		Type: &base.ExpandLeaf_Subjects{
																			Subjects: &base.Subjects{
																				Subjects: []*base.Subject{
																					{
																						Type: "user",
																						Id:   "2",
																					},
																					{
																						Type:     "folder",
																						Id:       "2",
																						Relation: "creator",
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
											{
												Entity: &base.Entity{
													Type: "doc",
													Id:   "1",
												},
												Permission: "update",
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_EXCLUSION,
														Children: []*base.Expand{
															{
																Entity: &base.Entity{
																	Type: "doc",
																	Id:   "1",
																},
																Permission: "owner",
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_UNION,
																		Children: []*base.Expand{
																			{
																				Entity: &base.Entity{
																					Type: "folder",
																					Id:   "2",
																				},
																				Permission: "creator",
																				Node: &base.Expand_Leaf{
																					Leaf: &base.ExpandLeaf{
																						Type: &base.ExpandLeaf_Subjects{
																							Subjects: &base.Subjects{
																								Subjects: []*base.Subject{
																									{
																										Type: "user",
																										Id:   "89",
																									},
																								},
																							},
																						},
																					},
																				},
																			},
																			{
																				Entity: &base.Entity{
																					Type: "doc",
																					Id:   "1",
																				},
																				Permission: "owner",
																				Node: &base.Expand_Leaf{
																					Leaf: &base.ExpandLeaf{
																						Type: &base.ExpandLeaf_Subjects{
																							Subjects: &base.Subjects{
																								Subjects: []*base.Subject{
																									{
																										Type: "user",
																										Id:   "2",
																									},
																									{
																										Type:     "folder",
																										Id:       "2",
																										Relation: "creator",
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
															{
																Entity: &base.Entity{
																	Type: "doc",
																	Id:   "1",
																},
																Permission: "org",
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_UNION,
																		Children: []*base.Expand{
																			{
																				Entity: &base.Entity{
																					Type: "organization",
																					Id:   "1",
																				},
																				Permission: "admin",
																				Node: &base.Expand_Leaf{
																					Leaf: &base.ExpandLeaf{
																						Type: &base.ExpandLeaf_Subjects{
																							Subjects: &base.Subjects{
																								Subjects: []*base.Subject{
																									{
																										Type: "user",
																										Id:   "1",
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
									},
								},
							},
						},
					},
					{
						entity: "doc:1",
						assertions: map[string]*base.Expand{
							"view": {
								Entity: &base.Entity{
									Type: "doc",
									Id:   "1",
								},
								Permission: "view",
								Node: &base.Expand_Expand{
									Expand: &base.ExpandTreeNode{
										Operation: base.ExpandTreeNode_OPERATION_EXCLUSION,
										Children: []*base.Expand{
											{
												Entity: &base.Entity{
													Type: "doc",
													Id:   "1",
												},
												Permission: "owner",
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Entity: &base.Entity{
																	Type: "folder",
																	Id:   "2",
																},
																Permission: "creator",
																Node: &base.Expand_Leaf{
																	Leaf: &base.ExpandLeaf{
																		Type: &base.ExpandLeaf_Subjects{
																			Subjects: &base.Subjects{
																				Subjects: []*base.Subject{
																					{
																						Type: "user",
																						Id:   "89",
																					},
																				},
																			},
																		},
																	},
																},
															},
															{
																Entity: &base.Entity{
																	Type: "doc",
																	Id:   "1",
																},
																Permission: "owner",
																Node: &base.Expand_Leaf{
																	Leaf: &base.ExpandLeaf{
																		Type: &base.ExpandLeaf_Subjects{
																			Subjects: &base.Subjects{
																				Subjects: []*base.Subject{
																					{
																						Type: "user",
																						Id:   "2",
																					},
																					{
																						Type:     "folder",
																						Id:       "2",
																						Relation: "creator",
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
											{
												Entity: &base.Entity{
													Type: "doc",
													Id:   "1",
												},
												Permission: "read",
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Entity: &base.Entity{
																	Type: "doc",
																	Id:   "1",
																},
																Permission: "read",
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_UNION,
																		Children: []*base.Expand{
																			{
																				Entity: &base.Entity{
																					Type: "doc",
																					Id:   "1",
																				},
																				Permission: "owner",
																				Node: &base.Expand_Expand{
																					Expand: &base.ExpandTreeNode{
																						Operation: base.ExpandTreeNode_OPERATION_UNION,
																						Children: []*base.Expand{
																							{
																								Entity: &base.Entity{
																									Type: "folder",
																									Id:   "2",
																								},
																								Permission: "creator",
																								Node: &base.Expand_Leaf{
																									Leaf: &base.ExpandLeaf{
																										Type: &base.ExpandLeaf_Subjects{
																											Subjects: &base.Subjects{
																												Subjects: []*base.Subject{
																													{
																														Type: "user",
																														Id:   "89",
																													},
																												},
																											},
																										},
																									},
																								},
																							},
																							{
																								Entity: &base.Entity{
																									Type: "doc",
																									Id:   "1",
																								},
																								Permission: "owner",
																								Node: &base.Expand_Leaf{
																									Leaf: &base.ExpandLeaf{
																										Type: &base.ExpandLeaf_Subjects{
																											Subjects: &base.Subjects{
																												Subjects: []*base.Subject{
																													{
																														Type: "user",
																														Id:   "2",
																													},
																													{
																														Type:     "folder",
																														Id:       "2",
																														Relation: "creator",
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
																			{
																				Entity: &base.Entity{
																					Type: "doc",
																					Id:   "1",
																				},
																				Permission: "parent",
																				Node: &base.Expand_Expand{
																					Expand: &base.ExpandTreeNode{
																						Operation: base.ExpandTreeNode_OPERATION_UNION,
																						Children: []*base.Expand{
																							{
																								Entity: &base.Entity{
																									Type: "folder",
																									Id:   "1",
																								},
																								Permission: "collaborator",
																								Node: &base.Expand_Leaf{
																									Leaf: &base.ExpandLeaf{
																										Type: &base.ExpandLeaf_Subjects{
																											Subjects: &base.Subjects{
																												Subjects: []*base.Subject{
																													{
																														Type: "user",
																														Id:   "1",
																													},
																													{
																														Type: "user",
																														Id:   "3",
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
															{
																Entity: &base.Entity{
																	Type: "doc",
																	Id:   "1",
																},
																Permission: "org",
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_UNION,
																		Children: []*base.Expand{
																			{
																				Entity: &base.Entity{
																					Type: "organization",
																					Id:   "1",
																				},
																				Permission: "admin",
																				Node: &base.Expand_Leaf{
																					Leaf: &base.ExpandLeaf{
																						Type: &base.ExpandLeaf_Subjects{
																							Subjects: &base.Subjects{
																								Subjects: []*base.Subject{
																									{
																										Type: "user",
																										Id:   "1",
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
									},
								},
							},
						},
					},
					{
						entity: "doc:1",
						assertions: map[string]*base.Expand{
							"admin": {
								Entity: &base.Entity{
									Type: "doc",
									Id:   "1",
								},
								Permission: "view",
								Node: &base.Expand_Expand{
									Expand: &base.ExpandTreeNode{
										Operation: base.ExpandTreeNode_OPERATION_EXCLUSION,
										Children: []*base.Expand{
											{
												Entity: &base.Entity{
													Type: "doc",
													Id:   "1",
												},
												Permission: "owner",
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Entity: &base.Entity{
																	Type: "folder",
																	Id:   "2",
																},
																Permission: "creator",
																Node: &base.Expand_Leaf{
																	Leaf: &base.ExpandLeaf{
																		Type: &base.ExpandLeaf_Subjects{
																			Subjects: &base.Subjects{
																				Subjects: []*base.Subject{
																					{
																						Type: "user",
																						Id:   "89",
																					},
																				},
																			},
																		},
																	},
																},
															},
															{
																Entity: &base.Entity{
																	Type: "doc",
																	Id:   "1",
																},
																Permission: "owner",
																Node: &base.Expand_Leaf{
																	Leaf: &base.ExpandLeaf{
																		Type: &base.ExpandLeaf_Subjects{
																			Subjects: &base.Subjects{
																				Subjects: []*base.Subject{
																					{
																						Type: "user",
																						Id:   "2",
																					},
																					{
																						Type:     "folder",
																						Id:       "2",
																						Relation: "creator",
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
											{
												Entity: &base.Entity{
													Type: "doc",
													Id:   "1",
												},
												Permission: "read",
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Entity: &base.Entity{
																	Type: "doc",
																	Id:   "1",
																},
																Permission: "read",
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_UNION,
																		Children: []*base.Expand{
																			{
																				Entity: &base.Entity{
																					Type: "doc",
																					Id:   "1",
																				},
																				Permission: "owner",
																				Node: &base.Expand_Expand{
																					Expand: &base.ExpandTreeNode{
																						Operation: base.ExpandTreeNode_OPERATION_UNION,
																						Children: []*base.Expand{
																							{
																								Entity: &base.Entity{
																									Type: "folder",
																									Id:   "2",
																								},
																								Permission: "creator",
																								Node: &base.Expand_Leaf{
																									Leaf: &base.ExpandLeaf{
																										Type: &base.ExpandLeaf_Subjects{
																											Subjects: &base.Subjects{
																												Subjects: []*base.Subject{
																													{
																														Type: "user",
																														Id:   "89",
																													},
																												},
																											},
																										},
																									},
																								},
																							},
																							{
																								Entity: &base.Entity{
																									Type: "doc",
																									Id:   "1",
																								},
																								Permission: "owner",
																								Node: &base.Expand_Leaf{
																									Leaf: &base.ExpandLeaf{
																										Type: &base.ExpandLeaf_Subjects{
																											Subjects: &base.Subjects{
																												Subjects: []*base.Subject{
																													{
																														Type: "user",
																														Id:   "2",
																													},
																													{
																														Type:     "folder",
																														Id:       "2",
																														Relation: "creator",
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
																			{
																				Entity: &base.Entity{
																					Type: "doc",
																					Id:   "1",
																				},
																				Permission: "parent",
																				Node: &base.Expand_Expand{
																					Expand: &base.ExpandTreeNode{
																						Operation: base.ExpandTreeNode_OPERATION_UNION,
																						Children: []*base.Expand{
																							{
																								Entity: &base.Entity{
																									Type: "folder",
																									Id:   "1",
																								},
																								Permission: "collaborator",
																								Node: &base.Expand_Leaf{
																									Leaf: &base.ExpandLeaf{
																										Type: &base.ExpandLeaf_Subjects{
																											Subjects: &base.Subjects{
																												Subjects: []*base.Subject{
																													{
																														Type: "user",
																														Id:   "1",
																													},
																													{
																														Type: "user",
																														Id:   "3",
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
															{
																Entity: &base.Entity{
																	Type: "doc",
																	Id:   "1",
																},
																Permission: "org",
																Node: &base.Expand_Expand{
																	Expand: &base.ExpandTreeNode{
																		Operation: base.ExpandTreeNode_OPERATION_UNION,
																		Children: []*base.Expand{
																			{
																				Entity: &base.Entity{
																					Type: "organization",
																					Id:   "1",
																				},
																				Permission: "admin",
																				Node: &base.Expand_Leaf{
																					Leaf: &base.ExpandLeaf{
																						Type: &base.ExpandLeaf_Subjects{
																							Subjects: &base.Subjects{
																								Subjects: []*base.Subject{
																									{
																										Type: "user",
																										Id:   "1",
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
									},
								},
							},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			expandEngine := NewExpandEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				nil,
				expandEngine,
				nil,
				nil,
			)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			reqCont := &base.Context{
				Tuples: []*base.Tuple{},
			}

			for _, contextual := range tests.contextual {
				t, err := tuple.Tuple(contextual)
				Expect(err).ShouldNot(HaveOccurred())
				reqCont.Tuples = append(reqCont.Tuples, t)
			}

			for _, expand := range tests.expands {
				entity, err := tuple.E(expand.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range expand.assertions {
					var response *base.PermissionExpandResponse
					response, err = invoker.Expand(context.Background(), &base.PermissionExpandRequest{
						TenantId:   "t1",
						Entity:     entity,
						Permission: permission,
						Metadata: &base.PermissionExpandRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
						},
						Context: reqCont,
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.Tree).Should(Equal(res))
				}
			}
		})
	})

	// POLYMORPHIC RELATIONS SAMPLE

	polymorphicRelationsSchema := `
	entity googleuser {}
	
	entity facebookuser {}
	
	entity company {
	relation member @googleuser @facebookuser
	}
	
	entity organization {
		relation member @googleuser @facebookuser
	
		action edit = member
	}
	
	entity repo {
		relation parent @company @organization
	
		permission push   = parent.member
		permission delete = push
	}
	`

	Context("Polymorphic Relations Sample: Expand", func() {
		It("Polymorphic Relations Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			// SCHEMA

			conf, err := newSchema(polymorphicRelationsSchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			// RELATIONSHIPS

			type expand struct {
				entity     string
				assertions map[string]*base.Expand
			}

			tests := struct {
				relationships []string
				contextual    []string
				expands       []expand
			}{
				relationships: []string{
					"repo:1#parent@organization:1",
					"repo:1#parent@company:1",
					"company:1#member@googleuser:2",
					"organization:1#member@facebookuser:3",
					"organization:1#member@facebookuser:4",
					"organization:1#member@facebookuser:5",
				},
				expands: []expand{
					{
						entity: "repo:1",
						assertions: map[string]*base.Expand{
							"push": {
								Entity: &base.Entity{
									Type: "repo",
									Id:   "1",
								},
								Permission: "parent",
								Node: &base.Expand_Expand{
									Expand: &base.ExpandTreeNode{
										Operation: base.ExpandTreeNode_OPERATION_UNION,
										Children: []*base.Expand{
											{
												Entity: &base.Entity{
													Type: "company",
													Id:   "1",
												},
												Permission: "member",
												Node: &base.Expand_Leaf{
													Leaf: &base.ExpandLeaf{
														Type: &base.ExpandLeaf_Subjects{
															Subjects: &base.Subjects{
																Subjects: []*base.Subject{
																	{
																		Type: "googleuser",
																		Id:   "2",
																	},
																},
															},
														},
													},
												},
											},
											{
												Entity: &base.Entity{
													Type: "organization",
													Id:   "1",
												},
												Permission: "member",
												Node: &base.Expand_Leaf{
													Leaf: &base.ExpandLeaf{
														Type: &base.ExpandLeaf_Subjects{
															Subjects: &base.Subjects{
																Subjects: []*base.Subject{
																	{
																		Type: "facebookuser",
																		Id:   "3",
																	},
																	{
																		Type: "facebookuser",
																		Id:   "4",
																	},
																	{
																		Type: "facebookuser",
																		Id:   "5",
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
								Entity: &base.Entity{
									Type: "repo",
									Id:   "1",
								},
								Permission: "parent",
								Node: &base.Expand_Expand{
									Expand: &base.ExpandTreeNode{
										Operation: base.ExpandTreeNode_OPERATION_UNION,
										Children: []*base.Expand{
											{
												Entity: &base.Entity{
													Type: "company",
													Id:   "1",
												},
												Permission: "member",
												Node: &base.Expand_Leaf{
													Leaf: &base.ExpandLeaf{
														Type: &base.ExpandLeaf_Subjects{
															Subjects: &base.Subjects{
																Subjects: []*base.Subject{
																	{
																		Type: "googleuser",
																		Id:   "2",
																	},
																},
															},
														},
													},
												},
											},
											{
												Entity: &base.Entity{
													Type: "organization",
													Id:   "1",
												},
												Permission: "member",
												Node: &base.Expand_Leaf{
													Leaf: &base.ExpandLeaf{
														Type: &base.ExpandLeaf_Subjects{
															Subjects: &base.Subjects{
																Subjects: []*base.Subject{
																	{
																		Type: "facebookuser",
																		Id:   "3",
																	},
																	{
																		Type: "facebookuser",
																		Id:   "4",
																	},
																	{
																		Type: "facebookuser",
																		Id:   "5",
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
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			expandEngine := NewExpandEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				nil,
				expandEngine,
				nil,
				nil,
			)

			var tuples []*base.Tuple

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection())
			Expect(err).ShouldNot(HaveOccurred())

			var reqContext *base.Context

			for _, expand := range tests.expands {
				entity, err := tuple.E(expand.entity)
				Expect(err).ShouldNot(HaveOccurred())

				for permission, res := range expand.assertions {
					var response *base.PermissionExpandResponse
					response, err = invoker.Expand(context.Background(), &base.PermissionExpandRequest{
						TenantId:   "t1",
						Entity:     entity,
						Permission: permission,
						Metadata: &base.PermissionExpandRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
						},
						Context: reqContext,
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.Tree).Should(Equal(res))
				}
			}
		})
	})

	// WORKDAY SAMPLE

	workdaySchema := `
		entity user {}
	
		entity organization {
	
			relation member @user
	
			attribute balance integer
	
			permission view = check_balance(balance) and member
		}
	
		entity repository {
	
			relation organization  @organization
	
			attribute is_public boolean
	
			permission view = is_public or organization.member
			permission edit = organization.view
			permission delete = is_workday(is_public)
		}
	
		rule check_balance(balance integer) {
			balance > 5000
		}
	
		rule is_workday(is_public boolean) {
			 is_public == true && (context.data.day_of_week != 'saturday' && context.data.day_of_week != 'sunday')
		}
		`

	Context("Weekday Sample: Expand", func() {
		It("Weekday Sample: Case 1", func() {
			db, err := factories.DatabaseFactory(
				config.Database{
					Engine: "memory",
				},
			)

			Expect(err).ShouldNot(HaveOccurred())

			// SCHEMA

			conf, err := newSchema(workdaySchema)
			Expect(err).ShouldNot(HaveOccurred())

			schemaWriter := factories.SchemaWriterFactory(db)
			err = schemaWriter.WriteSchema(context.Background(), conf)
			Expect(err).ShouldNot(HaveOccurred())

			// RELATIONSHIPS

			type expand struct {
				entity     string
				context    map[string]interface{}
				assertions map[string]*base.Expand
			}

			anyVal, _ := anypb.New(&base.BooleanValue{Data: true})

			tests := struct {
				relationships []string
				attributes    []string
				expands       []expand
			}{
				relationships: []string{
					"repository:1#organization@organization:1",
					"organization:1#member@user:1",
				},
				attributes: []string{
					"repository:1$is_public|boolean:true",
				},
				expands: []expand{
					{
						entity: "repository:1",
						context: map[string]interface{}{
							"day_of_week": "monday",
						},
						assertions: map[string]*base.Expand{
							"view": {
								Entity: &base.Entity{
									Type: "repository",
									Id:   "1",
								},
								Permission: "view",
								Node: &base.Expand_Expand{
									Expand: &base.ExpandTreeNode{
										Operation: base.ExpandTreeNode_OPERATION_UNION,
										Children: []*base.Expand{
											{
												Entity: &base.Entity{
													Type: "repository",
													Id:   "1",
												},
												Permission: "is_public",
												Node: &base.Expand_Leaf{
													Leaf: &base.ExpandLeaf{
														Type: &base.ExpandLeaf_Value{
															Value: anyVal,
														},
													},
												},
											},
											{
												Entity: &base.Entity{
													Type: "repository",
													Id:   "1",
												},
												Permission: "organization",
												Node: &base.Expand_Expand{
													Expand: &base.ExpandTreeNode{
														Operation: base.ExpandTreeNode_OPERATION_UNION,
														Children: []*base.Expand{
															{
																Entity: &base.Entity{
																	Type: "organization",
																	Id:   "1",
																},
																Permission: "member",
																Node: &base.Expand_Leaf{
																	Leaf: &base.ExpandLeaf{
																		Type: &base.ExpandLeaf_Subjects{
																			Subjects: &base.Subjects{
																				Subjects: []*base.Subject{
																					{
																						Type: "user",
																						Id:   "1",
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
								Entity: &base.Entity{
									Type: "repository",
									Id:   "1",
								},
								Permission: "is_workday",
								Arguments: []*base.Argument{
									{
										Type: &base.Argument_ComputedAttribute{
											ComputedAttribute: &base.ComputedAttribute{
												Name: "is_public",
											},
										},
									},
								},
								Node: &base.Expand_Leaf{
									Leaf: &base.ExpandLeaf{
										Type: &base.ExpandLeaf_Values{
											Values: &base.Values{
												Values: map[string]*anypb.Any{
													"is_public": anyVal,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			schemaReader := factories.SchemaReaderFactory(db)
			dataReader := factories.DataReaderFactory(db)
			dataWriter := factories.DataWriterFactory(db)

			expandEngine := NewExpandEngine(schemaReader, dataReader)

			invoker := invoke.NewDirectInvoker(
				schemaReader,
				dataReader,
				nil,
				expandEngine,
				nil,
				nil,
			)

			var tuples []*base.Tuple
			var attributes []*base.Attribute

			for _, relationship := range tests.relationships {
				t, err := tuple.Tuple(relationship)
				Expect(err).ShouldNot(HaveOccurred())
				tuples = append(tuples, t)
			}

			for _, attr := range tests.attributes {
				a, err := attribute.Attribute(attr)
				Expect(err).ShouldNot(HaveOccurred())
				attributes = append(attributes, a)
			}

			_, err = dataWriter.Write(context.Background(), "t1", database.NewTupleCollection(tuples...), database.NewAttributeCollection(attributes...))
			Expect(err).ShouldNot(HaveOccurred())

			for _, expand := range tests.expands {
				entity, err := tuple.E(expand.entity)
				Expect(err).ShouldNot(HaveOccurred())

				ctx := &base.Context{
					Tuples:     []*base.Tuple{},
					Attributes: []*base.Attribute{},
					Data:       &structpb.Struct{},
				}

				if expand.context != nil {
					value, err := structpb.NewStruct(expand.context)
					if err != nil {
						fmt.Printf("Error creating struct: %v", err)
					}
					ctx.Data = value
				}

				for permission, res := range expand.assertions {

					var response *base.PermissionExpandResponse
					response, err = invoker.Expand(context.Background(), &base.PermissionExpandRequest{
						TenantId:   "t1",
						Entity:     entity,
						Permission: permission,
						Metadata: &base.PermissionExpandRequestMetadata{
							SnapToken:     token.NewNoopToken().Encode().String(),
							SchemaVersion: "",
						},
						Context: ctx,
					})

					Expect(err).ShouldNot(HaveOccurred())
					Expect(response.Tree).Should(Equal(res))
				}
			}
		})
	})
})
