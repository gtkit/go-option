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
