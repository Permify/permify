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

func DataReaderFactory(db database.Database, logger logger.Interface) (repo storage.DataReader) {
	switch db.GetEngineType() {
	// case "postgres":
	//	return PQRepository.NewRelationshipReader(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewDataReader(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewDataReader(db.(*MMDatabase.Memory), logger)
	}
}

func DataWriterFactory(db database.Database, logger logger.Interface) (repo storage.DataWriter) {
	switch db.GetEngineType() {
	// case "postgres":
	//	return PQRepository.NewRelationshipWriter(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewDataWriter(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewDataWriter(db.(*MMDatabase.Memory), logger)
	}
}

func SchemaReaderFactory(db database.Database, logger logger.Interface) (repo storage.SchemaReader) {
	switch db.GetEngineType() {
	// case "postgres":
	//	return PQRepository.NewSchemaReader(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewSchemaReader(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewSchemaReader(db.(*MMDatabase.Memory), logger)
	}
}

func WatcherFactory(db database.Database, logger logger.Interface) (repo storage.Watcher) {
	switch db.GetEngineType() {
	case "postgres":
		return PQRepository.NewWatcher(db.(*PQDatabase.Postgres), logger)
	case "memory":
		return MMRepository.NewWatcher(db.(*MMDatabase.Memory), logger)
	default:
		return MMRepository.NewWatcher(db.(*MMDatabase.Memory), logger)
	}
}

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
