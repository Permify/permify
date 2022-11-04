package memory

import (
	"github.com/hashicorp/go-memdb"
)

var Schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		schemaDefinitionTable: {
			Name: schemaDefinitionTable,
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
		relationTuplesTable: {
			Name: relationTuplesTable,
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
			},
		},
	},
}
