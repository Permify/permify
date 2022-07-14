package parser

import (
	`strings`

	`github.com/Permify/permify/pkg/dsl/ast`
	`github.com/Permify/permify/pkg/dsl/schema`
)

// TranslateToSchema -
func TranslateToSchema(input string) (sch schema.Schema) {

	pr := NewParser(input)
	parsed := pr.Parse()

	var entities []schema.Entity

	for _, sc := range parsed.Statements {
		var entitySt = sc.(*ast.EntityStatement)
		var entity schema.Entity

		entity.Name = entitySt.Name.Literal

		if entitySt.Option.Literal != "" {
			entity.EntityOption = parseEntityOption(entitySt.Option.Literal)
		}

		if entity.EntityOption.Table == "" {
			entity.EntityOption.Table = entity.Name
		}

		if entity.EntityOption.Identifier == "" {
			entity.EntityOption.Identifier = "id"
		}

		// relations
		for _, rs := range entitySt.RelationStatements {
			var relationSt = rs.(*ast.RelationStatement)
			var relation schema.Relation
			relation.Name = relationSt.Name.Literal
			relation.Type = relationSt.Type.Literal

			if entitySt.Option.Literal != "" {
				relation.RelationOption = parseRelationOption(relationSt.Option.Literal)
			}

			if relation.RelationOption.Rel == "" {
				relation.RelationOption.Rel = schema.BelongsTo
			}

			if relation.RelationOption.Rel == schema.ManyToMany {
				if relation.RelationOption.Table == "" {
					relation.RelationOption.Table = relation.Name
				}
				if len(relation.RelationOption.Cols) < 2 {
					relation.RelationOption.Cols = append(relation.RelationOption.Cols, entity.Name+"_id", relation.Name+"_id")
				}
			}

			if relation.RelationOption.Rel == schema.BelongsTo {
				if len(relation.RelationOption.Cols) == 0 {
					relation.RelationOption.Cols = append(relation.RelationOption.Cols, relation.Name+"_id")
				}
			}

			entity.Relations = append(entity.Relations, relation)
		}

		// actions
		for _, as := range entitySt.ActionStatements {
			var st = as.(*ast.ActionStatement)
			var action schema.Action
			action.Name = st.Name.Literal
			action.Child = parseChild(st.ExpressionStatement.(*ast.ExpressionStatement))
			entity.Actions = append(entity.Actions, action)
		}

		entities = append(entities, entity)
	}

	return schema.NewSchema(entities...)
}

// parseChild -
func parseChild(expression *ast.ExpressionStatement) (re schema.Child) {
	return parseChildren(expression.Expression.(ast.Expression))
}

// parseChildren -
func parseChildren(expression ast.Expression) (children schema.Child) {
	if expression.IsInfix() {
		exp := expression.(*ast.InfixExpression)
		var child schema.Rewrite
		if exp.Operator == "or" {
			child.Type = schema.Union
		} else {
			child.Type = schema.Intersection
		}

		var ch []schema.Child
		ch = append(ch, parseChildren(exp.Left))
		ch = append(ch, parseChildren(exp.Right))

		child.Children = ch
		return child
	} else {
		exp := expression.(*ast.Identifier)
		var child schema.Leaf
		s := strings.Split(expression.String(), ".")
		if len(s) > 1 {
			child.Type = schema.TupleToUserSetType
			child.Value = exp.Value
		} else {
			child.Type = schema.ComputedUserSetType
			child.Value = exp.Value
		}
		return child
	}
}

// parseEntityOption -
func parseEntityOption(str string) (opt schema.EntityOption) {
	split := strings.Split(str, "|")
	for _, s := range split {
		spt := strings.Split(s, ":")
		if len(spt) < 2 {
			break
		}
		switch spt[0] {
		case "table":
			opt.Table = spt[1]
			break
		case "identifier":
			opt.Identifier = spt[1]
			break
		default:
			break
		}
	}
	return
}

// parseRelationOption -
func parseRelationOption(str string) (opt schema.RelationOption) {
	split := strings.Split(str, "|")

	for _, s := range split {
		spt := strings.Split(s, ":")
		if len(spt) < 2 {
			break
		}
		switch spt[0] {
		case "rel":
			opt.Rel = schema.RelationType(spt[1])
			break
		case "table":
			opt.Table = spt[1]
			break
		case "cols":
			opt.Cols = strings.Split(spt[1], ",")
			break
		default:
			break
		}
	}
	return
}
