package options

import (
	"strings"
	"unicode"
)

// CamelToSnake 驼峰命名转成蛇形命名，如果不是驼峰命名，则转成对应小写字符串
// UserAgent → user_agent
// User → user.
func CamelToSnake(camelCase string) string {
	var result strings.Builder
	for i, c := range camelCase {
		if unicode.IsUpper(c) && i > 0 {
			result.WriteByte('_')
		}
		result.WriteRune(unicode.ToLower(c))
	}
	return result.String()
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

// CapitalizeFirstLetter 将首字母转换为大写
// user → User.
func CapitalizeFirstLetter(input string) string {
	if len(input) == 0 {
		return input
	}

	firstChar := []rune(input)[:1]
	firstCharUpper := string(unicode.ToUpper(firstChar[0]))

	return firstCharUpper + input[1:]
}
