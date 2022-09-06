package memory

import (
	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/repositories/entities"
)

var Schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		entities.EntityConfig{}.Table(): {
			Name: entities.EntityConfig{}.Table(),
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "Entity"},
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
		entities.RelationTuple{}.Table(): {
			Name: entities.RelationTuple{}.Table(),
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "Entity"},
							&memdb.StringFieldIndex{Field: "ObjectID"},
							&memdb.StringFieldIndex{Field: "Relation"},
							&memdb.StringFieldIndex{Field: "UsersetEntity"},
							&memdb.StringFieldIndex{Field: "UsersetObjectID"},
							&memdb.StringFieldIndex{Field: "UsersetRelation"},
						},
						AllowMissing: true,
					},
				},
				"entity-index": {
					Name:   "entity-index",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "Entity"},
							&memdb.StringFieldIndex{Field: "ObjectID"},
							&memdb.StringFieldIndex{Field: "Relation"},
						},
					},
				},
				"entity": {
					Name:   "entity",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "Entity"},
						},
					},
				},
			},
		},
	},
}
