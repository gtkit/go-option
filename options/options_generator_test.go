package options

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratingOptions_SimpleStruct(t *testing.T) {
	// Change to test fixture directory
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	if err := os.Chdir(filepath.Join("testdata", "simple")); err != nil {
		t.Fatal(err)
	}

	g := NewGenerator()
	g.StructInfo.StructName = "User"
	g.StructInfo.NewStructName = BigCamelToSmallCamel("User")

	err = g.GeneratingOptions()
	if err != nil {
		t.Fatalf("GeneratingOptions() error: %v", err)
	}
	if !g.Found {
		t.Fatal("expected to find User struct")
	}
	if g.StructInfo.PackageName != "simple" {
		t.Errorf("PackageName = %q, want %q", g.StructInfo.PackageName, "simple")
	}
	// Name should be a required field (opt:"-")
	if len(g.StructInfo.Fields) != 1 {
		t.Fatalf("Fields count = %d, want 1", len(g.StructInfo.Fields))
	}
	if g.StructInfo.Fields[0].Name != "Name" {
		t.Errorf("Fields[0].Name = %q, want %q", g.StructInfo.Fields[0].Name, "Name")
	}
	// Age and Bio should be optional
	if len(g.StructInfo.OptionalFields) != 2 {
		t.Fatalf("OptionalFields count = %d, want 2", len(g.StructInfo.OptionalFields))
	}
}

func TestGeneratingOptions_GenericStruct(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	if err := os.Chdir(filepath.Join("testdata", "generic")); err != nil {
		t.Fatal(err)
	}

	g := NewGenerator()
	g.StructInfo.StructName = "Container"
	g.StructInfo.NewStructName = BigCamelToSmallCamel("Container")

	err = g.GeneratingOptions()
	if err != nil {
		t.Fatalf("GeneratingOptions() error: %v", err)
	}
	if !g.Found {
		t.Fatal("expected to find Container struct")
	}
	if len(g.StructInfo.GenericParams) != 2 {
		t.Fatalf("GenericParams count = %d, want 2", len(g.StructInfo.GenericParams))
	}
	if g.StructInfo.GenericParams[0].Name != "T" || g.StructInfo.GenericParams[0].Type != "any" {
		t.Errorf("GenericParams[0] = %+v, want {T any}", g.StructInfo.GenericParams[0])
	}
	if g.StructInfo.GenericParams[1].Name != "U" || g.StructInfo.GenericParams[1].Type != "comparable" {
		t.Errorf("GenericParams[1] = %+v, want {U comparable}", g.StructInfo.GenericParams[1])
	}
}

func TestGeneratingOptions_ComplexStruct(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	if err := os.Chdir(filepath.Join("testdata", "complex")); err != nil {
		t.Fatal(err)
	}

	g := NewGenerator()
	g.StructInfo.StructName = "Server"
	g.StructInfo.NewStructName = BigCamelToSmallCamel("Server")

	err = g.GeneratingOptions()
	if err != nil {
		t.Fatalf("GeneratingOptions() error: %v", err)
	}
	if !g.Found {
		t.Fatal("expected to find Server struct")
	}
	// Addr and Port are required
	if len(g.StructInfo.Fields) != 2 {
		t.Fatalf("Fields count = %d, want 2", len(g.StructInfo.Fields))
	}
	// Check optional fields include embedded struct, various types
	optNames := make([]string, len(g.StructInfo.OptionalFields))
	for i, f := range g.StructInfo.OptionalFields {
		optNames[i] = f.Name
	}
	expected := []string{"Timeout", "MaxConns", "TLSEnabled", "Tags", "Metadata", "Handler", "Embedded"}
	if len(optNames) != len(expected) {
		t.Fatalf("OptionalFields names = %v, want %v", optNames, expected)
	}
	for i, name := range expected {
		if optNames[i] != name {
			t.Errorf("OptionalFields[%d].Name = %q, want %q", i, optNames[i], name)
		}
	}
}

