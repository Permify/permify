package factories

import (
	"github.com/Permify/permify/internal/repositories"
	MMRepository "github.com/Permify/permify/internal/repositories/memory"
	PQRepository "github.com/Permify/permify/internal/repositories/postgres"
	"github.com/Permify/permify/pkg/database"
	MMDatabase "github.com/Permify/permify/pkg/database/memory"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
)

// RelationshipReaderFactory - Return relationship read operations according to given database interface
func RelationshipReaderFactory(db database.Database) (repo repositories.RelationshipReader) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewRelationshipReader(db.(*PQDatabase.Postgres))
	case "memory":
		return MMRepository.NewRelationshipReader(db.(*MMDatabase.Memory))
	default:
		return MMRepository.NewRelationshipReader(db.(*MMDatabase.Memory))
	}
}

// RelationshipWriterFactory - Return relationship write operations according to given database interface
func RelationshipWriterFactory(db database.Database) (repo repositories.RelationshipWriter) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewRelationshipWriter(db.(*PQDatabase.Postgres))
	case "memory":
		return MMRepository.NewRelationshipWriter(db.(*MMDatabase.Memory))
	default:
		return MMRepository.NewRelationshipWriter(db.(*MMDatabase.Memory))
	}
}

// SchemaReaderFactory - Return schema read operations according to given database interface
func SchemaReaderFactory(db database.Database) (repo repositories.SchemaReader) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewSchemaReader(db.(*PQDatabase.Postgres))
	case "memory":
		return MMRepository.NewSchemaReader(db.(*MMDatabase.Memory))
	default:
		return PQRepository.NewSchemaReader(db.(*PQDatabase.Postgres))
	}
}

// SchemaWriterFactory - Return schema write operations according to given database interface
func SchemaWriterFactory(db database.Database) (repo repositories.SchemaWriter) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewSchemaWriter(db.(*PQDatabase.Postgres))
	case "memory":
		return MMRepository.NewSchemaWriter(db.(*MMDatabase.Memory))
	default:
		return MMRepository.NewSchemaWriter(db.(*MMDatabase.Memory))
	}
}
