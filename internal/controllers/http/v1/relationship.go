package v1

import (
	"errors"
	"net/http"

	"go.opentelemetry.io/otel/codes"

	"github.com/labstack/echo/v4"

	"github.com/Permify/permify/internal/controllers/http/requests/relationship"
	"github.com/Permify/permify/internal/controllers/http/responses"
	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/logger"
)

// relationshipRoutes -
type relationshipRoutes struct {
	relationshipService services.IRelationshipService
	logger              logger.Interface
}

// newRelationshipRoutes -
func newRelationshipRoutes(handler *echo.Group, t services.IRelationshipService, l logger.Interface) {
	r := &relationshipRoutes{t, l}

	h := handler.Group("/relationships")
	{
		h.GET("/read", r.read)
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
// @Param       request body relationship.Write true "''"
// @Success     200 {object} []tuple.Tuple
// @Failure     400 {object} responses.HTTPErrorResponse
// @Router      /relationships/read [post]
func (r *relationshipRoutes) read(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "relationships.read")
	defer span.End()

	request := new(relationship.Read)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request.Body); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(v))
	}

	var tuples []entities.RelationTuple
	tuples, err = r.relationshipService.ReadRelationships(ctx, request.Body.Filter)
	if err != nil {
		span.SetStatus(codes.Error, echo.ErrInternalServerError.Error())
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, responses.SuccessResponse(entities.RelationTuples(tuples).ToTuple()))
}

// @Summary     Relationship
// @Description create new relation tuple
// @ID          relationships.write
// @Tags  	    Relationship
// @Accept      json
// @Produce     json
// @Param       request body relationship.Write true "''"
// @Success     200 {object} tuple.Tuple
// @Failure     400 {object} responses.HTTPErrorResponse
// @Router      /relationships/write [post]
func (r *relationshipRoutes) write(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "relationships.write")
	defer span.End()

	request := new(relationship.Write)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request.Body); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(v))
	}

	tuple := entities.RelationTuple{Entity: request.Body.Entity.Type, ObjectID: request.Body.Entity.ID, Relation: request.Body.Relation, UsersetEntity: request.Body.Subject.Type, UsersetObjectID: request.Body.Subject.ID, UsersetRelation: request.Body.Subject.Relation.String(), Type: "custom"}
	err = r.relationshipService.WriteRelationship(ctx, []entities.RelationTuple{tuple})
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			span.RecordError(database.ErrUniqueConstraint)
			return c.JSON(http.StatusUnprocessableEntity, responses.MResponse("tuple already exists"))
		}
		span.SetStatus(codes.Error, echo.ErrInternalServerError.Error())
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, responses.SuccessResponse(tuple.ToTuple()))
}

// @Summary     Relationship
// @Description delete relation tuple
// @ID          relationships.delete
// @Tags  	    Relationship
// @Accept      json
// @Produce     json
// @Param       request body relationship.Delete true "''"
// @Success     200 {object} tuple.Tuple
// @Failure     400 {object} responses.HTTPErrorResponse
// @Router      /relationships/delete [post]
func (r *relationshipRoutes) delete(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "relationships.delete")
	defer span.End()

	request := new(relationship.Delete)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request.Body); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(v))
	}

	tuple := entities.RelationTuple{Entity: request.Body.Entity.Type, ObjectID: request.Body.Entity.ID, Relation: request.Body.Relation, UsersetEntity: request.Body.Subject.Type, UsersetObjectID: request.Body.Subject.ID, UsersetRelation: request.Body.Subject.Relation.String(), Type: "custom"}
	err = r.relationshipService.DeleteRelationship(ctx, []entities.RelationTuple{tuple})
	if err != nil {
		span.SetStatus(codes.Error, echo.ErrInternalServerError.Error())
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, responses.SuccessResponse(tuple.ToTuple()))
}
