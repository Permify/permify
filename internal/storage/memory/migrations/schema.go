package migrations

import (
	"github.com/hashicorp/go-memdb" // In-memory database
	// Internal imports
	"github.com/Permify/permify/internal/storage/memory/constants"
)

// Schema - Database schema for memory db
var Schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		constants.SchemaDefinitionsTable: { // Schema definitions table
			Name: constants.SchemaDefinitionsTable, // Table name
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "Name"},
							&memdb.StringFieldIndex{Field: "Version"},
						},
					},
				},
				"version": {
					Name:   "version",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "Version"},
						},
					},
				},
				"tenant": {
					Name:   "tenant",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
						},
					},
				},
			},
		},
		constants.AttributesTable: { // Attributes table
			Name: constants.AttributesTable, // Table name
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
							&memdb.StringFieldIndex{Field: "EntityID"},
							&memdb.StringFieldIndex{Field: "Attribute"},
						},
					},
				},
				"entity-type-index": {
					Name:   "entity-type-index",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
						},
					},
				},
				"entity-type-and-attribute-index": {
					Name:   "entity-type-and-attribute-index",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
							&memdb.StringFieldIndex{Field: "Attribute"},
						},
					},
				},
			},
		},
		constants.RelationTuplesTable: { // Relation tuples table
			Name: constants.RelationTuplesTable, // Table name
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
							&memdb.StringFieldIndex{Field: "EntityID"},
							&memdb.StringFieldIndex{Field: "Relation"},
							&memdb.StringFieldIndex{Field: "SubjectType"},
							&memdb.StringFieldIndex{Field: "SubjectID"},
							&memdb.StringFieldIndex{Field: "SubjectRelation"},
						},
						AllowMissing: true,
					},
				},
				"entity-index": {
					Name:   "entity-index",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
							&memdb.StringFieldIndex{Field: "EntityID"},
							&memdb.StringFieldIndex{Field: "Relation"},
						},
					},
				},
				"relation-index": {
					Name:   "relation-index",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
							&memdb.StringFieldIndex{Field: "Relation"},
							&memdb.StringFieldIndex{Field: "SubjectType"},
						},
					},
				},
				"entity-type-index": {
					Name:   "entity-type-index",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
						},
					},
				},
				"entity-type-and-relation-index": {
					Name:   "entity-type-and-relation-index",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "EntityType"},
							&memdb.StringFieldIndex{Field: "Relation"},
						},
					},
				},
			},
		},
		constants.TenantsTable: { // Tenants table
			Name: constants.TenantsTable, // Table name
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "ID"},
						},
					},
				},
			},
		},
		constants.BundlesTable: { // Bundles table
			Name: constants.BundlesTable, // Table name
			Indexes: map[string]*memdb.IndexSchema{ // Index schemas
				"id": { // ID index
					Name:   "id", // Index name
					Unique: true, // Unique constraint
					Indexer: &memdb.CompoundIndex{ // Compound index
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "TenantID"},
							&memdb.StringFieldIndex{Field: "Name"},
						},
					}, // End of compound index
				}, // End of ID index
			}, // End of indexes
		}, // End of bundles table
	}, // End of tables
} // End of schema
