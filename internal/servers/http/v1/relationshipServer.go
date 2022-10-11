package v1

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/servers/http/common"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/errors"
	"github.com/Permify/permify/pkg/logger"
	v1 "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/tuple"
)

// RelationshipServer -
type RelationshipServer struct {
	relationshipService services.IRelationshipService
	l                   logger.Interface
}

// NewRelationshipServer -
func NewRelationshipServer(handler *echo.Group, t services.IRelationshipService, l logger.Interface) {
	r := &RelationshipServer{t, l}

	h := handler.Group("/relationships")
	{
		h.POST("/read", r.read)
		h.POST("/write", r.write)
		h.POST("/delete", r.delete)
	}
}

// @Summary     Relationship
// @Description read relation tuple(s)
// @ID          relationships.read
// @Tags  	    Relationship
// @Accept      json
// @Produce     json
// @Param       request body *v1.RelationshipReadRequest true "read relation tuple(s)"
// @Success     200 {object} []*v1.Tuple
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /relationships/read [post]
func (r *RelationshipServer) read(c echo.Context) error {
	ctx, span := tracer.Start(c.Request().Context(), "relationships.read")
	defer span.End()

	request := new(v1.RelationshipReadRequest)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	var err errors.Error

	var collection tuple.ITupleCollection
	collection, err = r.relationshipService.ReadRelationships(ctx, request.GetFilter())
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

	return c.JSON(http.StatusOK, &v1.RelationshipReadResponse{
		Tuples: collection.GetTuples(),
	})
}

// @Summary     Relationship
// @Description create new relation tuple
// @ID          relationships.write
// @Tags  	    Relationship
// @Accept      json
// @Produce     json
// @Param       request body *v1.RelationshipWriteRequest true "create new relation tuple"
// @Success     200 {object} *v1.Tuple
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /relationships/write [post]
func (r *RelationshipServer) write(c echo.Context) error {
	ctx, span := tracer.Start(c.Request().Context(), "relationships.write")
	defer span.End()

	request := new(v1.RelationshipWriteRequest)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	var err errors.Error

	t := &v1.Tuple{Entity: request.Entity, Relation: request.Relation, Subject: request.Subject}
	err = r.relationshipService.WriteRelationship(ctx, t, request.SchemaVersion)
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

	return c.JSON(http.StatusOK, common.SuccessResponse(t))
}

// @Summary     Relationship
// @Description delete relation tuple
// @ID          relationships.delete
// @Tags  	    Relationship
// @Accept      json
// @Produce     json
// @Param       request body *v1.RelationshipDeleteRequest true "delete relation tuple"
// @Success     200 {object} *v1.Tuple
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /relationships/delete [post]
func (r *RelationshipServer) delete(c echo.Context) error {
	ctx, span := tracer.Start(c.Request().Context(), "relationships.delete")
	defer span.End()

	request := new(v1.RelationshipDeleteRequest)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	var err errors.Error
	t := &v1.Tuple{Entity: request.Entity, Relation: request.Relation, Subject: request.Subject}
	err = r.relationshipService.DeleteRelationship(ctx, t)
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

	return c.JSON(http.StatusOK, common.SuccessResponse(t))
}
