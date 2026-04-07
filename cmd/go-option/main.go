// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gtkit/go-option/options"
)

type enumFlag struct {
	value       string
	validValues []string
}

func (m *enumFlag) String() string {
	return m.value
}

func (m *enumFlag) Set(s string) error {
	for _, v := range m.validValues {
		if s == v {
			m.value = s
			return nil
		}
	}
	return fmt.Errorf("invalid value %q, valid values are: %v", s, m.validValues)
}

var (
	outputMode = enumFlag{
		value:       "write",
		validValues: []string{"write", "append"},
	}
	styleFlag = enumFlag{
		value:       "interface",
		validValues: []string{"interface", "closure"},
	}

	structTypeNameArg = flag.String("type", "", "Struct type name of the functional options struct.")
	outputArg         = flag.String("output", "", "Output file name, default: srcDir/opt_<struct type>_gen.go")
	withPrefix        string
)

func usage() {
	fmt.Fprintf(os.Stderr, "go-option is a tool for generating functional options pattern.\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "\tgo-option [flags]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	fmt.Fprintf(os.Stderr, "\t-type <struct name>\n")
	fmt.Fprintf(os.Stderr, "\t-output <output path>, default: srcDir/opt_xxx_gen.go\n")
	fmt.Fprintf(os.Stderr, "\t-prefix <With function prefix>, default: With{FieldName}\n")
	fmt.Fprintf(os.Stderr, "\t-mode <file writing mode>, default: write\n")
	fmt.Fprintf(os.Stderr, "\t\t- write: Overwrites or creates a new file.\n")
	fmt.Fprintf(os.Stderr, "\t\t- append: Adds to the end of the file.\n")
	fmt.Fprintf(os.Stderr, "\t-style <code generation style>, default: interface\n")
	fmt.Fprintf(os.Stderr, "\t\t- interface: Interface-based options pattern.\n")
	fmt.Fprintf(os.Stderr, "\t\t- closure: Closure-based options pattern (go-optioner compatible).\n")
}

func main() {
	flag.Var(&outputMode, "mode", "The file writing mode: write or append (default: write)")
	flag.Var(&styleFlag, "style", "Code generation style: interface or closure (default: interface)")
	flag.StringVar(&withPrefix, "prefix", "", "The prefix of the With{FieldName} function")
	flag.StringVar(&withPrefix, "with_prefix", "", "Alias for -prefix (go-optioner compatible)")
	flag.Usage = usage
	flag.Parse()

	if len(*structTypeNameArg) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	g := options.NewGenerator()
	g.StructInfo.StructName = *structTypeNameArg
	g.StructInfo.NewStructName = options.BigCamelToSmallCamel(*structTypeNameArg)
	g.SetOutPath(*outputArg)
	g.SetMode(outputMode.value)
	g.SetStyle(options.Style(styleFlag.value))
	g.SetWithPrefix(withPrefix)

	if err := g.GeneratingOptions(); err != nil {
		log.Fatalf("Error: %v", err)
	}
	if !g.Found {
		log.Fatalf("Target %q not found in current directory", g.StructInfo.StructName)
	}

	if err := g.GenerateCodeByTemplate(); err != nil {
		log.Fatalf("Error generating code: %v", err)
	}
	if err := g.OutputToFile(); err != nil {
		log.Fatalf("Error writing output: %v", err)
	}

	log.Printf("Generated functional options code successfully.\nOut: %s\n", g.OutPath())
}
