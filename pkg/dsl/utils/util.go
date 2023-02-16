package utils

import (
	"strings"

	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/token"
)

// GetBaseEntityRelationTypeStatement - get entity reference statement from relation type statements
func GetBaseEntityRelationTypeStatement(sts []ast.RelationTypeStatement) ast.RelationTypeStatement {
	for _, st := range sts {
		if ast.IsDirectEntityReference(st) {
			return st
		}
	}
	return ast.RelationTypeStatement{
		Sign: token.Token{
			Type:    token.SIGN,
			Literal: "@",
		},
		Type: token.Token{
			Type:    token.IDENT,
			Literal: "user",
		},
	}
}

// Key -
func Key(v1, v2 string) string {
	var sb strings.Builder
	sb.WriteString(v1)
	sb.WriteString("#")
	sb.WriteString(v2)
	return sb.String()
}
