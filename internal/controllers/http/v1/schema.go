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

	req "github.com/Permify/permify/internal/controllers/http/requests/schema"
	"github.com/Permify/permify/internal/controllers/http/responses"
	res "github.com/Permify/permify/internal/controllers/http/responses/schema"
	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/dsl/schema"
	"github.com/Permify/permify/pkg/logger"
	"github.com/Permify/permify/pkg/tuple"
)

type schemaRoutes struct {
	schemaService services.ISchemaService
	l             logger.Interface
}

// newSchemaRoutes -
func newSchemaRoutes(handler *echo.Group, s services.ISchemaService, l logger.Interface) {
	r := &schemaRoutes{s, l}

	h := handler.Group("/schemas")
	{
		h.POST("/write", r.write)
		h.GET("/read/:schema_version", r.read)
		h.GET("/read", r.read)
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

	var isValid bool
	var cnf []entities.EntityConfig

	for _, st := range sch.Statements {
		name := st.(*ast.EntityStatement).Name.Literal
		if name == tuple.USER {
			isValid = true
		}

		cnf = append(cnf, entities.EntityConfig{
			Entity:           name,
			SerializedConfig: []byte(st.String()),
		})
	}

	if !isValid {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(map[string]string{
			"schema": "must have user entity",
		}))
	}

	var version string
	version, err = r.schemaService.Write(ctx, cnf)
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
// @Failure     400 {object} []schema.Entity
// @Router      /schemas/read [get]
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
	response, err = r.schemaService.All(ctx, request.PathParams.SchemaVersion.String())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, responses.SuccessResponse(response.Entities))
}