func TestGeneratingOptions_NotFound(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	if err := os.Chdir(filepath.Join("testdata", "simple")); err != nil {
		t.Fatal(err)
	}

	g := NewGenerator()
	g.StructInfo.StructName = "NonExistentStruct"
	g.StructInfo.NewStructName = BigCamelToSmallCamel("NonExistentStruct")

	err = g.GeneratingOptions()
	if err != nil {
		t.Fatalf("GeneratingOptions() unexpected error: %v", err)
	}
	if g.Found {
		t.Error("expected Found to be false for non-existent struct")
	}
}

func TestGenerateCodeByTemplate_InterfaceStyle(t *testing.T) {
	g := NewGenerator()
	g.StructInfo.PackageName = "test"
	g.StructInfo.StructName = "Config"
	g.StructInfo.NewStructName = "config"
	g.StructInfo.Fields = []FieldInfo{{Name: "Name", Type: "string"}}
	g.StructInfo.OptionalFields = []FieldInfo{{Name: "Port", Type: "int"}}
	g.SetMode("write")
	g.SetStyle(StyleInterface)

	err := g.GenerateCodeByTemplate()
	if err != nil {
		t.Fatalf("GenerateCodeByTemplate() error: %v", err)
	}

	code := g.code.String()
	// Verify interface-based pattern
	if !strings.Contains(code, "ConfigOption interface") {
		t.Error("expected interface definition in generated code")
	}
	if !strings.Contains(code, "func NewConfig(") {
		t.Error("expected NewConfig constructor")
	}
	if !strings.Contains(code, "opt.apply(config)") {
		t.Error("expected interface apply call")
	}
	if !strings.Contains(code, "WithPort") {
		t.Error("expected WithPort function")
	}
}

func TestGenerateCodeByTemplate_ClosureStyle(t *testing.T) {
	g := NewGenerator()
	g.StructInfo.PackageName = "test"
	g.StructInfo.StructName = "Config"
	g.StructInfo.NewStructName = "config"
	g.StructInfo.Fields = []FieldInfo{{Name: "Name", Type: "string"}}
	g.StructInfo.OptionalFields = []FieldInfo{{Name: "Port", Type: "int"}}
	g.SetMode("write")
	g.SetStyle(StyleClosure)

	err := g.GenerateCodeByTemplate()
	if err != nil {
		t.Fatalf("GenerateCodeByTemplate() error: %v", err)
	}

	code := g.code.String()
	// Verify closure-based pattern
	if !strings.Contains(code, "ConfigOption") {
		t.Error("expected ConfigOption type in generated code")
	}
	if !strings.Contains(code, "func(*Config)") {
		t.Error("expected closure type definition")
	}
	if !strings.Contains(code, "opt(config)") {
		t.Error("expected closure-style call")
	}
	if !strings.Contains(code, "WithPort") {
		t.Error("expected WithPort function")
	}
	if !strings.Contains(code, "return func(") {
		t.Error("expected closure return pattern")
	}
}

func TestGenerateCodeByTemplate_WithPrefix(t *testing.T) {
	g := NewGenerator()
	g.StructInfo.PackageName = "test"
	g.StructInfo.StructName = "Config"
	g.StructInfo.NewStructName = "config"
	g.StructInfo.OptionalFields = []FieldInfo{{Name: "Port", Type: "int"}}
	g.StructInfo.WithPrefix = "Server"
	g.SetMode("write")
	g.SetStyle(StyleInterface)

	err := g.GenerateCodeByTemplate()
	if err != nil {
		t.Fatalf("GenerateCodeByTemplate() error: %v", err)
	}

	code := g.code.String()
	if !strings.Contains(code, "WithServerPort") {
		t.Error("expected WithServerPort function with prefix")
	}
}

