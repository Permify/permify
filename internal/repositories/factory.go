package repositories

import (
	MNRepository "github.com/Permify/permify/internal/repositories/mongo"
	PQRepository "github.com/Permify/permify/internal/repositories/postgres"
	"github.com/Permify/permify/pkg/database"
	MNDatabase "github.com/Permify/permify/pkg/database/mongo"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
)

// RelationTupleFactory -
func RelationTupleFactory(db database.Database) (repo IRelationTupleRepository) {
	switch db.GetConnectionType() {
	case "postgres":
		return PQRepository.NewRelationTupleRepository(db.(*PQDatabase.Postgres))
	case "mongo":
		return MNRepository.NewRelationTupleRepository(db.(*MNDatabase.Mongo))
	default:
		return PQRepository.NewRelationTupleRepository(db.(*PQDatabase.Postgres))
	}
}

// EntityConfigFactory -
func EntityConfigFactory(db database.Database) (repo IEntityConfigRepository) {
	switch db.GetConnectionType() {
	case "postgres":
		return PQRepository.NewEntityConfigRepository(db.(*PQDatabase.Postgres))
	case "mongo":
		return MNRepository.NewEntityConfigRepository(db.(*MNDatabase.Mongo))
	default:
		return PQRepository.NewEntityConfigRepository(db.(*PQDatabase.Postgres))
	}
}
