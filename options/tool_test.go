package options

import "testing"

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"User", "user"},
		{"UserAgent", "user_agent"},
		{"HTTPServer", "http_server"},
		{"UserID", "user_id"},
		{"ID", "id"},
		{"A", "a"},
		{"", ""},
		{"already_snake", "already_snake"},
		{"lowercase", "lowercase"},
		{"TLSConfig", "tls_config"},
		{"getHTTPSURL", "get_httpsurl"},
		{"XMLParser", "xml_parser"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := CamelToSnake(tt.input)
			if got != tt.want {
				t.Errorf("CamelToSnake(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBigCamelToSmallCamel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"UserAgent", "userAgent"},
		{"User", "user"},
		{"A", "a"},
		{"", ""},
		{"already", "already"},
		{"HTTPServer", "hTTPServer"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := BigCamelToSmallCamel(tt.input)
			if got != tt.want {
				t.Errorf("BigCamelToSmallCamel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCapitalizeFirstLetter(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"user", "User"},
		{"User", "User"},
		{"a", "A"},
		{"", ""},
		{"123abc", "123abc"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := CapitalizeFirstLetter(tt.input)
			if got != tt.want {
				t.Errorf("CapitalizeFirstLetter(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSafeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"name", "name"},
		{"func", "func_"},
		{"map", "map_"},
		{"chan", "chan_"},
		{"type", "type_"},
		{"var", "var_"},
		{"struct", "struct_"},
		{"interface", "interface_"},
		{"select", "select_"},
		{"range", "range_"},
		{"return", "return_"},
		{"for", "for_"},
		{"go", "go_"},
		{"if", "if_"},
		{"switch", "switch_"},
		{"case", "case_"},
		{"break", "break_"},
		{"continue", "continue_"},
		{"defer", "defer_"},
		{"default", "default_"},
		{"else", "else_"},
		{"fallthrough", "fallthrough_"},
		{"goto", "goto_"},
		{"import", "import_"},
		{"package", "package_"},
		{"const", "const_"},
		// Non-keywords should pass through
		{"port", "port"},
		{"value", "value"},
		{"timeout", "timeout"},
		{"config", "config"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SafeName(tt.input)
			if got != tt.want {
				t.Errorf("SafeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetFirstLetter(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"User", "u"},
		{"user", "u"},
		{"A", "a"},
		{"", ""},
		{"123", "1"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := GetFirstLetter(tt.input)
			if got != tt.want {
				t.Errorf("GetFirstLetter(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
