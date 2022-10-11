package memory

import (
	"github.com/hashicorp/go-memdb"
)

var Schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		"entity_config": {
			Name: "entity_config",
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
		"relation_tuple": {
			Name: "relation_tuple",
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "Entity"},
							&memdb.StringFieldIndex{Field: "EntityID"},
							&memdb.StringFieldIndex{Field: "Relation"},
							&memdb.StringFieldIndex{Field: "SubjectEntity"},
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
							&memdb.StringFieldIndex{Field: "Entity"},
							&memdb.StringFieldIndex{Field: "EntityID"},
							&memdb.StringFieldIndex{Field: "Relation"},
						},
					},
				},
				"subject-index": {
					Name:   "subject-index",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "Entity"},
							&memdb.StringFieldIndex{Field: "Relation"},
							&memdb.StringFieldIndex{Field: "SubjectEntity"},
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
