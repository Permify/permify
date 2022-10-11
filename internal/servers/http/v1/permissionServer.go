package v1

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/servers/http/common"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

// PermissionServer -
type PermissionServer struct {
	service services.IPermissionService
	l       logger.Interface
}

// NewPermissionServer -
func NewPermissionServer(handler *echo.Group, t services.IPermissionService, l logger.Interface) {
	r := &PermissionServer{t, l}

	h := handler.Group("/permissions")
	{
		h.POST("/check", r.check)
		h.POST("/expand", r.expand)
		h.POST("/lookup-query", r.lookupQuery)
	}
}

// @Summary     Permission
// @Description check subject is authorized
// @ID          permissions.check
// @Tags  	    Permission
// @Accept      json
// @Produce     json
// @Param       request body *v1.CheckRequest true "check subject is authorized"
// @Success     200 {object} *v1.CheckResponse
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /permissions/check [post]
func (r *PermissionServer) check(c echo.Context) error {
	ctx, span := tracer.Start(c.Request().Context(), "permissions.check")
	defer span.End()

	request := new(v1.CheckRequest)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	var err errors.Error

	var depth int32 = 20
	if request.Depth != nil {
		depth = request.Depth.Value
	}

	var response commands.CheckResponse
	response, err = r.service.Check(ctx, request.GetSubject(), request.GetAction(), request.GetEntity(), "", depth)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		switch err.Kind() {
		case errors.Database:
			return c.JSON(database.GetKindToHttpStatus(err.SubKind()), common.MResponse(err.Error()))
		case errors.Validation:
			return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(err.Params()))
		case errors.Service:
			return c.JSON(http.StatusInternalServerError, common.MResponse(err.Error()))
		default:
			return c.JSON(http.StatusInternalServerError, common.MResponse(err.Error()))
		}
	}

	return c.JSON(http.StatusOK, &v1.CheckResponse{
		Can: response.Can,
	})
}

// @Summary     Permission
// @Description expand relationships according to schema
// @ID          permissions.expand
// @Tags  	    Permission
// @Accept      json
// @Produce     json
// @Param       request body *v1.ExpandRequest true "expand relationships according to schema"
// @Success     200 {object} *v1.ExpandResponse
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /permissions/expand [post]
func (r *PermissionServer) expand(c echo.Context) error {
	ctx, span := tracer.Start(c.Request().Context(), "permissions.expand")
	defer span.End()

	request := new(v1.ExpandRequest)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	var err errors.Error

	var response commands.ExpandResponse
	response, err = r.service.Expand(ctx, request.GetEntity(), request.GetAction(), request.GetSchemaVersion())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		switch err.Kind() {
		case errors.Database:
			return c.JSON(database.GetKindToHttpStatus(err.SubKind()), common.MResponse(err.Error()))
		case errors.Validation:
			return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(err.Params()))
		case errors.Service:
			return c.JSON(http.StatusInternalServerError, common.MResponse(err.Error()))
		default:
			return c.JSON(http.StatusInternalServerError, common.MResponse(err.Error()))
		}
	}

	return c.JSON(http.StatusOK, &v1.ExpandResponse{
		Tree: response.Tree,
	})
}

// @Summary     Permission
// @Description lookupQuery
// @ID          permissions.lookupQuery
// @Tags  	    Permission
// @Accept      json
// @Produce     json
// @Param       request body *v1.LookupQueryRequest true "''"
// @Success     200 {object} *v1.LookupQueryResponse
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /permissions/lookup-query [post]
func (r *PermissionServer) lookupQuery(c echo.Context) error {
	ctx, span := tracer.Start(c.Request().Context(), "permissions.lookupQuery")
	defer span.End()

	request := new(v1.LookupQueryRequest)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	var err errors.Error
	var response commands.LookupQueryResponse
	response, err = r.service.LookupQuery(ctx, request.EntityType, request.Subject, request.Action, request.SchemaVersion)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.l.Error(fmt.Sprintf(err.Error()))
		switch err.Kind() {
		case errors.Database:
			return c.JSON(database.GetKindToHttpStatus(err.SubKind()), common.MResponse(err.Error()))
		case errors.Validation:
			return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(err.Params()))
		case errors.Service:
			return c.JSON(http.StatusInternalServerError, common.MResponse(err.Error()))
		default:
			return c.JSON(http.StatusInternalServerError, common.MResponse(err.Error()))
		}
	}

	return c.JSON(http.StatusOK, &v1.LookupQueryResponse{
		Query: response.Query,
		Args:  response.Args,
	})
}
