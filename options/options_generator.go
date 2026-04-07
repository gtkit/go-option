// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package options

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"

	"github.com/gtkit/go-option/templates"
)

// Style represents the code generation style.
type Style string

const (
	// StyleInterface generates interface-based options (default).
	StyleInterface Style = "interface"
	// StyleClosure generates closure-based options (go-optioner compatible).
	StyleClosure Style = "closure"
)

type Generator struct {
	StructInfo *StructInfo
	outPath    string
	mode       string
	style      Style

	code   bytes.Buffer
	header bytes.Buffer
	Found  bool
}

func NewGenerator() *Generator {
	return &Generator{
		StructInfo: &StructInfo{
			Fields:         make([]FieldInfo, 0),
			OptionalFields: make([]FieldInfo, 0),
		},
		style: StyleInterface,
	}
}

type FieldInfo struct {
	Name string
	Type string
}

type StructInfo struct {
	PackageName    string
	StructName     string
	NewStructName  string
	Fields         []FieldInfo
	OptionalFields []FieldInfo
	GenericParams  []FieldInfo

	Imports []string

	WithPrefix string
}

// GeneratingOptions parses all Go files in the current directory to find the target struct.
func (g *Generator) GeneratingOptions() error {
	pkg, err := build.Default.ImportDir(".", 0)
	if err != nil {
		return fmt.Errorf("processing directory failed: %w", err)
	}
	for _, file := range pkg.GoFiles {
		found, err := g.parseStruct(file)
		if err != nil {
			return fmt.Errorf("parsing file %s: %w", file, err)
		}
		if found {
			g.Found = true
			break
		}
	}
	return nil
}

func (g *Generator) parseStruct(fileName string) (bool, error) {
	fSet := token.NewFileSet()
	file, err := parser.ParseFile(fSet, fileName, nil, 0)
	if err != nil {
		return false, fmt.Errorf("parsing Go file %s: %w", fileName, err)
	}

	g.StructInfo.PackageName = file.Name.Name

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if typeSpec.Name.String() != g.StructInfo.StructName {
				continue
			}
			structDecl, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return false, fmt.Errorf("target %q is not a struct type", g.StructInfo.StructName)
			}

			if typeSpec.TypeParams != nil {
				for _, param := range typeSpec.TypeParams.List {
					for _, name := range param.Names {
						typ, err := g.getTypeName(param.Type)
						if err != nil {
							return false, fmt.Errorf("resolving generic param type: %w", err)
						}
						g.StructInfo.GenericParams = append(g.StructInfo.GenericParams, FieldInfo{
							Name: name.Name,
							Type: typ,
						})
					}
				}
			}
			for _, field := range structDecl.Fields.List {
				var fieldName string
				if len(field.Names) == 0 {
					switch t := field.Type.(type) {
					case *ast.Ident:
						fieldName = t.Name
					case *ast.StarExpr:
						if ident, ok := t.X.(*ast.Ident); ok {
							fieldName = ident.Name
						} else {
							continue
						}
					default:
						continue
					}
				} else {
					fieldName = field.Names[0].Name
				}

				fieldType, err := g.getTypeName(field.Type)
				if err != nil {
					return false, fmt.Errorf("resolving type of field %s: %w", fieldName, err)
				}

				optionIgnore := false
				if field.Tag != nil {
					tags := strings.Replace(field.Tag.Value, "`", "", -1)
					tag := reflect.StructTag(tags).Get("opt")
					if tag == "-" {
						g.StructInfo.Fields = append(g.StructInfo.Fields, FieldInfo{
							Name: fieldName,
							Type: fieldType,
						})
						optionIgnore = true
					}
				}
				if !optionIgnore {
					g.StructInfo.OptionalFields = append(g.StructInfo.OptionalFields, FieldInfo{
						Name: fieldName,
						Type: fieldType,
					})
				}
			}
			return true, nil
		}
	}
	return false, nil
}

// GenerateCodeByTemplate renders the option code using the appropriate template.
func (g *Generator) GenerateCodeByTemplate() error {
	if g.mode == "append" {
		headerTmpl, err := template.New("header_options").Parse(templates.HeaderTmpl)
		if err != nil {
			return fmt.Errorf("parsing header template: %w", err)
		}
		if err = headerTmpl.Execute(&g.header, g.StructInfo); err != nil {
			return fmt.Errorf("executing header template: %w", err)
		}
	}

	tmpl := template.New("options").Funcs(
		template.FuncMap{
			"bigCamelToSmallCamel":  BigCamelToSmallCamel,
			"capitalizeFirstLetter": CapitalizeFirstLetter,
			"getFirstLetter":        GetFirstLetter,
		})

	tmplContent, err := g.selectTemplate()
	if err != nil {
		return err
	}

	tmpl, err = tmpl.Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	if err = tmpl.Execute(&g.code, g.StructInfo); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}
	return nil
}

// selectTemplate returns the appropriate template content based on style and mode.
func (g *Generator) selectTemplate() (string, error) {
	switch g.style {
	case StyleClosure:
		if g.mode == "append" {
			return templates.AdditionalClosureTmpl, nil
		}
		return templates.OptionClosureTmpl, nil
	case StyleInterface:
		if g.mode == "append" {
			return templates.AdditionalTmpl, nil
		}
		return templates.OptionTmpl, nil
	default:
		return "", fmt.Errorf("unsupported style: %s", g.style)
	}
}

