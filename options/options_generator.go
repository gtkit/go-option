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
	"html/template"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/tools/imports"

	"github.com/gtkit/go-option/templates"
)

type Generator struct {
	StructInfo *StructInfo
	outPath    string
	mode       string

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

func (g *Generator) GeneratingOptions() {
	pkg, err := build.Default.ImportDir(".", 0)
	if err != nil {
		log.Fatalf("Processsing directory failed: %s", err.Error())
	}
	for _, file := range pkg.GoFiles {
		if found := g.parseStruct(file); found {
			g.Found = found
			break
		}
	}
}

func (g *Generator) parseStruct(fileName string) bool {
	fSet := token.NewFileSet()
	file, err := parser.ParseFile(fSet, fileName, nil, 0)
	if err != nil {
		log.Fatal(err)
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
			if structDecl, ok := typeSpec.Type.(*ast.StructType); ok {
				log.Printf("Generating Struct \"%s\" \n", g.StructInfo.StructName)
				if typeSpec.TypeParams != nil {
					log.Println("This is a struct which contains generic type:", typeSpec.Name)
					for _, param := range typeSpec.TypeParams.List {
						for _, name := range param.Names {
							typ := g.getTypeName(param.Type)
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
						if ident, ok := field.Type.(*ast.Ident); ok { // combined struct
							fieldName = ident.Name
						} else if starExpr, ok := field.Type.(*ast.StarExpr); ok {
							if ident2, ok := starExpr.X.(*ast.Ident); ok { // combined struct
								fieldName = ident2.Name
							} else {
								continue
							}
						} else {
							continue
						}
					} else {
						fieldName = field.Names[0].Name
					}
					optionIgnore := false

					fieldType := g.getTypeName(field.Type)
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
				return true
			} else {
				log.Fatalf(fmt.Sprintf("Target[%s] type is not a struct", g.StructInfo.StructName))
			}
		}
	}
	return false
}

func (g *Generator) GenerateCodeByTemplate() {
	var (
		headerTmpl *template.Template
		tmpl       *template.Template
		err        error
	)
	if g.mode == "append" {
		headerTmpl, err = template.New("header_options").Parse(templates.HeaderTmpl)
		if err != nil {
			log.Fatal("Failed to parse header template:", err)
		}
		err = headerTmpl.Execute(&g.header, g.StructInfo)
		if err != nil {
			log.Fatal(err)
		}
	}
	tmpl = template.New("options").Funcs(
		template.FuncMap{
			"bigCamelToSmallCamel":  BigCamelToSmallCamel,
			"capitalizeFirstLetter": CapitalizeFirstLetter,
			"getFirstLetter":        GetFirstLetter,
		})
	if g.mode == "write" {
		tmpl, err = tmpl.Parse(templates.OptionTmpl)
	} else {
		tmpl, err = tmpl.Parse(templates.AdditionalTmpl)
	}
	if err != nil {
		log.Fatal("Failed to parse template:", err)
	}

	err = tmpl.Execute(&g.code, g.StructInfo)
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Generator) OutputToFile() {
	var src []byte
	if g.mode == "write" {
		dir := filepath.Dir(g.outPath)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatal("mkdir failed, error: ", err)
		}
		src = g.forMart()
	} else {
		readFile, err := os.ReadFile(g.outPath)
		if err != nil {
			log.Fatal("Open the specified file failed, error: ", err)
		}
		header := g.header.Bytes()
		if !bytes.HasPrefix(readFile, header) {
			readFile = append(header, readFile...)
		}
		readFile = append(readFile, g.code.Bytes()...)
		src, err = imports.Process("", readFile, nil)
		if err != nil {
			log.Fatal("Failed to format the generated code: ", err, ", the grammar of the specified file code maybe incorrect.")
		}
	}
	err := os.WriteFile(g.outPath, src, 0644)
	if err != nil {
		log.Fatal("write file failed, error: ", err)
	}
	log.Printf("Generating Functional Options Code Successfully.\nOut: %s\n", g.outPath)
}

func (g *Generator) forMart() []byte {
	source, err := imports.Process("", g.code.Bytes(), nil)
	if err != nil {
		return nil
	}
	return source
}

func (g *Generator) SetOutPath(outPath *string) {
	fileName := fmt.Sprintf("opt_%s_gen.go", CamelToSnake(g.StructInfo.StructName))
	if len(*outPath) > 0 {
		g.outPath = *outPath
	} else {
		g.outPath = fileName
	}
}

func (g *Generator) getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", g.getTypeName(t.X), t.Sel.Name)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + g.getTypeName(t.Elt)
		}
		if basicLit, ok := t.Len.(*ast.BasicLit); ok && basicLit.Kind == token.INT {
			return "[" + basicLit.Value + "]" + g.getTypeName(t.Elt)
		} else {
			log.Fatalf("Array len error: %T", t)
			return ""
		}
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", g.getTypeName(t.Key), g.getTypeName(t.Value))
	case *ast.StarExpr:
		return "*" + g.getTypeName(t.X)
	// case *ast.InterfaceType:
	//	return "" // ignore
	case *ast.StructType:
		return "struct{}"
	case *ast.FuncType:
		return g.parseFuncType(t)
	case *ast.ChanType:
		return "chan " + g.getTypeName(t.Value)
	case *ast.UnaryExpr:
		return "~" + g.getTypeName(t.X)
	default:
		log.Fatalf("Unsupported type for field: %T", t)
		return ""
	}
}

func (g *Generator) parseFuncType(f *ast.FuncType) string {
	var params, results []string
	if f.Params != nil {
		for _, field := range f.Params.List {
			paramType := g.getTypeName(field.Type)
			for _, name := range field.Names {
				params = append(params, fmt.Sprintf("%s %s", name.Name, paramType))
			}
		}
	}

	if f.Results != nil {
		for _, field := range f.Results.List {
			resultType := g.getTypeName(field.Type)
			if len(field.Names) > 0 {
				for _, name := range field.Names {
					results = append(results, fmt.Sprintf("%s %s", name.Name, resultType))
				}
			} else {
				results = append(results, resultType)
			}
		}
	}

	if len(results) == 1 {
		return fmt.Sprintf("func(%s) %s", strings.Join(params, ", "), results[0])
	}
	return fmt.Sprintf("func(%s) (%s)", strings.Join(params, ", "), strings.Join(results, ", "))
}

func (g *Generator) SetMod(mode string) {
	g.mode = mode
}

func (g *Generator) SetWithPrefix(withPrefix string) {
	g.StructInfo.WithPrefix = withPrefix
}