func TestGenerateCodeByTemplate_GenericClosure(t *testing.T) {
	g := NewGenerator()
	g.StructInfo.PackageName = "test"
	g.StructInfo.StructName = "Container"
	g.StructInfo.NewStructName = "container"
	g.StructInfo.Fields = []FieldInfo{{Name: "ID", Type: "string"}}
	g.StructInfo.OptionalFields = []FieldInfo{{Name: "Value", Type: "T"}}
	g.StructInfo.GenericParams = []FieldInfo{{Name: "T", Type: "any"}}
	g.SetMode("write")
	g.SetStyle(StyleClosure)

	err := g.GenerateCodeByTemplate()
	if err != nil {
		t.Fatalf("GenerateCodeByTemplate() error: %v", err)
	}

	code := g.code.String()
	if !strings.Contains(code, "ContainerOption[T any]") {
		t.Error("expected generic type in closure option type")
	}
	if !strings.Contains(code, "func NewContainer[T any]") {
		t.Error("expected generic constructor")
	}
}

func TestOutputToFile_WriteMode(t *testing.T) {
	g := NewGenerator()
	g.StructInfo.PackageName = "test"
	g.StructInfo.StructName = "Config"
	g.StructInfo.NewStructName = "config"
	g.StructInfo.Fields = []FieldInfo{{Name: "Name", Type: "string"}}
	g.StructInfo.OptionalFields = []FieldInfo{{Name: "Port", Type: "int"}}
	g.SetMode("write")
	g.SetStyle(StyleInterface)

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "opt_config_gen.go")
	g.SetOutPath(outPath)

	if err := g.GenerateCodeByTemplate(); err != nil {
		t.Fatal(err)
	}
	if err := g.OutputToFile(); err != nil {
		t.Fatalf("OutputToFile() error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "package test") {
		t.Error("expected package declaration in output")
	}
	if !strings.Contains(content, "ConfigOption") {
		t.Error("expected ConfigOption in output")
	}
}

func TestOutputToFile_ClosureWriteMode(t *testing.T) {
	g := NewGenerator()
	g.StructInfo.PackageName = "test"
	g.StructInfo.StructName = "Config"
	g.StructInfo.NewStructName = "config"
	g.StructInfo.Fields = []FieldInfo{{Name: "Name", Type: "string"}}
	g.StructInfo.OptionalFields = []FieldInfo{{Name: "Port", Type: "int"}}
	g.SetMode("write")
	g.SetStyle(StyleClosure)

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "opt_config_gen.go")
	g.SetOutPath(outPath)

	if err := g.GenerateCodeByTemplate(); err != nil {
		t.Fatal(err)
	}
	if err := g.OutputToFile(); err != nil {
		t.Fatalf("OutputToFile() error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "func(*Config)") {
		t.Error("expected closure type in output")
	}
}

func TestSetOutPath(t *testing.T) {
	g := NewGenerator()
	g.StructInfo.StructName = "UserConfig"

	// Default path
	g.SetOutPath("")
	if g.outPath != "opt_user_config_gen.go" {
		t.Errorf("default outPath = %q, want %q", g.outPath, "opt_user_config_gen.go")
	}

	// Custom path
	g.SetOutPath("/tmp/custom.go")
	if g.outPath != "/tmp/custom.go" {
		t.Errorf("custom outPath = %q, want %q", g.outPath, "/tmp/custom.go")
	}
}

func TestSetMode(t *testing.T) {
	g := NewGenerator()
	g.SetMode("append")
	if g.mode != "append" {
		t.Errorf("mode = %q, want %q", g.mode, "append")
	}
}

func TestSetStyle(t *testing.T) {
	g := NewGenerator()
	g.SetStyle(StyleClosure)
	if g.style != StyleClosure {
		t.Errorf("style = %q, want %q", g.style, StyleClosure)
	}
}

