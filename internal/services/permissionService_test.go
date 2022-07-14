package services

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("permission-service", func() {
	// var permissionService *PermissionService

	// sampleDriveSchema := schema.Schema{}

	Context("Check", func() {
		//It("Drive Sample: User Is Folder Direct Collaborator", func() {
		//	relationTupleRepository := new(mocks.RelationTupleRepository)
		//
		//	tuples1 := []entities.RelationTuple{{
		//		Entity:          "doc",
		//		ObjectID:        "1",
		//		Relation:        "parent",
		//		UsersetEntity:   "folder",
		//		UsersetObjectID: "1",
		//		UsersetRelation: tuple.ELLIPSIS,
		//	}}
		//
		//	tuples2 := []entities.RelationTuple{{
		//		Entity:          "doc",
		//		ObjectID:        "1",
		//		Relation:        "organization",
		//		UsersetEntity:   "organization",
		//		UsersetObjectID: "1",
		//		UsersetRelation: tuple.ELLIPSIS,
		//	}}
		//
		//	tuples3 := []entities.RelationTuple{{
		//		Entity:          "doc",
		//		ObjectID:        "1",
		//		Relation:        "owner",
		//		UsersetEntity:   "",
		//		UsersetObjectID: "2",
		//		UsersetRelation: "",
		//	}}
		//
		//	tuples4 := []entities.RelationTuple{
		//		{
		//			Entity:          "organization",
		//			ObjectID:        "1",
		//			Relation:        "admin",
		//			UsersetEntity:   "",
		//			UsersetObjectID: "2",
		//			UsersetRelation: "",
		//		},
		//		{
		//			Entity:          "organization",
		//			ObjectID:        "1",
		//			Relation:        "admin",
		//			UsersetEntity:   "",
		//			UsersetObjectID: "3",
		//			UsersetRelation: "",
		//		},
		//		{
		//			Entity:          "organization",
		//			ObjectID:        "1",
		//			Relation:        "admin",
		//			UsersetEntity:   "",
		//			UsersetObjectID: "8",
		//			UsersetRelation: "",
		//		},
		//	}
		//
		//	tuples5 := []entities.RelationTuple{
		//		{
		//			Entity:          "folder",
		//			ObjectID:        "1",
		//			Relation:        "collaborators",
		//			UsersetEntity:   "",
		//			UsersetObjectID: "2",
		//			UsersetRelation: "",
		//		},
		//		{
		//			Entity:          "folder",
		//			ObjectID:        "1",
		//			Relation:        "collaborators",
		//			UsersetEntity:   "",
		//			UsersetObjectID: "1",
		//			UsersetRelation: "",
		//		},
		//	}
		//
		//	relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(tuples1, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(tuples3, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "doc", "1", "organization").Return(tuples2, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(tuples4, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "folder", "1", "collaborators").Return(tuples5, nil).Once()
		//
		//	permissionService = &PermissionService{
		//		schema:     sampleDriveSchema,
		//		repository: relationTupleRepository,
		//	}
		//
		//	actualResult, _, err := permissionService.Check(context.Background(), "1", "read", "doc:1", 5)
		//	Expect(err).ShouldNot(HaveOccurred())
		//	Expect(true).Should(Equal(actualResult))
		//})
		//
		//It("Drive Sample: User Is Folder Indirect Collaborator", func() {
		//	relationTupleRepository := new(mocks.RelationTupleRepository)
		//
		//	tuples1 := []entities.RelationTuple{{
		//		Entity:          "doc",
		//		ObjectID:        "1",
		//		Relation:        "parent",
		//		UsersetEntity:   "folder",
		//		UsersetObjectID: "1",
		//		UsersetRelation: tuple.ELLIPSIS,
		//	}}
		//
		//	tuples2 := []entities.RelationTuple{{
		//		Entity:          "doc",
		//		ObjectID:        "1",
		//		Relation:        "organization",
		//		UsersetEntity:   "",
		//		UsersetObjectID: "1",
		//		UsersetRelation: "",
		//	}}
		//
		//	tuples3 := []entities.RelationTuple{
		//		{
		//			Entity:          "doc",
		//			ObjectID:        "1",
		//			Relation:        "owner",
		//			UsersetEntity:   "",
		//			UsersetObjectID: "2",
		//			UsersetRelation: "",
		//		},
		//		{
		//			Entity:          "doc",
		//			ObjectID:        "1",
		//			Relation:        "owner",
		//			UsersetEntity:   "organization",
		//			UsersetObjectID: "1",
		//			UsersetRelation: "admin",
		//		},
		//	}
		//
		//	tuples4 := []entities.RelationTuple{
		//		{
		//			Entity:          "organization",
		//			ObjectID:        "1",
		//			Relation:        "admin",
		//			UsersetEntity:   "",
		//			UsersetObjectID: "3",
		//			UsersetRelation: "",
		//		},
		//	}
		//
		//	tuples5 := []entities.RelationTuple{
		//		{
		//			Entity:          "folder",
		//			ObjectID:        "1",
		//			Relation:        "collaborators",
		//			UsersetEntity:   "",
		//			UsersetObjectID: "2",
		//			UsersetRelation: "",
		//		},
		//		{
		//			Entity:          "folder",
		//			ObjectID:        "1",
		//			Relation:        "collaborators",
		//			UsersetEntity:   "",
		//			UsersetObjectID: "1",
		//			UsersetRelation: "",
		//		},
		//	}
		//
		//	relationTupleRepository.On("QueryTuples", "doc", "1", "parent").Return(tuples1, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "doc", "1", "organization").Return(tuples2, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(tuples3, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(tuples4, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "folder", "1", "collaborators").Return(tuples5, nil).Once()
		//
		//	permissionService = &PermissionService{
		//		schema:     sampleDriveSchema,
		//		repository: relationTupleRepository,
		//	}
		//
		//	actualResult, _, err := permissionService.Check(context.Background(), "1", "read", "doc:1", 5)
		//	Expect(err).ShouldNot(HaveOccurred())
		//	Expect(true).Should(Equal(actualResult))
		//})
		//
		//It("Drive Sample: User Is Organization Admin", func() {
		//	relationTupleRepository := new(mocks.RelationTupleRepository)
		//
		//	tuples1 := []entities.RelationTuple{{
		//		Entity:          "folder",
		//		ObjectID:        "1",
		//		Relation:        "organization",
		//		UsersetEntity:   "organization",
		//		UsersetObjectID: "1",
		//		UsersetRelation: tuple.ELLIPSIS,
		//	}}
		//	tuples2 := []entities.RelationTuple{
		//		{
		//			Entity:          "organization",
		//			ObjectID:        "1",
		//			Relation:        "admin",
		//			UsersetEntity:   "",
		//			UsersetObjectID: "1",
		//			UsersetRelation: "",
		//		},
		//	}
		//
		//	relationTupleRepository.On("QueryTuples", "folder", "1", "organization").Return(tuples1, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(tuples2, nil).Once()
		//
		//	permissionService = &PermissionService{
		//		schema:     sampleDriveSchema,
		//		repository: relationTupleRepository,
		//	}
		//
		//	actualResult, _, err := permissionService.Check(context.Background(), "1", "delete", "folder:1", 5)
		//	Expect(err).ShouldNot(HaveOccurred())
		//	Expect(true).Should(Equal(actualResult))
		//})
		//
		//It("Drive Sample: User Is Is Not Organization Admin And Doc Owner", func() {
		//	relationTupleRepository := new(mocks.RelationTupleRepository)
		//
		//	tuples1 := []entities.RelationTuple{{
		//		Entity:          "folder",
		//		ObjectID:        "1",
		//		Relation:        "organization",
		//		UsersetEntity:   "organization",
		//		UsersetObjectID: "1",
		//		UsersetRelation: tuple.ELLIPSIS,
		//	}}
		//	tuples2 := []entities.RelationTuple{
		//		{
		//			Entity:          "organization",
		//			ObjectID:        "1",
		//			Relation:        "admin",
		//			UsersetEntity:   "",
		//			UsersetObjectID: "2",
		//			UsersetRelation: "",
		//		},
		//		{
		//			Entity:          "organization",
		//			ObjectID:        "1",
		//			Relation:        "admin",
		//			UsersetEntity:   "doc",
		//			UsersetObjectID: "1",
		//			UsersetRelation: "owner",
		//		},
		//	}
		//	tuples3 := []entities.RelationTuple{
		//		{
		//			Entity:          "doc",
		//			ObjectID:        "1",
		//			Relation:        "owner",
		//			UsersetEntity:   "",
		//			UsersetObjectID: "1",
		//			UsersetRelation: "",
		//		},
		//	}
		//
		//	relationTupleRepository.On("QueryTuples", "folder", "1", "organization").Return(tuples1, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(tuples2, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(tuples3, nil).Once()
		//
		//	permissionService = &PermissionService{
		//		schema:     sampleDriveSchema,
		//		repository: relationTupleRepository,
		//	}
		//
		//	actualResult, _, err := permissionService.Check(context.Background(), "1", "delete", "folder:1", 5)
		//	Expect(err).ShouldNot(HaveOccurred())
		//	Expect(true).Should(Equal(actualResult))
		//})
		//
		//It("Drive Sample: Circle Depth Test", func() {
		//	relationTupleRepository := new(mocks.RelationTupleRepository)
		//
		//	tuples1 := []entities.RelationTuple{{
		//		Entity:          "folder",
		//		ObjectID:        "1",
		//		Relation:        "organization",
		//		UsersetEntity:   "organization",
		//		UsersetObjectID: "1",
		//		UsersetRelation: tuple.ELLIPSIS,
		//	}}
		//	tuples2 := []entities.RelationTuple{
		//		{
		//			Entity:          "organization",
		//			ObjectID:        "1",
		//			Relation:        "admin",
		//			UsersetEntity:   "doc",
		//			UsersetObjectID: "1",
		//			UsersetRelation: "owner",
		//		},
		//	}
		//	tuples3 := []entities.RelationTuple{
		//		{
		//			Entity:          "doc",
		//			ObjectID:        "1",
		//			Relation:        "owner",
		//			UsersetEntity:   "organization",
		//			UsersetObjectID: "1",
		//			UsersetRelation: "admin",
		//		},
		//	}
		//
		//	relationTupleRepository.On("QueryTuples", "folder", "1", "organization").Return(tuples1, nil).Once()
		//
		//	relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(tuples2, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(tuples3, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(tuples2, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(tuples3, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(tuples2, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(tuples3, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(tuples2, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(tuples3, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(tuples2, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(tuples3, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "organization", "1", "admin").Return(tuples2, nil).Once()
		//	relationTupleRepository.On("QueryTuples", "doc", "1", "owner").Return(tuples3, nil).Once()
		//
		//	permissionService = &PermissionService{
		//		schema:     sampleDriveSchema,
		//		repository: relationTupleRepository,
		//	}
		//
		//	actualResult, _, err := permissionService.Check(context.Background(), "1", "delete", "folder:1", 5)
		//	Expect(err).Should(Equal(DepthError))
		//	Expect(false).Should(Equal(actualResult))
		//})
	})
})
