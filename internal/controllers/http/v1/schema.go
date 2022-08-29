package v1

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"go.opentelemetry.io/otel/codes"

	"github.com/labstack/echo/v4"

	"github.com/Permify/permify/internal/commands"
	req "github.com/Permify/permify/internal/controllers/http/requests/schema"
	"github.com/Permify/permify/internal/controllers/http/responses"
	res "github.com/Permify/permify/internal/controllers/http/responses/schema"
	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/managers"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/logger"
)

type schemaRoutes struct {
	schemaManager managers.IEntityConfigManager
	schemaService services.ISchemaService
	l             logger.Interface
}

// newSchemaRoutes -
func newSchemaRoutes(handler *echo.Group, ss services.ISchemaService, m managers.IEntityConfigManager, l logger.Interface) {
	r := &schemaRoutes{m, ss, l}

	h := handler.Group("/schemas")
	{
		h.POST("/write", r.write)
		h.GET("/read/:schema_version", r.read)
		h.GET("/read", r.read)
		h.POST("/lookup", r.lookup)
	}
}

// @Summary     Schema
// @Description replace your authorization model
// @ID          schemas.write
// @Tags  	    Schema
// @Accept      json
// @Produce     json
// @Success     200 {object} schema.WriteResponse
// @Failure     400 {object} responses.HTTPErrorResponse
// @Router      /schemas/write [post]
func (r *schemaRoutes) write(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "schemas.replace")
	defer span.End()

	var file *multipart.FileHeader
	file, err = c.FormFile("schema")
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(map[string]string{
			"schema": err.Error(),
		}))
	}

	extension := filepath.Ext(strings.ToLower(file.Filename))
	if extension != ".perm" {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(map[string]string{
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
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(map[string]string{
			"schema": pr.Error().Error(),
		}))
	}

	err = sch.Validate()
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(map[string]string{
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

	return c.JSON(http.StatusOK, responses.SuccessResponse(res.WriteResponse{
		Version: version,
	}))
}

// @Summary     Schema
// @Description read your authorization model
// @ID          schemas.read
// @Tags  	    Schema
// @Accept      json
// @Produce     json
// @Param       request body schema.ReadRequest true "''"
// @Success     200 {object} responses.Message
// @Failure     400 {object} responses.HTTPErrorResponse
// @Router      /schemas/read/:schema_version [get]
func (r *schemaRoutes) read(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "schemas.read")
	defer span.End()

	request := new(req.ReadRequest)
	if err = (&echo.DefaultBinder{}).BindPathParams(c, &request.Body); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(v))
	}

	var response schema.Schema
	response, err = r.schemaManager.All(ctx, request.PathParams.SchemaVersion.String())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, responses.SuccessResponse(response.Entities))
}

// @Summary     Schema
// @Description lookup your authorization model
// @ID          schemas.lookup
// @Tags  	    Schema
// @Accept      json
// @Produce     json
// @Param       request body schema.LookupRequest true "''"
// @Success     200 {object} schema.LookupResponse
// @Failure     400 {object} responses.HTTPErrorResponse
// @Router      /schemas/lookup [post]
func (r *schemaRoutes) lookup(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "schemas.lookup")
	defer span.End()

	request := new(req.LookupRequest)
	if err = (&echo.DefaultBinder{}).BindBody(c, &request.Body); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(v))
	}

	var response commands.SchemaLookupResponse
	response, err = r.schemaService.Lookup(ctx, request.Body.EntityType, request.Body.RelationNames, request.Body.SchemaVersion.String())

	if err != nil {
		span.SetStatus(codes.Error, echo.ErrInternalServerError.Error())
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, responses.SuccessResponse(res.LookupResponse{
		ActionNames: response.ActionNames,
	}))
}
