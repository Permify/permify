package migrations

import (
	"github.com/Permify/permify/internal/storage/memory"
	"github.com/hashicorp/go-memdb"
)

// Schema - Database schema for memory db
var Schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		memory.SchemaDefinitionsTable: {
			Name: memory.SchemaDefinitionsTable,
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
		memory.AttributesTable: {
			Name: memory.AttributesTable,
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
		memory.RelationTuplesTable: {
			Name: memory.RelationTuplesTable,
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
		memory.TenantsTable: {
			Name: memory.TenantsTable,
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
		memory.BundlesTable: {
			Name: memory.BundlesTable,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.StringFieldIndex{
						Field: "Name",
					},
				},
			},
		},
	},
}
