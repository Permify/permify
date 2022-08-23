package v1

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	// Swagger docs.
	_ "github.com/Permify/permify/docs"
	"github.com/Permify/permify/internal/managers"

	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/logger"
)

// NewRouter -.
// Swagger spec:
// @title       Permify API
// @description
// @version     1.0
// @host        localhost:3476
// @BasePath    /v1
func NewRouter(handler *echo.Echo, l logger.Interface, r services.IRelationshipService, t services.IPermissionService, s managers.IEntityConfigManager) {
	// Options
	handler.Use(middleware.Logger())
	handler.Use(middleware.Recover())
	handler.Use(middleware.CORS())

	// Swagger
	handler.GET("/doc/*", echoSwagger.WrapHandler)

	// Routers
	h := handler.Group("/v1")
	{
		newPermissionRoutes(h, t, l)
		newRelationshipRoutes(h, r, l)
		newSchemaRoutes(h, s, l)
		newServerRoutes(h)
	}
}
