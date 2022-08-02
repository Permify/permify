package v1

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/Permify/permify/internal/controllers/http/responses"
	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/parser"
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
		h.POST("/replace", r.replace)
	}
}

// @Summary     Schema
// @Description replace your authorization model
// @ID          replace
// @Tags  	    Schema
// @Accept      json
// @Produce     json
// @Success     200 {object} responses.Message
// @Failure     400 {object} responses.HTTPErrorResponse
// @Router      /schemas/replace [post]
func (r *schemaRoutes) replace(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "replace")
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

	err = r.schemaService.Replace(ctx, cnf)
	if err != nil {
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, responses.MResponse("success"))
}
