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

// DataReaderFactory creates and returns a DataReader based on the database engine type.
func DataReaderFactory(db database.Database, logger logger.Interface) (repo storage.DataReader) {
	switch db.GetEngineType() {
	case "postgres":
		// If the database engine is Postgres, create a new DataReader using the Postgres implementation
		return PQRepository.NewDataReader(db.(*PQDatabase.Postgres), logger)
	case "memory":
		// If the database engine is in-memory, create a new DataReader using the in-memory implementation
		return MMRepository.NewDataReader(db.(*MMDatabase.Memory), logger)
	default:
		// For any other type, use the in-memory implementation as a default
		return MMRepository.NewDataReader(db.(*MMDatabase.Memory), logger)
	}
}

// DataWriterFactory creates and returns a DataWriter based on the database engine type.
func DataWriterFactory(db database.Database, logger logger.Interface) (repo storage.DataWriter) {
	switch db.GetEngineType() {
	case "postgres":
		// If the database engine is Postgres, create a new DataWriter using the Postgres implementation
		return PQRepository.NewDataWriter(db.(*PQDatabase.Postgres), logger)
	case "memory":
		// If the database engine is in-memory, create a new DataWriter using the in-memory implementation
		return MMRepository.NewDataWriter(db.(*MMDatabase.Memory), logger)
	default:
		// For any other type, use the in-memory implementation as a default
		return MMRepository.NewDataWriter(db.(*MMDatabase.Memory), logger)
	}
}

// SchemaReaderFactory creates and returns a SchemaReader based on the database engine type.
func SchemaReaderFactory(db database.Database, logger logger.Interface) (repo storage.SchemaReader) {
	switch db.GetEngineType() {
	case "postgres":
		// If the database engine is Postgres, create a new SchemaReader using the Postgres implementation
		return PQRepository.NewSchemaReader(db.(*PQDatabase.Postgres), logger)
	case "memory":
		// If the database engine is in-memory, create a new SchemaReader using the in-memory implementation
		return MMRepository.NewSchemaReader(db.(*MMDatabase.Memory), logger)
	default:
		// For any other type, use the in-memory implementation as a default
		return MMRepository.NewSchemaReader(db.(*MMDatabase.Memory), logger)
	}
}

// WatcherFactory creates and returns a Watcher based on the database engine type.
func WatcherFactory(db database.Database, logger logger.Interface) (repo storage.Watcher) {
	switch db.GetEngineType() {
	case "postgres":
		// If the database engine is Postgres, create a new Watcher using the Postgres implementation
		return PQRepository.NewWatcher(db.(*PQDatabase.Postgres), logger)
	case "memory":
		// If the database engine is in-memory, create a new Watcher using the in-memory implementation
		return MMRepository.NewWatcher(db.(*MMDatabase.Memory), logger)
	default:
		// For any other type, use the in-memory implementation as a default
		return MMRepository.NewWatcher(db.(*MMDatabase.Memory), logger)
	}
}

// SchemaWriterFactory creates and returns a SchemaWriter based on the database engine type.
func SchemaWriterFactory(db database.Database, logger logger.Interface) (repo storage.SchemaWriter) {
	switch db.GetEngineType() {
	case "postgres":
		// If the database engine is Postgres, create a new SchemaWriter using the Postgres implementation
		return PQRepository.NewSchemaWriter(db.(*PQDatabase.Postgres), logger)
	case "memory":
		// If the database engine is in-memory, create a new SchemaWriter using the in-memory implementation
		return MMRepository.NewSchemaWriter(db.(*MMDatabase.Memory), logger)
	default:
		// For any other type, use the in-memory implementation as a default
		return MMRepository.NewSchemaWriter(db.(*MMDatabase.Memory), logger)
	}
}

// TenantReaderFactory creates and returns a TenantReader based on the database engine type.
func TenantReaderFactory(db database.Database, logger logger.Interface) (repo storage.TenantReader) {
	switch db.GetEngineType() {
	case "postgres":
		// If the database engine is Postgres, create a new TenantReader using the Postgres implementation
		return PQRepository.NewTenantReader(db.(*PQDatabase.Postgres), logger)
	case "memory":
		// If the database engine is in-memory, create a new TenantReader using the in-memory implementation
		return MMRepository.NewTenantReader(db.(*MMDatabase.Memory), logger)
	default:
		// For any other type, use the in-memory implementation as a default
		return MMRepository.NewTenantReader(db.(*MMDatabase.Memory), logger)
	}
}

// TenantWriterFactory creates and returns a TenantWriter based on the database engine type.
func TenantWriterFactory(db database.Database, logger logger.Interface) (repo storage.TenantWriter) {
	switch db.GetEngineType() {
	case "postgres":
		// If the database engine is Postgres, create a new TenantWriter using the Postgres implementation
		return PQRepository.NewTenantWriter(db.(*PQDatabase.Postgres), logger)
	case "memory":
		// If the database engine is in-memory, create a new TenantWriter using the in-memory implementation
		return MMRepository.NewTenantWriter(db.(*MMDatabase.Memory), logger)
	default:
		// For any other type, use the in-memory implementation as a default
		return MMRepository.NewTenantWriter(db.(*MMDatabase.Memory), logger)
	}
}
