package options

import (
	"strings"
	"unicode"
)

// CamelToSnake converts camelCase or PascalCase to snake_case.
// Handles consecutive capitals correctly: HTTPServer → http_server, UserID → user_id.
func CamelToSnake(camelCase string) string {
	var result strings.Builder
	runes := []rune(camelCase)
	for i, c := range runes {
		if unicode.IsUpper(c) {
			if i > 0 {
				// Insert underscore before an uppercase letter when:
				// - Previous char is lowercase, OR
				// - Next char is lowercase (handles "HTTPServer" → "http_server")
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

// BigCamelToSmallCamel converts PascalCase to camelCase.
// UserAgent → userAgent.
func BigCamelToSmallCamel(bigCamel string) string {
	if len(bigCamel) == 0 {
		return ""
	}

	firstChar := strings.ToLower(string(bigCamel[0]))
	return firstChar + bigCamel[1:]
}

// CapitalizeFirstLetter converts the first letter to uppercase.
// user → User.
func CapitalizeFirstLetter(input string) string {
	if len(input) == 0 {
		return input
	}

	runes := []rune(input)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// GetFirstLetter returns the lowercase first letter of a string.
// User → u.
func GetFirstLetter(input string) string {
	if len(input) == 0 {
		return input
	}

	return strings.ToLower(string([]rune(input)[0]))
}
