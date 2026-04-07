package options

import (
	"strings"
	"unicode"
)

// CamelToSnake 将驼峰命名（PascalCase/camelCase）转换为蛇形命名（snake_case）。
// 正确处理连续大写字母：HTTPServer → http_server，UserID → user_id。
func CamelToSnake(camelCase string) string {
	var result strings.Builder
	runes := []rune(camelCase)
	for i, c := range runes {
		if unicode.IsUpper(c) {
			if i > 0 {
				prev := runes[i-1]
				if unicode.IsLower(prev) {
					result.WriteByte('_')
				} else if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					result.WriteByte('_')
				}
			}
			result.WriteRune(unicode.ToLower(c))
		} else {
			result.WriteRune(c)
		}
	}
	return result.String()
}

// BigCamelToSmallCamel 将大驼峰（PascalCase）转换为小驼峰（camelCase）。
// 示例：UserAgent → userAgent，HTTPServer → hTTPServer。
// 使用 rune 安全处理，支持多字节 UTF-8 字符。
func BigCamelToSmallCamel(bigCamel string) string {
	runes := []rune(bigCamel)
	if len(runes) == 0 {
		return ""
	}

	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// CapitalizeFirstLetter 将字符串首字母转为大写。
// 示例：user → User，config → Config。
func CapitalizeFirstLetter(input string) string {
	if len(input) == 0 {
		return input
	}

	runes := []rune(input)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// GetFirstLetter 获取字符串的小写首字母。
// 示例：User → u，Config → c。
// 用于生成接收者变量名。
func GetFirstLetter(input string) string {
	runes := []rune(input)
	if len(runes) == 0 {
		return input
	}

	return string(unicode.ToLower(runes[0]))
}

// goKeywords 包含所有 Go 语言保留关键字。
// 用于 SafeName 检测生成的标识符是否与关键字冲突。
var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}

// SafeName 确保标识符不与 Go 关键字冲突。
// 如果 name 是 Go 保留关键字（如 func、map、type 等），自动追加下划线后缀。
// 示例：func → func_，map → map_，port → port（不变）。
func SafeName(name string) string {
	if goKeywords[name] {
		return name + "_"
	}
	return name
}
