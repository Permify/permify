package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	swagger "github.com/swaggo/echo-swagger"
	"go.opentelemetry.io/otel"

	// Swagger docs.
	_ "github.com/Permify/permify/docs"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/logger"
)

var tracer = otel.Tracer("http.servers")

// NewServer -.
// Swagger spec:
// @title       Permify API
// @description Permify is an open-source authorization service for creating and maintaining fine-grained authorizations across your individual applications and services.
// @description Permify converts authorization data as relational tuples into a database you point at. We called that database a Write Database (WriteDB) and it behaves as a centralized data source for your authorization system. You can model of your authorization with Permify's DSL - Permify Schema - and perform access checks with a single API call anywhere on your stack. Access decisions made according to stored relational tuples.
// @version     v0.0.0-alpha7
// @contact.name API Support
// @contact.url https://github.com/Permify/permify/issues
// @contact.email hello@permify.co
// @license.name GNU Affero General Public License v3.0
// @schemes 	http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name api-key-auth
// @license.url https://github.com/Permify/permify/blob/master/LICENSE
// @host        localhost:3476
// @BasePath    /v1
func NewServer(handler *echo.Echo, l logger.Interface, r services.IRelationshipService, t services.IPermissionService, s services.ISchemaService, e managers.IEntityConfigManager) {
	// Options
	handler.Use(middleware.Logger())
	handler.Use(middleware.Recover())
	handler.Use(middleware.CORS())

	// Swagger
	handler.GET("/docs/*", swagger.WrapHandler)

	// Routers
	h := handler.Group("/v1")
	{
		NewPermissionServer(h, t, l)
		NewRelationshipServer(h, r, l)
		NewSchemaServer(h, s, e, l)
		NewHealthServer(h)
	}
}
