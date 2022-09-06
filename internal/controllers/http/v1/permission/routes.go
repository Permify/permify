package permission

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/controllers/http/common"
	internalErrors "github.com/Permify/permify/internal/internal-errors"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/logger"
)

var tracer = otel.Tracer("routes")

// permissionRoutes -
type permissionRoutes struct {
	service services.IPermissionService
	logger  logger.Interface
}

// NewPermissionRoutes -
func NewPermissionRoutes(handler *echo.Group, t services.IPermissionService, l logger.Interface) {
	r := &permissionRoutes{t, l}

	h := handler.Group("/permissions")
	{
		h.POST("/check", r.check)
		h.POST("/expand", r.expand)
	}
}

// @Summary     Permission
// @Description check subject is authorized
// @ID          permissions.check
// @Tags  	    Permission
// @Accept      json
// @Produce     json
// @Param       request body CheckRequest true "check subject is authorized"
// @Success     200 {object} CheckResponse
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /permissions/check [post]
func (r *permissionRoutes) check(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "permissions.check")
	defer span.End()

	request := new(CheckRequest)
	if err = (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	if request.Depth == 0 {
		request.Depth = 20
	}

	var response commands.CheckResponse
	response, err = r.service.Check(ctx, request.Subject, request.Action, request.Entity, request.SchemaVersion.String(), request.Depth)
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

	return c.JSON(http.StatusOK, CheckResponse{
		Can:            response.Can,
		RemainingDepth: response.RemainingDepth,
		Decisions:      response.Visits,
	})
}

// @Summary     Permission
// @Description expand relationships according to schema
// @ID          permissions.expand
// @Tags  	    Permission
// @Accept      json
// @Produce     json
// @Param       request body ExpandRequest true "expand relationships according to schema"
// @Success     200 {object} ExpandResponse
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /permissions/expand [post]
func (r *permissionRoutes) expand(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "permissions.expand")
	defer span.End()

	request := new(ExpandRequest)
	if err = (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	var response commands.ExpandResponse
	response, err = r.service.Expand(ctx, request.Entity, request.Action, request.SchemaVersion.String())
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

	return c.JSON(http.StatusOK, ExpandResponse{
		Tree: response.Tree,
	})
}
