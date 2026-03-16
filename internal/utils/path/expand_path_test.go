package path

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not get home dir: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"tilde only", "~", home, false},
		{"tilde with subpath", "~/Documents", filepath.Join(home, "Documents"), false},
		{"tilde nested", "~/a/b/c", filepath.Join(home, "a/b/c"), false},
		{"absolute path unchanged", "/tmp/test", "/tmp/test", false},
		{"relative path unchanged", "relative/path", "relative/path", false},
		{"empty path", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandPath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandPath(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
