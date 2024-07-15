package test_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"testing"
)

type User[T any, R any] struct {
	Name            string
	NecGenericFiled T
	Age             int
	Gender          string
	GenericFiled    R
}

func TestAST(t *testing.T) {
	src := `package test

import (
	"testing"
)

type User[T any, R any] struct {
Name string 
NecGenericFiled T 
Age int
Gender string
GenericFiled R
}
`
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		panic(err)
	}

	// Print the AST.
	ast.Print(fset, f)
}

func TestDirectory(t *testing.T) {
	str, _ := os.Getwd()
	t.Log("Current directory is:", str)
}
