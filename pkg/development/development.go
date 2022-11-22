package development

import (
	"github.com/Permify/permify/internal/services"
)

// Container - Structure for container instance
type Container struct {
	P services.IPermissionService
	R services.IRelationshipService
	S services.ISchemaService
}

// NewContainer - Creates new container instance
func NewContainer() *Container {
	//l := logger.New("debug")
	//
	//var err error

	//var db database.Database
	//db, err = factories.DatabaseFactory(config.Database{Engine: database.MEMORY.String()})
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//err = db.Migrate(factories.MigrationFactory(database.Engine(db.GetEngineType())))
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//// defer db.Close()
	//
	//// Repositories
	//relationTupleRepository := factories.RelationTupleFactory(db)
	//entityConfigRepository := factories.EntityConfigFactory(db)
	//
	//// manager
	//schemaManager := managers.NewEntityConfigManager(entityConfigRepository, nil)
	//
	//// commands
	//checkCommand := commands.NewCheckCommand(relationTupleRepository, l)
	//expandCommand := commands.NewExpandCommand(relationTupleRepository, l)
	//lookupQueryCommand := commands.NewLookupQueryCommand(relationTupleRepository, l)
	//
	//// Services
	//relationshipService := services.NewRelationshipService(relationTupleRepository, schemaManager)
	//permissionService := services.NewPermissionService(checkCommand, expandCommand, lookupQueryCommand, schemaManager)

	return &Container{
		P: nil,
		R: nil,
		S: nil,
	}
}
