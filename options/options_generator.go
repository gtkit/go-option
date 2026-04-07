// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package options 提供函数选项模式（Functional Options Pattern）的代码生成核心逻辑。
// 通过解析 Go 源文件中的结构体定义，自动生成对应的 Option 类型、构造函数和 With 函数。
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

// Style 表示代码生成风格。
type Style string

const (
	// StyleInterface 生成基于接口的选项模式（默认）。
	// 每个可选字段生成一个实现了 apply 方法的 unexported 结构体。
	StyleInterface Style = "interface"

	// StyleClosure 生成基于闭包的选项模式（兼容 go-optioner）。
	// Option 类型为 func(*Struct)，每个 With 函数返回一个闭包。
	StyleClosure Style = "closure"
)

// Generator 是代码生成器的核心结构体。
// 负责解析源文件、提取结构体信息、渲染模板并输出生成的代码文件。
type Generator struct {
	// StructInfo 包含目标结构体的元数据信息。
	StructInfo *StructInfo

	outPath string // 输出文件路径
	mode    string // 文件写入模式："write"（覆盖）或 "append"（追加）
	style   Style  // 代码生成风格

	code   bytes.Buffer // 生成的代码缓冲区
	header bytes.Buffer // 文件头注释缓冲区（仅 append 模式使用）

	// Found 标记目标结构体是否在当前目录中找到。
	Found bool
}

// NewGenerator 创建一个新的代码生成器实例。
// 默认使用 interface 风格生成代码。
func NewGenerator() *Generator {
	return &Generator{
		StructInfo: &StructInfo{
			Fields:         make([]FieldInfo, 0),
			OptionalFields: make([]FieldInfo, 0),
		},
		style: StyleInterface,
	}
}

// FieldInfo 表示结构体字段的名称和类型信息。
type FieldInfo struct {
	Name string // 字段名称，如 "Username"
	Type string // 字段类型字符串，如 "string"、"*int"、"map[string]int"
}

// StructInfo 包含目标结构体的全部元数据，用于模板渲染。
type StructInfo struct {
	PackageName    string      // 包名，如 "example"
	StructName     string      // 结构体名称（大驼峰），如 "UserConfig"
	NewStructName  string      // 结构体名称（小驼峰），如 "userConfig"，用作构造函数中的变量名
	Fields         []FieldInfo // 必填字段列表（标记了 opt:"-" 的字段），作为构造函数的必传参数
	OptionalFields []FieldInfo // 可选字段列表（未标记的字段），会生成对应的 With 函数
	GenericParams  []FieldInfo // 泛型类型参数列表，如 [{Name:"T", Type:"any"}, {Name:"U", Type:"comparable"}]
	WithPrefix     string      // With 函数的自定义前缀，如设置为 "User" 则生成 WithUserName 而非 WithName
}

// GeneratingOptions 解析当前目录下的所有 Go 源文件，查找目标结构体。
// 找到后提取结构体的字段信息、泛型参数等元数据，存储到 g.StructInfo 中。
// 如果成功找到目标结构体，g.Found 会被设置为 true。
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

// parseStruct 解析单个 Go 源文件，查找并提取目标结构体的信息。
// 使用 Go 标准库的 AST 解析器遍历语法树，识别结构体定义及其字段。
// 通过 opt:"-" 标签区分必填字段和可选字段。
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

			// 解析泛型类型参数（Go 1.18+）
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

			// 遍历结构体字段
			for _, field := range structDecl.Fields.List {
				var fieldName string
				if len(field.Names) == 0 {
					// 匿名嵌入字段：从类型表达式中提取名称
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

				// 通过 opt:"-" 标签判断字段是否为必填
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

// GenerateCodeByTemplate 使用 Go 模板引擎渲染生成代码。
// 根据当前的 style（interface/closure）和 mode（write/append）选择对应的模板，
// 将 StructInfo 中的元数据填充到模板中，生成最终的 Go 源代码。
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
			"safeName":              SafeName,
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

// selectTemplate 根据 style 和 mode 的组合选择对应的模板内容。
//
//	style=interface + mode=write  → OptionTmpl（完整文件，接口模式）
//	style=interface + mode=append → AdditionalTmpl（追加片段，接口模式）
//	style=closure   + mode=write  → OptionClosureTmpl（完整文件，闭包模式）
//	style=closure   + mode=append → AdditionalClosureTmpl（追加片段，闭包模式）
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

// OutputToFile 将生成的代码写入输出文件。
//   - write 模式：格式化代码后覆盖写入（自动创建目录）。
//   - append 模式：读取现有文件，追加生成的代码，再整体格式化。
//
// 使用 golang.org/x/tools/imports 自动管理导入和格式化。
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
		src, err = imports.Process(g.outPath, readFile, nil)
		if err != nil {
			return fmt.Errorf("formatting appended code (target file may have syntax errors): %w", err)
		}
	}
	if err := os.WriteFile(g.outPath, src, 0644); err != nil {
		return fmt.Errorf("writing file %s: %w", g.outPath, err)
	}
	return nil
}

