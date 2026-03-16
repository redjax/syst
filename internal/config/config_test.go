package config

import "testing"

func TestParserForFile(t *testing.T) {
	tests := []struct {
		path    string
		wantErr bool
	}{
		{"config.yaml", false},
		{"config.yml", false},
		{"config.json", false},
		{"config.toml", false},
		{"config.env", false},
		{"config.YAML", false}, // case insensitive
		{"config.JSON", false},
		{"config.xml", true},  // unsupported
		{"config.ini", true},  // unsupported
		{"noextension", true}, // no extension
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			parser, err := parserForFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parserForFile(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
			if !tt.wantErr && parser == nil {
				t.Errorf("parserForFile(%q) returned nil parser", tt.path)
			}
		})
	}
}
