package v1

import (
	"errors"
	"net/http"

	"go.opentelemetry.io/otel/codes"

	"github.com/labstack/echo/v4"

	"github.com/Permify/permify/internal/commands"
	req "github.com/Permify/permify/internal/controllers/http/requests/permission"
	"github.com/Permify/permify/internal/controllers/http/responses"
	res "github.com/Permify/permify/internal/controllers/http/responses/permission"
	internalErrors "github.com/Permify/permify/internal/internal-errors"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/logger"
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
		h.POST("/expand", r.expand)
	}
}

// @Summary     Permission
// @Description Check subject is authorized
// @ID          permissions.check
// @Tags  	    Permission
// @Accept      json
// @Produce     json
// @Param       request body permission.CheckRequest true "''"
// @Success     200 {object} permission.ExpandResponse
// @Failure     400 {object} responses.HTTPErrorResponse
// @Router      /permissions/check [post]
func (r *permissionRoutes) check(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "permissions.check")
	defer span.End()

	request := new(req.CheckRequest)
	if err = (&echo.DefaultBinder{}).BindBody(c, &request.Body); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(v))
	}

	if request.Body.Depth == 0 {
		request.Body.Depth = 20
	}

	var response commands.CheckResponse
	response, err = r.service.Check(ctx, request.Body.Subject, request.Body.Action, request.Body.Entity, request.Body.SchemaVersion.String(), request.Body.Depth)
	if err != nil {
		if errors.Is(err, internalErrors.DepthError) {
			span.RecordError(internalErrors.DepthError)
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{"depth": "depth is not enough to check"})
		}
		if errors.Is(err, internalErrors.ActionCannotFoundError) {
			span.RecordError(internalErrors.ActionCannotFoundError)
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{"action": "action cannot be found"})
		}
		if errors.Is(err, internalErrors.EntityConfigCannotFoundError) {
			span.RecordError(internalErrors.EntityConfigCannotFoundError)
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{"entity": "entity config cannot be found"})
		}
		span.SetStatus(codes.Error, echo.ErrInternalServerError.Error())
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, res.CheckResponse{
		Can:            response.Can,
		RemainingDepth: response.RemainingDepth,
		Decisions:      response.Visits,
	})
}

// @Summary     Permission
// @Description
// @ID          permissions.expand
// @Tags  	    Permission
// @Accept      json
// @Produce     json
// @Param       request body permission.ExpandRequest true "''"
// @Success     200 {object} permission.ExpandResponse
// @Failure     400 {object} responses.HTTPErrorResponse
// @Router      /permissions/expand [post]
func (r *permissionRoutes) expand(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "permissions.expand")
	defer span.End()

	request := new(req.ExpandRequest)
	if err = (&echo.DefaultBinder{}).BindBody(c, &request.Body); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(v))
	}

	var response commands.ExpandResponse
	response, err = r.service.Expand(ctx, request.Body.Entity, request.Body.Action, request.Body.SchemaVersion.String())
	if err != nil {
		if errors.Is(err, internalErrors.DepthError) {
			span.RecordError(internalErrors.DepthError)
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{"depth": "depth is not enough to check"})
		}
		if errors.Is(err, internalErrors.ActionCannotFoundError) {
			span.RecordError(internalErrors.ActionCannotFoundError)
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{"action": "action cannot be found"})
		}
		if errors.Is(err, internalErrors.EntityConfigCannotFoundError) {
			span.RecordError(internalErrors.EntityConfigCannotFoundError)
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{"entity": "entity config cannot be found"})
		}
		span.SetStatus(codes.Error, echo.ErrInternalServerError.Error())
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, res.ExpandResponse{
		Tree: response.Tree,
	})
}
