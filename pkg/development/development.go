package development

import (
	`fmt`

	`github.com/Permify/permify/internal/commands`
	`github.com/Permify/permify/internal/config`
	`github.com/Permify/permify/internal/factories`
	`github.com/Permify/permify/internal/keys`
	"github.com/Permify/permify/internal/services"
	`github.com/Permify/permify/pkg/database`
	`github.com/Permify/permify/pkg/logger`
)

// Container -
type Container struct {
	P services.IPermissionService
	R services.IRelationshipService
	S services.ISchemaService
}

// NewContainer -
func NewContainer() *Container {
	l := logger.New("debug")

	var err error

	var db database.Database
	db, err = factories.DatabaseFactory(config.Database{Engine: database.MEMORY.String()})
	if err != nil {
		fmt.Println(err)
	}

	err = db.Migrate(factories.MigrationFactory(database.Engine(db.GetEngineType())))
	if err != nil {
		fmt.Println(err)
	}

	// Repositories
	relationshipReader := factories.RelationshipReaderFactory(db)
	relationshipWriter := factories.RelationshipWriterFactory(db)

	schemaReader := factories.SchemaReaderFactory(db)
	schemaWriter := factories.SchemaWriterFactory(db)

	// commands
	checkCommand := commands.NewCheckCommand(keys.NewNoopCheckCommandKeys(), schemaReader, relationshipReader, l)
	expandCommand := commands.NewExpandCommand(schemaReader, relationshipReader, l)
	lookupSchemaCommand := commands.NewLookupSchemaCommand(schemaReader, l)
	lookupEntityCommand := commands.NewLookupEntityCommand(checkCommand, schemaReader, relationshipReader, l)

	return &Container{
		P: services.NewPermissionService(checkCommand, expandCommand, lookupSchemaCommand, lookupEntityCommand),
		R: services.NewRelationshipService(relationshipReader, relationshipWriter, schemaReader),
		S: services.NewSchemaService(schemaWriter, schemaReader),
	}
}
