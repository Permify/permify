package relationship

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/Permify/permify/internal/controllers/http/common"
	internalErrors "github.com/Permify/permify/internal/internal-errors"
	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/database"
	"github.com/Permify/permify/pkg/logger"
	"github.com/Permify/permify/pkg/tuple"
)

var tracer = otel.Tracer("routes")

// relationshipRoutes -
type relationshipRoutes struct {
	relationshipService services.IRelationshipService
	logger              logger.Interface
}

// NewRelationshipRoutes -
func NewRelationshipRoutes(handler *echo.Group, t services.IRelationshipService, l logger.Interface) {
	r := &relationshipRoutes{t, l}

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
// @Param       request body ReadRequest true "read relation tuple(s)"
// @Success     200 {object} []tuple.Tuple
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /relationships/read [post]
func (r *relationshipRoutes) read(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "relationships.read")
	defer span.End()

	request := new(ReadRequest)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	var tuples []tuple.Tuple
	tuples, err = r.relationshipService.ReadRelationships(ctx, request.Filter)
	if err != nil {
		span.SetStatus(codes.Error, echo.ErrInternalServerError.Error())
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, common.SuccessResponse(tuples))
}

// @Summary     Relationship
// @Description create new relation tuple
// @ID          relationships.write
// @Tags  	    Relationship
// @Accept      json
// @Produce     json
// @Param       request body WriteRequest true "create new relation tuple"
// @Success     200 {object} tuple.Tuple
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /relationships/write [post]
func (r *relationshipRoutes) write(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "relationships.write")
	defer span.End()

	request := new(WriteRequest)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	t := tuple.Tuple{Entity: request.Entity, Relation: tuple.Relation(request.Relation), Subject: request.Subject}
	err = r.relationshipService.WriteRelationship(ctx, t, request.SchemaVersion.String())
	if err != nil {
		if errors.Is(err, database.ErrUniqueConstraint) {
			span.RecordError(database.ErrUniqueConstraint)
			return c.JSON(http.StatusUnprocessableEntity, common.MResponse("tuple already exists"))
		}
		if errors.Is(err, tuple.NotFoundInSpecifiedRelationTypes) {
			span.RecordError(tuple.NotFoundInSpecifiedRelationTypes)
			return c.JSON(http.StatusUnprocessableEntity, common.MResponse(tuple.NotFoundInSpecifiedRelationTypes.Error()))
		}
		if errors.Is(err, internalErrors.EntityConfigCannotFoundError) {
			span.RecordError(internalErrors.EntityConfigCannotFoundError)
			return c.JSON(http.StatusUnprocessableEntity, common.MResponse(internalErrors.EntityConfigCannotFoundError.Error()))
		}
		span.SetStatus(codes.Error, echo.ErrInternalServerError.Error())
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, common.SuccessResponse(t))
}

// @Summary     Relationship
// @Description delete relation tuple
// @ID          relationships.delete
// @Tags  	    Relationship
// @Accept      json
// @Produce     json
// @Param       request body DeleteRequest true "delete relation tuple"
// @Success     200 {object} tuple.Tuple
// @Failure     400 {object} common.HTTPErrorResponse
// @Router      /relationships/delete [post]
func (r *relationshipRoutes) delete(c echo.Context) (err error) {
	ctx, span := tracer.Start(c.Request().Context(), "relationships.delete")
	defer span.End()

	request := new(DeleteRequest)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, common.ValidationResponse(v))
	}

	t := tuple.Tuple{Entity: request.Entity, Relation: tuple.Relation(request.Relation), Subject: request.Subject}
	err = r.relationshipService.DeleteRelationship(ctx, t)
	if err != nil {
		span.SetStatus(codes.Error, echo.ErrInternalServerError.Error())
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, common.SuccessResponse(t))
}
