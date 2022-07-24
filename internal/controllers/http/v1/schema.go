package v1

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Permify/permify/internal/controllers/http/responses"
	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/parser"
	"github.com/Permify/permify/pkg/logger"
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
	var file *multipart.FileHeader
	file, err = c.FormFile("schema")

	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(map[string]string{
			"schema": err.Error(),
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
		return echo.ErrInternalServerError
	}

	var cnf []entities.EntityConfig

	for _, st := range sch.Statements {
		cnf = append(cnf, entities.EntityConfig{
			Entity:           st.(*ast.EntityStatement).Name.Literal,
			SerializedConfig: []byte(st.String()),
		})
	}

	err = r.schemaService.Replace(context.Background(), cnf)
	if err != nil {
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusInternalServerError, responses.MResponse("success"))
}
