package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Permify/permify/internal/controllers/http/requests/permission"
	"github.com/Permify/permify/pkg/logger"

	"github.com/Permify/permify/internal/controllers/http/responses"
	"github.com/Permify/permify/internal/services"
)

// permissionRoutes -
type permissionRoutes struct {
	service services.IPermissionService
	logger  logger.Interface
}

// newPermissionRoutes -
func newPermissionRoutes(handler *echo.Group, t services.IPermissionService, l logger.Interface) {
	r := &permissionRoutes{t, l}

	h := handler.Group("/permissions")
	{
		h.POST("/check", r.check)
	}
}

// @Summary     Permission
// @Description Check subject is authorized
// @ID          check
// @Tags  	    Permission
// @Accept      json
// @Produce     json
// @Param       request body permission.Check true "''"
// @Success     200 {object} responses.Check
// @Failure     400 {object} responses.HTTPErrorResponse
// @Router      /permissions/check [post]
func (r *permissionRoutes) check(c echo.Context) (err error) {
	request := new(permission.Check)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request.Body); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(v))
	}

	if request.Body.Depth == 0 {
		request.Body.Depth = 8
	}

	var can bool
	var vi *services.VisitMap
	can, vi, err = r.service.Check(context.Background(), request.Body.User, request.Body.Action, request.Body.Object, request.Body.Depth)
	if err != nil {
		if errors.Is(err, services.DepthError) {
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{"depth": "depth is not enough to check"})
		}
		if errors.Is(err, services.ActionCannotFoundError) {
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{"action": "action cannot be found"})
		}
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, responses.Check{
		Can:       can,
		Decisions: vi,
	})
}
