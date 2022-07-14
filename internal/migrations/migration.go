package migrations

import (
	"fmt"

	"github.com/Permify/permify/migrations/postgres/notifier"
	"github.com/Permify/permify/migrations/postgres/subscriber"
	"github.com/Permify/permify/pkg/migration"
)

// RegisterSubscriberMigrations -
func RegisterSubscriberMigrations() (mi *migration.Migration, err error) {
	mi = migration.New()
	err = mi.Register(migration.TABLE, "initial_tuple", subscriber.CreateRelationTupleMigration())
	err = mi.Register(migration.TABLE, "initial_config", subscriber.CreateEntityConfigMigration())
	return
}

// RegisterNotifierMigrations -
func RegisterNotifierMigrations(tables []string) (mi *migration.Migration, err error) {
	mi = migration.New()
	err = mi.Register(migration.FUNCTION, "initial_function", notifier.CreateNotifyFunctionMigration())
	for i, table := range tables {
		err = mi.Register(migration.TRIGGER, fmt.Sprintf("initial_%v", i), notifier.CreateNotifyTriggerMigration(table))
	}
	return
}
