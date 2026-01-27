package coverage

import (
	"testing"

	"github.com/Permify/permify/pkg/dsl/ast"
	"github.com/Permify/permify/pkg/dsl/token"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	info1 := SourceInfo{Line: 1, Column: 1}
	info2 := SourceInfo{Line: 2, Column: 5}

	r.Register("path1", info1, "OR")
	r.Register("path2", info2, "AND")

	r.Visit("path1")

	uncovered := r.Report()

	if len(uncovered) != 1 {
		t.Errorf("expected 1 uncovered node, got %d", len(uncovered))
	}

	if uncovered[0].Path != "path2" {
		t.Errorf("expected path2 to be uncovered, got %s", uncovered[0].Path)
	}

	r.Visit("path2")
	uncovered = r.Report()
	if len(uncovered) != 0 {
		t.Errorf("expected 0 uncovered nodes, got %d", len(uncovered))
	}
}

func TestDiscover(t *testing.T) {
	sch := &ast.Schema{
		Statements: []ast.Statement{
			&ast.EntityStatement{
				Name: token.Token{Literal: "repository"},
				PermissionStatements: []ast.Statement{
					&ast.PermissionStatement{
						Name: token.Token{Literal: "edit", PositionInfo: token.PositionInfo{LinePosition: 1, ColumnPosition: 12}},
						ExpressionStatement: &ast.ExpressionStatement{
							Expression: &ast.InfixExpression{
								Op:       token.Token{Literal: "or", PositionInfo: token.PositionInfo{LinePosition: 1, ColumnPosition: 20}},
								Operator: ast.OR,
								Left: &ast.Identifier{
									Idents: []token.Token{
										{Literal: "owner", PositionInfo: token.PositionInfo{LinePosition: 1, ColumnPosition: 15}},
									},
								},
								Right: &ast.Identifier{
									Idents: []token.Token{
										{Literal: "admin", PositionInfo: token.PositionInfo{LinePosition: 1, ColumnPosition: 25}},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	r := NewRegistry()
	Discover(sch, r)

	report := r.Report()
	if len(report) != 3 {
		t.Errorf("expected 3 nodes (PERMISSION, OR, LEAF), got %d", len(report))
	}

	// Verify paths
	foundEdit := false
	foundEdit0 := false
	foundEdit1 := false

	for _, node := range report {
		switch node.Path {
		case "repository#edit":
			foundEdit = true
			if node.Type != "PERMISSION" {
				t.Errorf("expected PERMISSION type for repository#edit, got %s", node.Type)
			}
		case "repository#edit.0":
			foundEdit0 = true
			if node.Type != "LEAF" {
				t.Errorf("expected LEAF type for repository#edit.0, got %s", node.Type)
			}
		case "repository#edit.1":
			foundEdit1 = true
			if node.Type != "LEAF" {
				t.Errorf("expected LEAF type for repository#edit.1, got %s", node.Type)
			}
		}
	}

	if !foundEdit || !foundEdit0 || !foundEdit1 {
		t.Errorf("missing paths: edit:%v, edit.0:%v, edit.1:%v", foundEdit, foundEdit0, foundEdit1)
	}
}
