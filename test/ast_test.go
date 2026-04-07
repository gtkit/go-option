package test_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestAST_ParseGenericStruct(t *testing.T) {
	src := `package test

type User[T any, R any] struct {
	Name            string
	NecGenericFiled T
	Age             int
	Gender          string
	GenericFiled    R
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	// Verify we can find the struct declaration
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		if ts.Name.Name != "User" {
			return true
		}
		found = true

		// Verify type params
		if ts.TypeParams == nil {
			t.Error("expected type params")
			return false
		}
		if len(ts.TypeParams.List) != 2 {
			t.Errorf("type params count = %d, want 2", len(ts.TypeParams.List))
			return false
		}

		// Verify struct fields
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			t.Error("expected struct type")
			return false
		}
		if len(st.Fields.List) != 5 {
			t.Errorf("fields count = %d, want 5", len(st.Fields.List))
		}
		return false
	})

	if !found {
		t.Error("User struct not found in AST")
	}
}

func TestAST_ParseStructTags(t *testing.T) {
	src := `package test

type Config struct {
	Name string ` + "`opt:\"-\"`" + `
	Age  int
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		st, ok := n.(*ast.StructType)
		if !ok {
			return true
		}
		if len(st.Fields.List) != 2 {
			t.Errorf("fields count = %d, want 2", len(st.Fields.List))
			return false
		}
		// First field should have tag
		if st.Fields.List[0].Tag == nil {
			t.Error("expected tag on first field")
		}
		// Second field should have no tag
		if st.Fields.List[1].Tag != nil {
			t.Error("expected no tag on second field")
		}
		return false
	})
}