func TestGetTypeName_AllTypes(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	if err := os.Chdir(filepath.Join("testdata", "complex")); err != nil {
		t.Fatal(err)
	}

	g := NewGenerator()
	g.StructInfo.StructName = "Server"
	g.StructInfo.NewStructName = BigCamelToSmallCamel("Server")

	err = g.GeneratingOptions()
	if err != nil {
		t.Fatalf("GeneratingOptions() error: %v", err)
	}

	// Verify type resolution for various field types
	typeMap := make(map[string]string)
	for _, f := range g.StructInfo.Fields {
		typeMap[f.Name] = f.Type
	}
	for _, f := range g.StructInfo.OptionalFields {
		typeMap[f.Name] = f.Type
	}

	checks := map[string]string{
		"Addr":       "string",
		"Port":       "int",
		"Timeout":    "time.Duration",
		"Tags":       "[]string",
		"Metadata":   "map[string]string",
		"Handler":    "func() error",
		"TLSEnabled": "bool",
	}
	for name, wantType := range checks {
		if got, ok := typeMap[name]; !ok {
			t.Errorf("field %q not found", name)
		} else if got != wantType {
			t.Errorf("field %q type = %q, want %q", name, got, wantType)
		}
	}
}

func TestGetTypeName_AllTypeVariants(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	if err := os.Chdir(filepath.Join("testdata", "alltypes")); err != nil {
		t.Fatal(err)
	}

	g := NewGenerator()
	g.StructInfo.StructName = "AllTypes"
	g.StructInfo.NewStructName = BigCamelToSmallCamel("AllTypes")

	err = g.GeneratingOptions()
	if err != nil {
		t.Fatalf("GeneratingOptions() error: %v", err)
	}
	if !g.Found {
		t.Fatal("expected to find AllTypes struct")
	}

	typeMap := make(map[string]string)
	for _, f := range g.StructInfo.Fields {
		typeMap[f.Name] = f.Type
	}
	for _, f := range g.StructInfo.OptionalFields {
		typeMap[f.Name] = f.Type
	}

	checks := map[string]string{
		"Name":        "string",
		"Age":         "int",
		"Ptr":         "*string",
		"Slice":       "[]int",
		"Array":       "[3]byte",
		"Map":         "map[string]int",
		"Chan":        "chan int",
		"SendChan":    "chan<- int",
		"RecvChan":    "<-chan int",
		"Func":        "func()",
		"FuncArgs":    "func(a int, b string)",
		"FuncReturn":  "func() error",
		"FuncMultiRet": "func(x int) (int, error)",
		"EmptyStruct": "struct{}",
		"Iface":       "interface{}",
		"NamedIface":  "io.Reader",
		"Embedded":    "Embedded",
		"EmbeddedPtr": "*Embedded",
		"Variadic":    "func(...int)",
	}
	for name, wantType := range checks {
		got, ok := typeMap[name]
		if !ok {
			t.Errorf("field %q not found", name)
			continue
		}
		if got != wantType {
			t.Errorf("field %q type = %q, want %q", name, got, wantType)
		}
	}
}

func TestOutputToFile_AppendMode(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "opt_config_gen.go")

	// First, create the base file
	baseContent := `// Generated by [go-option] command-line tool; DO NOT EDIT
// If you have any questions, please create issues and submit contributions at:
// https://github.com/gtkit/go-option
package test
`
	if err := os.WriteFile(outPath, []byte(baseContent), 0644); err != nil {
		t.Fatal(err)
	}

	g := NewGenerator()
	g.StructInfo.PackageName = "test"
	g.StructInfo.StructName = "Config"
	g.StructInfo.NewStructName = "config"
	g.StructInfo.OptionalFields = []FieldInfo{{Name: "Port", Type: "int"}}
	g.SetMode("append")
	g.SetStyle(StyleInterface)
	g.SetOutPath(outPath)

	if err := g.GenerateCodeByTemplate(); err != nil {
		t.Fatal(err)
	}
	if err := g.OutputToFile(); err != nil {
		t.Fatalf("OutputToFile() append error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "ConfigOption") {
		t.Error("expected ConfigOption in appended output")
	}
	if !strings.Contains(content, "package test") {
		t.Error("expected package declaration preserved")
	}
}

