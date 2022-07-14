package consumers

import (
	`fmt`

	`github.com/Permify/permify/internal/services`
	`github.com/Permify/permify/pkg/dsl/schema`
	`github.com/Permify/permify/pkg/tuple`
)

// Consumer -
type Consumer struct {
	service services.IRelationshipService
	schema  schema.Schema
}

// New -
func New(service services.IRelationshipService, schema schema.Schema) Consumer {
	return Consumer{
		service: service,
		schema:  schema,
	}
}

// GetService -
func (c *Consumer) GetService() services.IRelationshipService {
	return c.service
}

// GetStatement -
func (c *Consumer) GetStatement() schema.Schema {
	return c.schema
}

// Parse -
func (c *Consumer) Parse(notification Notification) (writeTuples []tuple.Tuple, deleteTuples []tuple.Tuple) {
	switch notification.Action {
	case INSERT:
		writeTuples = c.Convert(notification.Entity, notification.NewData)
		break
	case UPDATE:
		deleteTuples = c.Convert(notification.Entity, notification.OldData)
		writeTuples = c.Convert(notification.Entity, notification.NewData)
		break
	case DELETE:
		deleteTuples = c.Convert(notification.Entity, notification.OldData)
		break
	default:
		break
	}
	return
}

// Convert -
func (c *Consumer) Convert(table string, data map[string]interface{}) []tuple.Tuple {
	var tuples []tuple.Tuple

	var rel schema.RelationType
	var entity schema.Entity
	var relations []schema.Relation

	if c.schema.Tables[table] == schema.Main {
		rel = schema.BelongsTo
		entity = c.schema.Entities[table]
		relations = schema.Relations(entity.Relations).Filter(schema.BelongsTo)
	} else {
		rel = schema.ManyToMany
		entity = c.schema.PivotToEntity[table]
		relations = schema.Relations{c.schema.PivotToRelation[table]}
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

	return tuples
}