// OutputToFile writes the generated code to the output file.
func (g *Generator) OutputToFile() error {
	var src []byte
	if g.mode == "write" {
		dir := filepath.Dir(g.outPath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
		var err error
		src, err = g.format()
		if err != nil {
			return fmt.Errorf("formatting generated code: %w", err)
		}
	} else {
		readFile, err := os.ReadFile(g.outPath)
		if err != nil {
			return fmt.Errorf("reading file %s: %w", g.outPath, err)
		}
		header := g.header.Bytes()
		if !bytes.HasPrefix(readFile, header) {
			readFile = append(header, readFile...)
		}
		readFile = append(readFile, g.code.Bytes()...)
		src, err = imports.Process("", readFile, nil)
		if err != nil {
			return fmt.Errorf("formatting appended code (target file may have syntax errors): %w", err)
		}
	}
	if err := os.WriteFile(g.outPath, src, 0644); err != nil {
		return fmt.Errorf("writing file %s: %w", g.outPath, err)
	}
	return nil
}

func (g *Generator) format() ([]byte, error) {
	source, err := imports.Process("", g.code.Bytes(), nil)
	if err != nil {
		return nil, fmt.Errorf("imports.Process failed: %w", err)
	}
	return source, nil
}

// SetOutPath sets the output file path. If outPath is empty, uses the default naming convention.
func (g *Generator) SetOutPath(outPath string) {
	if len(outPath) > 0 {
		g.outPath = outPath
	} else {
		g.outPath = fmt.Sprintf("opt_%s_gen.go", CamelToSnake(g.StructInfo.StructName))
	}
}

func (g *Generator) getTypeName(expr ast.Expr) (string, error) {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name, nil
	case *ast.SelectorExpr:
		x, err := g.getTypeName(t.X)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.%s", x, t.Sel.Name), nil
	case *ast.ArrayType:
		elt, err := g.getTypeName(t.Elt)
		if err != nil {
			return "", err
		}
		if t.Len == nil {
			return "[]" + elt, nil
		}
		if basicLit, ok := t.Len.(*ast.BasicLit); ok && basicLit.Kind == token.INT {
			return "[" + basicLit.Value + "]" + elt, nil
		}
		return "", fmt.Errorf("unsupported array length expression: %T", t.Len)
	case *ast.MapType:
		key, err := g.getTypeName(t.Key)
		if err != nil {
			return "", err
		}
		val, err := g.getTypeName(t.Value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("map[%s]%s", key, val), nil
	case *ast.StarExpr:
		x, err := g.getTypeName(t.X)
		if err != nil {
			return "", err
		}
		return "*" + x, nil
	case *ast.StructType:
		return "struct{}", nil
	case *ast.FuncType:
		return g.parseFuncType(t)
	case *ast.ChanType:
		val, err := g.getTypeName(t.Value)
		if err != nil {
			return "", err
		}
		switch t.Dir {
		case ast.SEND:
			return "chan<- " + val, nil
		case ast.RECV:
			return "<-chan " + val, nil
		default:
			return "chan " + val, nil
		}
	case *ast.InterfaceType:
		return "interface{}", nil
	case *ast.UnaryExpr:
		x, err := g.getTypeName(t.X)
		if err != nil {
			return "", err
		}
		return "~" + x, nil
	case *ast.BinaryExpr:
		left, err := g.getTypeName(t.X)
		if err != nil {
			return "", err
		}
		right, err := g.getTypeName(t.Y)
		if err != nil {
			return "", err
		}
		return left + " | " + right, nil
	case *ast.Ellipsis:
		elt, err := g.getTypeName(t.Elt)
		if err != nil {
			return "", err
		}
		return "..." + elt, nil
	default:
		return "", fmt.Errorf("unsupported AST type: %T", expr)
	}
}

func (g *Generator) parseFuncType(f *ast.FuncType) (string, error) {
	var params, results []string
	if f.Params != nil {
		for _, field := range f.Params.List {
			paramType, err := g.getTypeName(field.Type)
			if err != nil {
				return "", err
			}
			if len(field.Names) > 0 {
				for _, name := range field.Names {
					params = append(params, fmt.Sprintf("%s %s", name.Name, paramType))
				}
			} else {
				params = append(params, paramType)
			}
		}
	}

	if f.Results != nil {
		for _, field := range f.Results.List {
			resultType, err := g.getTypeName(field.Type)
			if err != nil {
				return "", err
			}
			if len(field.Names) > 0 {
				for _, name := range field.Names {
					results = append(results, fmt.Sprintf("%s %s", name.Name, resultType))
				}
			} else {
				results = append(results, resultType)
			}
		}
	}

	if len(results) == 0 {
		return fmt.Sprintf("func(%s)", strings.Join(params, ", ")), nil
	}
	if len(results) == 1 {
		return fmt.Sprintf("func(%s) %s", strings.Join(params, ", "), results[0]), nil
	}
	return fmt.Sprintf("func(%s) (%s)", strings.Join(params, ", "), strings.Join(results, ", ")), nil
}

// SetMode sets the file writing mode ("write" or "append").
func (g *Generator) SetMode(mode string) {
	g.mode = mode
}

// SetStyle sets the code generation style ("interface" or "closure").
func (g *Generator) SetStyle(style Style) {
	g.style = style
}

func (g *Generator) SetWithPrefix(withPrefix string) {
	g.StructInfo.WithPrefix = withPrefix
}

// OutPath returns the configured output file path.
func (g *Generator) OutPath() string {
	return g.outPath
}
