package consumers

import (
	"context"
	"fmt"

	"github.com/Permify/permify/internal/services"
	"github.com/Permify/permify/pkg/dsl/schema"
	publisher "github.com/Permify/permify/pkg/publisher/postgres"
	"github.com/Permify/permify/pkg/tuple"
)

// Parser -
type Parser struct {
	relationshipService services.IRelationshipService
	schemaService       services.ISchemaService
}

// New -
func New(service services.IRelationshipService, schema services.ISchemaService) Parser {
	return Parser{
		relationshipService: service,
		schemaService:       schema,
	}
}

// GetService -
func (c *Parser) GetService() services.IRelationshipService {
	return c.relationshipService
}

// Parse -
func (c *Parser) Parse(notification publisher.Notification) (writeTuples []tuple.Tuple, deleteTuples []tuple.Tuple, err error) {
	switch notification.Action {
	case publisher.INSERT:
		writeTuples, err = c.Convert(notification.Entity, notification.NewData)
		break
	case publisher.UPDATE:
		deleteTuples, err = c.Convert(notification.Entity, notification.OldData)
		writeTuples, err = c.Convert(notification.Entity, notification.NewData)
		break
	case publisher.DELETE:
		deleteTuples, err = c.Convert(notification.Entity, notification.OldData)
		break
	default:
		break
	}
	return
}

// Convert -
func (c *Parser) Convert(table string, data map[string]interface{}) (tuples []tuple.Tuple, err error) {
	var rel schema.RelationType
	var entity schema.Entity
	var relations []schema.Relation

	var sch schema.Schema
	sch, err = c.schemaService.Schema(context.Background())
	if err != nil {
		return nil, err
	}

	switch sch.GetTableType(table) {
	case schema.Main:
		rel = schema.BelongsTo
		entity = sch.GetEntityByTableName(table)
		relations = schema.Relations(entity.Relations).Filter(schema.BelongsTo)
	case schema.Pivot:
		rel = schema.ManyToMany
		entity = sch.PivotToEntity[table]
		relations = schema.Relations{sch.PivotToRelation[table]}
	default:
		return nil, err
	}

	// Relations
	for _, relation := range relations {

		var object tuple.Object
		var objectKey string
		var userKey string

		switch rel {
		case schema.BelongsTo:
			objectKey = "id"
			userKey = relation.RelationOption.Cols[0]
		case schema.ManyToMany:
			objectKey = relation.RelationOption.Cols[0]
			userKey = relation.RelationOption.Cols[1]
		default:
			objectKey = "id"
			userKey = relation.RelationOption.Cols[0]
		}

		object = tuple.Object{
			Namespace: entity.Name,
			ID:        fmt.Sprintf("%v", data[objectKey]),
		}

		user := tuple.User{}

		if relation.Type == tuple.USER {
			user.ID = fmt.Sprintf("%v", data[userKey])
		} else {
			user.UserSet.Object = tuple.Object{
				Namespace: relation.Type,
				ID:        fmt.Sprintf("%v", data[userKey]),
			}
			user.UserSet.Relation = tuple.ELLIPSIS
		}

		tuples = append(tuples, tuple.Tuple{
			Object:   object,
			Relation: relation.Name,
			User:     user,
		})
	}

	return tuples, nil
}
