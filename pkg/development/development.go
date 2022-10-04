package development

import (
	"fmt"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/logger"
)

// Development -
type Development struct {
	P services.IPermissionService
	R services.IRelationshipService
	M managers.IEntityConfigManager
}

// NewDevelopment -
func NewDevelopment() *Development {
	l := logger.New("debug")

	var err error

	var db database.Database
	db, err = factories.DatabaseFactory(config.Write{Connection: database.MEMORY.String()})
	if err != nil {
		db.Close()
		fmt.Println(err)
	}
	// defer db.Close()

	// Repositories
	relationTupleRepository := factories.RelationTupleFactory(db)
	err = relationTupleRepository.Migrate()
	if err != nil {
		fmt.Println(err)
	}

	entityConfigRepository := factories.EntityConfigFactory(db)
	err = entityConfigRepository.Migrate()
	if err != nil {
		fmt.Println(err)
	}

	// manager
	schemaManager := managers.NewEntityConfigManager(entityConfigRepository, nil)

	// commands
	checkCommand := commands.NewCheckCommand(relationTupleRepository, l)
	expandCommand := commands.NewExpandCommand(relationTupleRepository, l)
	lookupQueryCommand := commands.NewLookupQueryCommand(relationTupleRepository, l)

	// Services
	relationshipService := services.NewRelationshipService(relationTupleRepository, schemaManager)
	permissionService := services.NewPermissionService(checkCommand, expandCommand, lookupQueryCommand, schemaManager)

	return &Development{
		P: permissionService,
		R: relationshipService,
		M: schemaManager,
	}
}
