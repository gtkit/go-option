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
	"strings"

	"github.com/gtkit/go-option/options"
)

type ModeValue struct {
	value       string
	validValues []string
}

func (m *ModeValue) String() string {
	return m.value
}

func (m *ModeValue) Set(s string) error {
	for _, v := range m.validValues {
		if s == v {
			m.value = s
			return nil
		}
	}
	return fmt.Errorf("invalid value %q for mode, valid values are: %v", s, m.validValues)
}

var (
	outputMode = ModeValue{
		value:       "write",
		validValues: []string{"write", "append"},
	}

	structTypeNameArg = flag.String("type", "", "Struct type name of the functional options struct.")
	outputArg         = flag.String("output", "", "Output file name, default: srcDir/opt_<struct type>_gen.go")
	withPrefix        string
	g                 = options.NewGenerator()
)

func usage() {
	fmt.Fprintf(os.Stderr, "go-option is a tool for generating functional options pattern.\n")
	fmt.Fprintf(os.Stderr, "Usage: \n")
	fmt.Fprintf(os.Stderr, "\t go-option [flags]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	fmt.Fprintf(os.Stderr, "\t -type <struct name>\n")
	fmt.Fprintf(os.Stderr, "\t -output <output path>, default: srcDir/opt_xxx_gen.go\n")
	fmt.Fprintf(os.Stderr, "\t -with_prefix <the prefix of the With{filed_name} function>, default is With{filed_name}.If specified, such as User, it will generate WithUser{filed_name}\n")
	fmt.Fprintf(os.Stderr, "\t -mode <the file writing mode>, default: write\n")
	fmt.Fprintf(os.Stderr, "\t there are two available modes:\n")
	fmt.Fprintf(os.Stderr, "\t\t - write(Write/Overwrite): Overwrites or creates a new file.\n")
	fmt.Fprintf(os.Stderr, "\t\t - append (Append): Adds to the end of the file.\n")
}

func main() {
	flag.Var(&outputMode, "mode", "The file writing mode, default: write")
	flag.StringVar(&withPrefix, "with_prefix", "", "The prefix of the With{filed_name} function, default is With{filed_name}.If specified, such as User, it will generate WithUser{filed_name}")
	flag.Usage = usage
	flag.Parse()
	if len(*structTypeNameArg) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	g.StructInfo.StructName = *structTypeNameArg
	g.StructInfo.NewStructName = BigCamelToSmallCamel(*structTypeNameArg)
	g.SetOutPath(outputArg)
	g.SetMod(outputMode.value)
	g.SetWithPrefix(withPrefix)

	g.GeneratingOptions()
	if !g.Found {
		log.Printf("Target \"[%s]\" is not be found\n", g.StructInfo.StructName)
		os.Exit(1)
	}

	g.GenerateCodeByTemplate()
	g.OutputToFile()
}

// BigCamelToSmallCamel 大驼峰格式的字符串转小驼峰格式的字符串
// UserAgent → userAgent.
func BigCamelToSmallCamel(bigCamel string) string {
	if len(bigCamel) == 0 {
		return ""
	}

	firstChar := strings.ToLower(string(bigCamel[0]))
	return firstChar + bigCamel[1:]
}