func TestOutputToFile_AppendClosureMode(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "opt_config_gen.go")

	baseContent := `// Generated by [go-option] command-line tool; DO NOT EDIT
// If you have any questions, please create issues and submit contributions at:
// https://github.com/gtkit/go-option
package test
`
	if err := os.WriteFile(outPath, []byte(baseContent), 0644); err != nil {
		t.Fatal(err)
	}

	g := NewGenerator()
	g.StructInfo.PackageName = "test"
	g.StructInfo.StructName = "Config"
	g.StructInfo.NewStructName = "config"
	g.StructInfo.OptionalFields = []FieldInfo{{Name: "Port", Type: "int"}}
	g.SetMode("append")
	g.SetStyle(StyleClosure)
	g.SetOutPath(outPath)

	if err := g.GenerateCodeByTemplate(); err != nil {
		t.Fatal(err)
	}
	if err := g.OutputToFile(); err != nil {
		t.Fatalf("OutputToFile() append closure error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "func(*Config)") {
		t.Error("expected closure type in appended output")
	}
}

func TestGenerateAndOutput_EndToEnd_InterfaceGeneric(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	if err := os.Chdir(filepath.Join("testdata", "generic")); err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "opt_container_gen.go")

	g := NewGenerator()
	g.StructInfo.StructName = "Container"
	g.StructInfo.NewStructName = BigCamelToSmallCamel("Container")
	g.SetOutPath(outPath)
	g.SetMode("write")
	g.SetStyle(StyleInterface)

	if err := g.GeneratingOptions(); err != nil {
		t.Fatal(err)
	}
	if !g.Found {
		t.Fatal("Container not found")
	}
	if err := g.GenerateCodeByTemplate(); err != nil {
		t.Fatal(err)
	}
	if err := g.OutputToFile(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "ContainerOption[T any, U comparable]") {
		t.Error("expected generic ContainerOption interface")
	}
	if !strings.Contains(content, "func NewContainer[T any, U comparable]") {
		t.Error("expected generic constructor")
	}
}

func TestGenerateAndOutput_EndToEnd_ClosureGeneric(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	if err := os.Chdir(filepath.Join("testdata", "generic")); err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "opt_container_gen.go")

	g := NewGenerator()
	g.StructInfo.StructName = "Container"
	g.StructInfo.NewStructName = BigCamelToSmallCamel("Container")
	g.SetOutPath(outPath)
	g.SetMode("write")
	g.SetStyle(StyleClosure)

	if err := g.GeneratingOptions(); err != nil {
		t.Fatal(err)
	}
	if !g.Found {
		t.Fatal("Container not found")
	}
	if err := g.GenerateCodeByTemplate(); err != nil {
		t.Fatal(err)
	}
	if err := g.OutputToFile(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "ContainerOption[T any, U comparable]") {
		t.Error("expected generic ContainerOption closure type")
	}
	if !strings.Contains(content, "func(*Container[T, U])") {
		t.Error("expected closure function type")
	}
}

func TestWithPrefixAndOutPath(t *testing.T) {
	g := NewGenerator()
	g.SetWithPrefix("Custom")
	if g.StructInfo.WithPrefix != "Custom" {
		t.Errorf("WithPrefix = %q, want %q", g.StructInfo.WithPrefix, "Custom")
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "my_output.go")
	g.StructInfo.StructName = "Test"
	g.SetOutPath(outPath)
	if g.OutPath() != outPath {
		t.Errorf("OutPath() = %q, want %q", g.OutPath(), outPath)
	}
}

func TestGetTypeName_UnionConstraint(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	if err := os.Chdir(filepath.Join("testdata", "union")); err != nil {
		t.Fatal(err)
	}

	g := NewGenerator()
	g.StructInfo.StructName = "Calc"
	g.StructInfo.NewStructName = BigCamelToSmallCamel("Calc")

	err = g.GeneratingOptions()
	if err != nil {
		t.Fatalf("GeneratingOptions() error: %v", err)
	}
	if !g.Found {
		t.Fatal("expected to find Calc struct")
	}
	if len(g.StructInfo.GenericParams) != 1 {
		t.Fatalf("GenericParams count = %d, want 1", len(g.StructInfo.GenericParams))
	}
	// Number is an interface defined in the same file, should resolve to its name
	if g.StructInfo.GenericParams[0].Type != "Number" {
		t.Errorf("GenericParam type = %q, want %q", g.StructInfo.GenericParams[0].Type, "Number")
	}
}

