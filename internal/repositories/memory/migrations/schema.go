package migrations

import (
	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/repositories/memory"
)

// Schema - Database schema for memory db
var Schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		memory.SchemaDefinitionTable: {
			Name: memory.SchemaDefinitionTable,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "EntityType"},
							&memdb.StringFieldIndex{Field: "Version"},
						},
					},
				},
				"version": {
					Name:    "version",
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "Version"},
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
							&memdb.StringFieldIndex{Field: "EntityType"},
						},
					},
				},
				"entity-type-and-relation-index": {
					Name:   "entity-type-and-relation-index",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "EntityType"},
							&memdb.StringFieldIndex{Field: "Relation"},
						},
					},
				},
			},
		},
	},
}
