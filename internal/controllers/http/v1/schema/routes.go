package schema

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/commands"
	"github.com/Permify/permify/internal/controllers/http/common"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/repositories/entities"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/logger"
)

var tracer = otel.Tracer("routes")

type schemaRoutes struct {
	schemaManager managers.IEntityConfigManager
	schemaService services.ISchemaService
	l             logger.Interface
}

// NewSchemaRoutes -
func NewSchemaRoutes(handler *echo.Group, ss services.ISchemaService, m managers.IEntityConfigManager, l logger.Interface) {
	r := &schemaRoutes{m, ss, l}

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
// @Param       schema body []byte true "schema file (expected extension .perm)"
// @Success     200 {object} WriteResponse
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /schemas/write [post]
func (r *schemaRoutes) write(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "schemas.write")
	defer span.End()

	var file *multipart.FileHeader
	file, err = c.FormFile("schema")
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(map[string]string{
			"schema": err.Error(),
		}))
	}

	extension := filepath.Ext(strings.ToLower(file.Filename))
	if extension != ".perm" {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(map[string]string{
			"schema": "file extension must be .perm",
		}))
	}

	var src multipart.File
	src, err = file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, src); err != nil {
		return err
	}

	pr := parser.NewParser(string(buf.Bytes()))
	sch := pr.Parse()
	if pr.Error() != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(map[string]string{
			"schema": pr.Error().Error(),
		}))
	}

	err = sch.Validate()
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(map[string]string{
			"schema": "must have user entity",
		}))
	}

	var cnf []entities.EntityConfig
	for _, st := range sch.Statements {
		cnf = append(cnf, entities.EntityConfig{
			Entity:           st.(*ast.EntityStatement).Name.Literal,
			SerializedConfig: []byte(st.String()),
		})
	}

	var version string
	version, err = r.schemaManager.Write(ctx, cnf)
	if err != nil {
		span.SetStatus(codes.Error, echo.ErrInternalServerError.Error())
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, common.SuccessResponse(WriteResponse{
		Version: version,
	}))
}

// @Summary     Schema
// @Description read your authorization model
// @ID          schemas.read
// @Tags  	    Schema
// @Accept      json
// @Produce     json
// @Param       request body ReadRequest true "read your authorization model"
// @Success     200 {object} ReadResponse
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /schemas/read [post]
func (r *schemaRoutes) read(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "schemas.read")
	defer span.End()

	request := new(ReadRequest)
	if err = (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	var response schema.Schema
	response, err = r.schemaManager.All(ctx, request.SchemaVersion.String())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, common.SuccessResponse(ReadResponse{
		Entities: response.Entities,
	}))
}

// @Summary     Schema
// @Description lookup your authorization model
// @ID          schemas.lookup
// @Tags  	    Schema
// @Accept      json
// @Produce     json
// @Param       request body LookupRequest true "lookup your authorization model"
// @Success     200 {object} LookupResponse
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /schemas/lookup [post]
func (r *schemaRoutes) lookup(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "schemas.lookup")
	defer span.End()

	request := new(LookupRequest)
	if err = (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	var response commands.SchemaLookupResponse
	response, err = r.schemaService.Lookup(ctx, request.EntityType, request.RelationNames, request.SchemaVersion.String())

	if err != nil {
		span.SetStatus(codes.Error, echo.ErrInternalServerError.Error())
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, common.SuccessResponse(LookupResponse{
		ActionNames: response.ActionNames,
	}))
}