func TestGenerateCodeByTemplate_AppendMode(t *testing.T) {
	g := NewGenerator()
	g.StructInfo.PackageName = "test"
	g.StructInfo.StructName = "Config"
	g.StructInfo.NewStructName = "config"
	g.StructInfo.OptionalFields = []FieldInfo{{Name: "Port", Type: "int"}}
	g.SetMode("append")
	g.SetStyle(StyleInterface)

	err := g.GenerateCodeByTemplate()
	if err != nil {
		t.Fatalf("GenerateCodeByTemplate() append error: %v", err)
	}

	code := g.code.String()
	if !strings.Contains(code, "ConfigOption") {
		t.Error("expected ConfigOption in append code")
	}
	// Header should be generated for append mode
	header := g.header.String()
	if len(header) == 0 {
		t.Error("expected header to be generated in append mode")
	}
}

func TestGenerateCodeByTemplate_AppendClosureMode(t *testing.T) {
	g := NewGenerator()
	g.StructInfo.PackageName = "test"
	g.StructInfo.StructName = "Config"
	g.StructInfo.NewStructName = "config"
	g.StructInfo.OptionalFields = []FieldInfo{{Name: "Port", Type: "int"}}
	g.SetMode("append")
	g.SetStyle(StyleClosure)

	err := g.GenerateCodeByTemplate()
	if err != nil {
		t.Fatalf("GenerateCodeByTemplate() error: %v", err)
	}

	code := g.code.String()
	if !strings.Contains(code, "func(*Config)") {
		t.Error("expected closure type in append code")
	}
}

func TestOutputToFile_WriteMode_SubDir(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "sub", "dir", "opt_config_gen.go")

	g := NewGenerator()
	g.StructInfo.PackageName = "test"
	g.StructInfo.StructName = "Config"
	g.StructInfo.NewStructName = "config"
	g.StructInfo.OptionalFields = []FieldInfo{{Name: "Port", Type: "int"}}
	g.SetMode("write")
	g.SetStyle(StyleInterface)
	g.SetOutPath(outPath)

	if err := g.GenerateCodeByTemplate(); err != nil {
		t.Fatal(err)
	}
	if err := g.OutputToFile(); err != nil {
		t.Fatalf("OutputToFile() error creating subdirs: %v", err)
	}

	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("expected output file to be created in subdirectory")
	}
}

func TestOutputToFile_AppendMode_FileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "nonexistent.go")

	g := NewGenerator()
	g.StructInfo.PackageName = "test"
	g.StructInfo.StructName = "Config"
	g.StructInfo.NewStructName = "config"
	g.StructInfo.OptionalFields = []FieldInfo{{Name: "Port", Type: "int"}}
	g.SetMode("append")
	g.SetStyle(StyleInterface)
	g.SetOutPath(outPath)

	if err := g.GenerateCodeByTemplate(); err != nil {
		t.Fatal(err)
	}
	err := g.OutputToFile()
	if err == nil {
		t.Error("expected error when appending to non-existent file")
	}
}

func TestGeneratingOptions_InvalidDir(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	g := NewGenerator()
	g.StructInfo.StructName = "Test"
	g.StructInfo.NewStructName = "test"

	err = g.GeneratingOptions()
	if err == nil {
		t.Error("expected error for directory with no Go files")
	}
}

func TestSelectTemplate(t *testing.T) {
	tests := []struct {
		style   Style
		mode    string
		wantErr bool
	}{
		{StyleInterface, "write", false},
		{StyleInterface, "append", false},
		{StyleClosure, "write", false},
		{StyleClosure, "append", false},
		{Style("invalid"), "write", true},
	}
	for _, tt := range tests {
		t.Run(string(tt.style)+"_"+tt.mode, func(t *testing.T) {
			g := NewGenerator()
			g.SetStyle(tt.style)
			g.SetMode(tt.mode)
			_, err := g.selectTemplate()
			if (err != nil) != tt.wantErr {
				t.Errorf("selectTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