// format 使用 goimports 对生成的代码进行格式化和导入管理。
// 传入 outPath 使 goimports 能正确解析本地包的导入路径。
func (g *Generator) format() ([]byte, error) {
	source, err := imports.Process(g.outPath, g.code.Bytes(), nil)
	if err != nil {
		return nil, fmt.Errorf("imports.Process failed: %w", err)
	}
	return source, nil
}

// SetOutPath 设置输出文件路径。
// 如果 outPath 为空，使用默认命名约定：opt_<snake_case_struct_name>_gen.go。
// 例如：结构体 UserConfig → opt_user_config_gen.go。
func (g *Generator) SetOutPath(outPath string) {
	if len(outPath) > 0 {
		g.outPath = outPath
	} else {
		g.outPath = fmt.Sprintf("opt_%s_gen.go", CamelToSnake(g.StructInfo.StructName))
	}
}

// getTypeName 递归地将 AST 类型表达式转换为 Go 类型字符串。
// 支持以下所有 Go 类型表达式：
//   - 基本类型：string, int, bool 等
//   - 指针：*Type
//   - 切片：[]Type，数组：[N]Type
//   - Map：map[K]V
//   - 通道：chan T, chan<- T, <-chan T
//   - 函数：func(params) (results)
//   - 接口：interface{}
//   - 结构体：struct{}
//   - 包限定类型：pkg.Type
//   - 泛型实例化：Type[T], Type[T, U]
//   - 联合约束：~int | ~float64
//   - 可变参数：...Type
//   - 括号表达式：(Type)
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
	case *ast.IndexExpr:
		base, err := g.getTypeName(t.X)
		if err != nil {
			return "", err
		}
		idx, err := g.getTypeName(t.Index)
		if err != nil {
			return "", err
		}
		return base + "[" + idx + "]", nil
	case *ast.IndexListExpr:
		base, err := g.getTypeName(t.X)
		if err != nil {
			return "", err
		}
		var indices []string
		for _, index := range t.Indices {
			idx, err := g.getTypeName(index)
			if err != nil {
				return "", err
			}
			indices = append(indices, idx)
		}
		return base + "[" + strings.Join(indices, ", ") + "]", nil
	case *ast.ParenExpr:
		return g.getTypeName(t.X)
	default:
		return "", fmt.Errorf("unsupported AST type: %T", expr)
	}
}

// parseFuncType 将函数类型 AST 节点转换为字符串表示。
// 正确处理有名参数、无名参数、多返回值等情况。
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

// SetMode 设置文件写入模式。
//   - "write"：覆盖或创建新文件（默认）。
//   - "append"：追加到已有文件末尾。
func (g *Generator) SetMode(mode string) {
	g.mode = mode
}

// SetStyle 设置代码生成风格。
//   - StyleInterface（"interface"）：基于接口的选项模式（默认）。
//   - StyleClosure（"closure"）：基于闭包的选项模式。
func (g *Generator) SetStyle(style Style) {
	g.style = style
}

// SetWithPrefix 设置 With 函数的自定义前缀。
// 例如设置为 "User"，字段 Name 将生成 WithUserName 而非 WithName。
func (g *Generator) SetWithPrefix(withPrefix string) {
	g.StructInfo.WithPrefix = withPrefix
}

// OutPath 返回当前配置的输出文件路径。
func (g *Generator) OutPath() string {
	return g.outPath
}
