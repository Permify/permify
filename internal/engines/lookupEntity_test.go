package engines

//var _ = Describe("lookup-entity-command", func() {
//	var checkCommand *CheckCommand
//
//	// DRIVE SAMPLE
//
//	driveSchema := `
//entity user {}
//
//entity organization {
//	relation admin @user
//}
//
//entity folder {
//	relation org @organization
//	relation creator @user
//	relation collaborator @user
//
//	action read = collaborator
//	action update = collaborator
//	action delete = creator or org.admin
//}
//
//entity doc {
//	relation org @organization
//	relation parent @folder
//	relation owner @user
//
//	action read = (owner or parent.collaborator) or org.admin
//	action update = owner and org.admin
//	action delete = owner or org.admin
//	action share = update and (owner or parent.update)
//}
//`
//
//	Context("Drive Sample: Check", func() {
//		It("Drive Sample: Case 1", func() {
//			var err error
//
//			// SCHEMA
//
//			schemaReader := new(mocks.SchemaReader)
//
//			var sch *base.SchemaDefinition
//			sch, err = schema.NewSchemaFromStringDefinitions(true, driveSchema)
//			Expect(err).ShouldNot(HaveOccurred())
//
//			var doc *base.EntityDefinition
//			doc, err = schema.GetEntityByName(sch, "doc")
//			Expect(err).ShouldNot(HaveOccurred())
//
//			var folder *base.EntityDefinition
//			folder, err = schema.GetEntityByName(sch, "folder")
//			Expect(err).ShouldNot(HaveOccurred())
//
//			var organization *base.EntityDefinition
//			organization, err = schema.GetEntityByName(sch, "organization")
//			Expect(err).ShouldNot(HaveOccurred())
//
//			schemaReader.On("ReadSchemaDefinition", "t1", "doc", "noop").Return(doc, "noop", nil).Times(4)
//			schemaReader.On("ReadSchemaDefinition", "t1", "folder", "noop").Return(folder, "noop", nil).Times(1)
//			schemaReader.On("ReadSchemaDefinition", "t1", "organization", "noop").Return(organization, "noop", nil).Times(1)
//
//			// RELATIONSHIPS
//
//			relationshipReader := new(mocks.RelationshipReader)
//			relationshipReaderForLookupCommand := new(mocks.RelationshipReader)
//
//			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
//				Entity: &base.EntityFilter{
//					Type: "doc",
//					Ids:  []string{"1"},
//				},
//				Relation: "owner",
//			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
//				{
//					Entity: &base.Entity{
//						Type: "doc",
//						Id:   "1",
//					},
//					Relation: "owner",
//					Subject: &base.Subject{
//						Type:     tuple.USER,
//						Id:       "2",
//						Relation: "",
//					},
//				},
//			}...), nil).Times(1)
//
//			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
//				Entity: &base.EntityFilter{
//					Type: "doc",
//					Ids:  []string{"2"},
//				},
//				Relation: "owner",
//			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
//				{
//					Entity: &base.Entity{
//						Type: "doc",
//						Id:   "2",
//					},
//					Relation: "owner",
//					Subject: &base.Subject{
//						Type:     tuple.USER,
//						Id:       "8",
//						Relation: "",
//					},
//				},
//			}...), nil).Times(1)
//
//			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
//				Entity: &base.EntityFilter{
//					Type: "doc",
//					Ids:  []string{"1"},
//				},
//				Relation: "parent",
//			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
//				{
//					Entity: &base.Entity{
//						Type: "doc",
//						Id:   "1",
//					},
//					Relation: "parent",
//					Subject: &base.Subject{
//						Type:     "folder",
//						Id:       "1",
//						Relation: tuple.ELLIPSIS,
//					},
//				},
//			}...), nil).Times(1)
//
//			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
//				Entity: &base.EntityFilter{
//					Type: "doc",
//					Ids:  []string{"2"},
//				},
//				Relation: "parent",
//			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{}...), nil).Times(1)
//
//			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
//				Entity: &base.EntityFilter{
//					Type: "folder",
//					Ids:  []string{"1"},
//				},
//				Relation: "collaborator",
//			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
//				{
//					Entity: &base.Entity{
//						Type: "folder",
//						Id:   "1",
//					},
//					Relation: "collaborator",
//					Subject: &base.Subject{
//						Type:     tuple.USER,
//						Id:       "1",
//						Relation: "",
//					},
//				},
//				{
//					Entity: &base.Entity{
//						Type: "folder",
//						Id:   "1",
//					},
//					Relation: "collaborator",
//					Subject: &base.Subject{
//						Type:     tuple.USER,
//						Id:       "3",
//						Relation: "",
//					},
//				},
//			}...), nil).Times(1)
//
//			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
//				Entity: &base.EntityFilter{
//					Type: "doc",
//					Ids:  []string{"1"},
//				},
//				Relation: "org",
//			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
//				{
//					Entity: &base.Entity{
//						Type: "doc",
//						Id:   "1",
//					},
//					Relation: "org",
//					Subject: &base.Subject{
//						Type:     "organization",
//						Id:       "1",
//						Relation: tuple.ELLIPSIS,
//					},
//				},
//			}...), nil).Times(1)
//
//			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
//				Entity: &base.EntityFilter{
//					Type: "doc",
//					Ids:  []string{"2"},
//				},
//				Relation: "org",
//			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{}...), nil).Times(1)
//
//			relationshipReader.On("QueryRelationships", "t1", &base.TupleFilter{
//				Entity: &base.EntityFilter{
//					Type: "organization",
//					Ids:  []string{"1"},
//				},
//				Relation: "admin",
//			}, token.NewNoopToken().Encode().String()).Return(database.NewTupleIterator([]*base.Tuple{
//				{
//					Entity: &base.Entity{
//						Type: "organization",
//						Id:   "1",
//					},
//					Relation: "admin",
//					Subject: &base.Subject{
//						Type:     tuple.USER,
//						Id:       "1",
//						Relation: "",
//					},
//				},
//			}...), nil).Times(1)
//
//			relationshipReaderForLookupCommand.On("GetUniqueEntityIDsByEntityType", "t1", "doc", token.NewNoopToken().Encode().String()).Return([]string{"1", "2"}, nil).Times(1)
//
//			checkCommand, _ = NewCheckCommand(keys.NewNoopCheckCommandKeys(), schemaReader, relationshipReader, telemetry.NewNoopMeter())
//			lookupEntityCommand := NewLookupEntityCommand(checkCommand, schemaReader, relationshipReaderForLookupCommand)
//
//			req := &base.PermissionLookupEntityRequest{
//				TenantId:   "t1",
//				EntityType: "doc",
//				Subject:    &base.Subject{Type: tuple.USER, Id: "1"},
//				Permission: "read",
//				Metadata: &base.PermissionLookupEntityRequestMetadata{
//					SnapToken:     token.NewNoopToken().Encode().String(),
//					SchemaVersion: "noop",
//					Depth:         20,
//				},
//			}
//
//			var response *base.PermissionLookupEntityResponse
//			response, err = lookupEntityCommand.Execute(context.Background(), req)
//			Expect(err).ShouldNot(HaveOccurred())
//			Expect(response.EntityIds).Should(Equal([]string{"1"}))
//		})
//	})
//})
