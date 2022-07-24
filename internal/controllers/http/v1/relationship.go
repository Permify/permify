package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Permify/permify/internal/controllers/http/requests/relationship"
	"github.com/Permify/permify/internal/controllers/http/responses"
	"github.com/Permify/permify/internal/entities"
	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/services"
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
		h.POST("/write", r.write)
		h.POST("/delete", r.delete)
	}
}

// @Summary     Relationship
// @Description create new relation tuple
// @ID          write
// @Tags  	    Relationship
// @Accept      json
// @Produce     json
// @Param       request body relationship.Write true "''"
// @Success     200 {object} responses.Message
// @Failure     400 {object} responses.HTTPErrorResponse
// @Router      /relationships/write [post]
func (r *relationshipRoutes) write(c echo.Context) (err error) {
	request := new(relationship.Write)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request.Body); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(v))
	}

	err = r.relationshipService.WriteRelationship(context.Background(), []entities.RelationTuple{{Entity: request.Body.Entity, ObjectID: request.Body.ObjectID, Relation: request.Body.Relation, UsersetEntity: request.Body.UsersetEntity, UsersetObjectID: request.Body.UsersetObjectID, UsersetRelation: request.Body.UsersetRelation, Type: "custom"}})
	if err != nil {
		if errors.Is(err, repositories.ErrUniqueConstraint) {
			return c.JSON(http.StatusUnprocessableEntity, responses.MResponse("tuple already exists"))
		}
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, responses.MResponse("success"))
}

// @Summary     Relationship
// @Description delete relation tuple
// @ID          delete
// @Tags  	    Relationship
// @Accept      json
// @Produce     json
// @Param       request body relationship.Delete true "''"
// @Success     200 {object} responses.Message
// @Failure     400 {object} responses.HTTPErrorResponse
// @Router      /relationships/delete [post]
func (r *relationshipRoutes) delete(c echo.Context) (err error) {
	request := new(relationship.Delete)
	if err := (&echo.DefaultBinder{}).BindBody(c, &request.Body); err != nil {
		return err
	}
	v := request.Validate()
	if v != nil {
		return c.JSON(http.StatusUnprocessableEntity, responses.ValidationResponse(v))
	}

	err = r.relationshipService.DeleteRelationship(context.Background(), []entities.RelationTuple{{Entity: request.Body.Entity, ObjectID: request.Body.ObjectID, Relation: request.Body.Relation, UsersetEntity: request.Body.UsersetEntity, UsersetObjectID: request.Body.UsersetObjectID, UsersetRelation: request.Body.UsersetRelation}})
	if err != nil {
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, responses.MResponse("success"))
}
