package factories

import (
	"github.com/Permify/permify/internal/storage"
	MMRepository "github.com/Permify/permify/internal/storage/memory"
	PQRepository "github.com/Permify/permify/internal/storage/postgres"
	"github.com/Permify/permify/pkg/database"
	MMDatabase "github.com/Permify/permify/pkg/database/memory"
	PQDatabase "github.com/Permify/permify/pkg/database/postgres"
	"github.com/Permify/permify/pkg/logger"
)

// RelationshipReaderFactory is a factory function that returns a relationship reader instance according to the
// given database interface. It supports different types of databases, such as PostgreSQL and in-memory databases.
//
// db: the database.Database instance for which the relationship reader should be created
// logger: the logger.Interface instance to be used by the relationship reader for logging purposes
//
// Returns a storage.RelationshipReader instance that performs read operations on the relationships stored
// in the given database. If the database engine type is not recognized, it defaults to an in-memory database.
func RelationshipReaderFactory(db database.Database, logger logger.Interface) (repo storage.RelationshipReader) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewRelationshipReader(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewRelationshipReader(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewRelationshipReader(db.(*MMDatabase.Memory), logger)
	}
}

// RelationshipWriterFactory is a factory function that returns a relationship writer instance according to the
// given database interface. It supports different types of databases, such as PostgreSQL and in-memory databases.
//
// db: the database.Database instance for which the relationship writer should be created
// logger: the logger.Interface instance to be used by the relationship writer for logging purposes
//
// Returns a storage.RelationshipWriter instance that performs write operations on the relationships stored
// in the given database. If the database engine type is not recognized, it defaults to an in-memory database.
func RelationshipWriterFactory(db database.Database, logger logger.Interface) (repo storage.RelationshipWriter) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewRelationshipWriter(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewRelationshipWriter(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewRelationshipWriter(db.(*MMDatabase.Memory), logger)
	}
}

// SchemaReaderFactory is a factory function that returns a schema reader instance according to the
// given database interface. It supports different types of databases, such as PostgreSQL and in-memory databases.
//
// db: the database.Database instance for which the schema reader should be created
// logger: the logger.Interface instance to be used by the schema reader for logging purposes
//
// Returns a storage.SchemaReader instance that performs read operations on the schema stored
// in the given database. If the database engine type is not recognized, it defaults to an in-memory database.
func SchemaReaderFactory(db database.Database, logger logger.Interface) (repo storage.SchemaReader) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewSchemaReader(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewSchemaReader(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewSchemaReader(db.(*MMDatabase.Memory), logger)
	}
}

// SchemaWriterFactory is a factory function that returns a schema writer instance according to the
// given database interface. It supports different types of databases, such as PostgreSQL and in-memory databases.
//
// db: the database.Database instance for which the schema writer should be created
// logger: the logger.Interface instance to be used by the schema writer for logging purposes
//
// Returns a storage.SchemaWriter instance that performs write operations on the schema stored
// in the given database. If the database engine type is not recognized, it defaults to an in-memory database.
func SchemaWriterFactory(db database.Database, logger logger.Interface) (repo storage.SchemaWriter) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewSchemaWriter(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewSchemaWriter(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewSchemaWriter(db.(*MMDatabase.Memory), logger)
	}
}

// TenantReaderFactory is a factory function that returns a tenant reader instance according to the
// given database interface. It supports different types of databases, such as PostgreSQL and in-memory databases.
//
// db: the database.Database instance for which the tenant reader should be created
// logger: the logger.Interface instance to be used by the tenant reader for logging purposes
//
// Returns a storage.TenantReader instance that performs read operations on the tenants stored
// in the given database. If the database engine type is not recognized, it defaults to an in-memory database.
func TenantReaderFactory(db database.Database, logger logger.Interface) (repo storage.TenantReader) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewTenantReader(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewTenantReader(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewTenantReader(db.(*MMDatabase.Memory), logger)
	}
}

// TenantWriterFactory is a factory function that returns a tenant writer instance according to the
// given database interface. It supports different types of databases, such as PostgreSQL and in-memory databases.
//
// db: the database.Database instance for which the tenant writer should be created
// logger: the logger.Interface instance to be used by the tenant writer for logging purposes
//
// Returns a storage.TenantWriter instance that performs write operations on the tenants stored
// in the given database. If the database engine type is not recognized, it defaults to an in-memory database.
func TenantWriterFactory(db database.Database, logger logger.Interface) (repo storage.TenantWriter) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewTenantWriter(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewTenantWriter(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewTenantWriter(db.(*MMDatabase.Memory), logger)
	}
}
