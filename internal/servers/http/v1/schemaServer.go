package v1

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/servers/http/common"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
)

type SchemaServer struct {
	schemaManager managers.IEntityConfigManager
	schemaService services.ISchemaService
	l             logger.Interface
}

// NewSchemaServer -
func NewSchemaServer(handler *echo.Group, ss services.ISchemaService, m managers.IEntityConfigManager, l logger.Interface) {
	r := &SchemaServer{m, ss, l}

	h := handler.Group("/schemas")
	{
		h.POST("/write", r.write)
		h.POST("/read", r.read)
		h.POST("/lookup", r.lookup)
	}
}

// @Summary     Schema
// @Description write your authorization model
// @ID          schemas.write
// @Tags  	    Schema
// @Accept      json
// @Produce     json
// @Param       request body *v1.SchemaWriteRequest true "read your authorization model"
// @Success     200 {object} *v1.SchemaWriteResponse
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /schemas/write [post]
func (r *SchemaServer) write(c echo.Context) error {
	ctx, span := tracer.Start(c.Request().Context(), "schemas.write")
	defer span.End()

	request := new(v1.SchemaWriteRequest)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	var err errors.Error
	var version string
	version, err = r.schemaManager.Write(ctx, request.Schema)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.l.Info(fmt.Sprintf(err.Error()))
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

	return c.JSON(http.StatusOK, &v1.SchemaWriteResponse{
		SchemaVersion: version,
	})
}

// @Summary     Schema
// @Description read your authorization model
// @ID          schemas.read
// @Tags  	    Schema
// @Accept      json
// @Produce     json
// @Param       request body *v1.SchemaReadRequest true "read your authorization model"
// @Success     200 {object} *v1.SchemaReadResponse
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /schemas/read [post]
func (r *SchemaServer) read(c echo.Context) error {
	// ctx, span := tracer.Start(c.Request().Context(), "schemas.read")
	// defer span.End()

	//request := new(ReadRequest)
	//if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
	//	return err
	//}
	//v := request.Validate()
	//if v != nil {
	//	return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	//}

	var err errors.Error

	// var response schema.Schema
	// response, err = r.schemaManager.All(ctx, request.SchemaVersion.String())
	if err != nil {
		// span.RecordError(err)
		// span.SetStatus(codes.Error, err.Error())
		r.l.Info(fmt.Sprintf(err.Error()))
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

	return nil
	//return c.JSON(http.StatusOK, common.SuccessResponse(ReadResponse{
	//	Entities: response.Entities,
	//}))
}

// @Summary     Schema
// @Description lookup your authorization model
// @ID          schemas.lookup
// @Tags  	    Schema
// @Accept      json
// @Produce     json
// @Param       request body *v1.SchemaLookupRequest true "lookup your authorization model"
// @Success     200 {object} *v1.SchemaLookupResponse
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /schemas/lookup [post]
func (r *SchemaServer) lookup(c echo.Context) error {
	ctx, span := tracer.Start(c.Request().Context(), "schemas.lookup")
	defer span.End()

	request := new(v1.SchemaLookupRequest)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	var err errors.Error
	var response commands.SchemaLookupResponse
	response, err = r.schemaService.Lookup(ctx, request.EntityType, request.RelationNames, request.SchemaVersion)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		r.l.Info(fmt.Sprintf(err.Error()))
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

	return c.JSON(http.StatusOK, &v1.SchemaLookupResponse{
		ActionNames: response.ActionNames,
	})
}
