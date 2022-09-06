package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Permify/permify/internal/controllers/http/common"
)

type serverRoutes struct{}

func newServerRoutes(handler *echo.Group) {
	r := &serverRoutes{}

	h := handler.Group("/status")
	{
		h.GET("/ping", r.ping)
		h.GET("/version", r.version)
	}
}

// @Summary     Server
// @Description
// @ID          ping
// @Tags  	    Server
// @Accept      json
// @Produce     json
// @Success     200 {object} common.Message
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /status/ping [get]
func (serverRoutes) ping(c echo.Context) (err error) {
	return c.JSON(http.StatusOK, common.MResponse("pong"))
}

// @Summary     Server
// @Description
// @ID          version
// @Tags  	    Server
// @Accept      json
// @Produce     json
// @Success     200 {object} common.Message
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /status/version [get]
func (serverRoutes) version(c echo.Context) (err error) {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"version": "v0.0.0-alpha3",
	})
}
