package factories

import (
	"github.com/Permify/permify/internal/repositories"
	MMRepository "github.com/Permify/permify/internal/repositories/memory"
	PQRepository "github.com/Permify/permify/internal/repositories/postgres"
	"github.com/Permify/permify/pkg/database"
	MMDatabase "github.com/Permify/permify/pkg/database/memory"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
)

// RelationshipReaderFactory - Return relationship read operations according to given database interface
func RelationshipReaderFactory(db database.Database, logger logger.Interface) (repo repositories.RelationshipReader) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewRelationshipReader(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewRelationshipReader(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewRelationshipReader(db.(*MMDatabase.Memory), logger)
	}
}

// RelationshipWriterFactory - Return relationship write operations according to given database interface
func RelationshipWriterFactory(db database.Database, logger logger.Interface) (repo repositories.RelationshipWriter) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewRelationshipWriter(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewRelationshipWriter(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewRelationshipWriter(db.(*MMDatabase.Memory), logger)
	}
}

// SchemaReaderFactory - Return schema read operations according to given database interface
func SchemaReaderFactory(db database.Database, logger logger.Interface) (repo repositories.SchemaReader) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewSchemaReader(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewSchemaReader(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewSchemaReader(db.(*MMDatabase.Memory), logger)
	}
}

// SchemaWriterFactory - Return schema write operations according to given database interface
func SchemaWriterFactory(db database.Database, logger logger.Interface) (repo repositories.SchemaWriter) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewSchemaWriter(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewSchemaWriter(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewSchemaWriter(db.(*MMDatabase.Memory), logger)
	}
}
