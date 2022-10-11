package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Permify/permify/internal/servers/http/common"
)

type HealthServer struct{}

// NewHealthServer -
func NewHealthServer(handler *echo.Group) {
	r := &HealthServer{}

	h := handler.Group("/health")
	{
		h.GET("/ping", r.ping)
	}
}

// @Summary     HealthServer
// @Description
// @ID          ping
// @Tags  	    Server
// @Accept      json
// @Produce     json
// @Success     200 {object} common.Message
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /status/ping [get]
func (HealthServer) ping(c echo.Context) (err error) {
	return c.JSON(http.StatusOK, common.MResponse("pong"))
}
