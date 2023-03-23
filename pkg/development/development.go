package development

import (
	"fmt"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/internal/engines"
	"github.com/Permify/permify/internal/factories"
	"github.com/Permify/permify/internal/keys"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/logger"
)

// Container - Structure for container instance
type Container struct {
	P services.IPermissionService
	R services.IRelationshipService
	S services.ISchemaService
}

// NewContainer - Creates new container instance
func NewContainer() *Container {
	var err error

	var db database.Database
	db, err = factories.DatabaseFactory(config.Database{Engine: database.MEMORY.String()})
	if err != nil {
		fmt.Println(err)
	}

	l := logger.New("debug")

	// Repositories
	relationshipReader := factories.RelationshipReaderFactory(db, l)
	relationshipWriter := factories.RelationshipWriterFactory(db, l)

	schemaReader := factories.SchemaReaderFactory(db, l)
	schemaWriter := factories.SchemaWriterFactory(db, l)

	// engines
	checkEngine := engines.NewCheckEngine(keys.NewNoopCheckEngineKeys(), schemaReader, relationshipReader)
	expandEngine := engines.NewExpandEngine(schemaReader, relationshipReader)
	lookupSchemaEngine := engines.NewLookupSchemaEngine(schemaReader)
	lookupEntityEngine := engines.NewLookupEntityEngine(checkEngine, relationshipReader)

	return &Container{
		P: services.NewPermissionService(checkEngine, expandEngine, lookupSchemaEngine, lookupEntityEngine),
		R: services.NewRelationshipService(relationshipReader, relationshipWriter, schemaReader),
		S: services.NewSchemaService(schemaWriter, schemaReader),
	}
}
