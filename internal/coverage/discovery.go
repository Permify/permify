package coverage

import (
	"fmt"

	"github.com/Permify/permify/pkg/dsl/ast"
)

// Discover walks the AST and registers all logical nodes in the registry.
func Discover(sch *ast.Schema, r *Registry) {
	for _, st := range sch.Statements {
		switch v := st.(type) {
		case *ast.EntityStatement:
			discoverEntity(v, r)
		}
	}
}

func discoverEntity(es *ast.EntityStatement, r *Registry) {
	for _, ps := range es.PermissionStatements {
		st := ps.(*ast.PermissionStatement)
		path := fmt.Sprintf("%s#%s", es.Name.Literal, st.Name.Literal)

		// Register the root perm node
		r.Register(path, SourceInfo{
			Line:   int32(st.Name.PositionInfo.LinePosition),
			Column: int32(st.Name.PositionInfo.ColumnPosition),
		}, "PERMISSION")

		if st.ExpressionStatement != nil {
			expr := st.ExpressionStatement.(*ast.ExpressionStatement)
			discoverExpression(expr.Expression, path, r)
		}
	}
}

func discoverExpression(expr ast.Expression, path string, r *Registry) {
	if expr == nil {
		return
	}

	if expr.IsInfix() {
		node := expr.(*ast.InfixExpression)
		r.Register(path, SourceInfo{
			Line:   int32(node.Op.PositionInfo.LinePosition),
			Column: int32(node.Op.PositionInfo.ColumnPosition),
		}, string(node.Operator))

		discoverExpression(node.Left, AppendPath(path, "0"), r)
		discoverExpression(node.Right, AppendPath(path, "1"), r)
	} else {
		// Leaf node
		var info SourceInfo
		var nodeType string

		switch v := expr.(type) {
		case *ast.Identifier:
			if len(v.Idents) > 0 {
				info = SourceInfo{
					Line:   int32(v.Idents[0].PositionInfo.LinePosition),
					Column: int32(v.Idents[0].PositionInfo.ColumnPosition),
				}
			}
			nodeType = "LEAF"
		case *ast.Call:
			info = SourceInfo{
				Line:   int32(v.Name.PositionInfo.LinePosition),
				Column: int32(v.Name.PositionInfo.ColumnPosition),
			}
			nodeType = "CALL"
		}

		r.Register(path, info, nodeType)
	}
}
