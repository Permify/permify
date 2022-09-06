package factories

import (
	"github.com/Permify/permify/internal/repositories"
	MMRepository "github.com/Permify/permify/internal/repositories/memory"
	MNRepository "github.com/Permify/permify/internal/repositories/mongo"
	PQRepository "github.com/Permify/permify/internal/repositories/postgres"
	"github.com/Permify/permify/pkg/database"
	MMDatabase "github.com/Permify/permify/pkg/database/memory"
	MNDatabase "github.com/Permify/permify/pkg/database/mongo"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
)

// RelationTupleFactory -
func RelationTupleFactory(db database.Database) (repo repositories.IRelationTupleRepository) {
	switch db.GetConnectionType() {
	case "postgres":
		return PQRepository.NewRelationTupleRepository(db.(*PQDatabase.Postgres))
	case "mongo":
		return MNRepository.NewRelationTupleRepository(db.(*MNDatabase.Mongo))
	case "memory":
		return MMRepository.NewRelationTupleRepository(db.(*MMDatabase.Memory))
	default:
		return MMRepository.NewRelationTupleRepository(db.(*MMDatabase.Memory))
	}
}

// EntityConfigFactory -
func EntityConfigFactory(db database.Database) (repo repositories.IEntityConfigRepository) {
	switch db.GetConnectionType() {
	case "postgres":
		return PQRepository.NewEntityConfigRepository(db.(*PQDatabase.Postgres))
	case "mongo":
		return MNRepository.NewEntityConfigRepository(db.(*MNDatabase.Mongo))
	case "memory":
		return MMRepository.NewEntityConfigRepository(db.(*MMDatabase.Memory))
	default:
		return PQRepository.NewEntityConfigRepository(db.(*PQDatabase.Postgres))
	}
}
