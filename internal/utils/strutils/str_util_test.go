package strutils

import "testing"

func TestToUpper(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", "HELLO"},
		{"Hello World", "HELLO WORLD"},
		{"", ""},
		{"ALREADY", "ALREADY"},
		{"hello123", "HELLO123"},
	}
	for _, tt := range tests {
		if got := ToUpper(tt.input); got != tt.want {
			t.Errorf("ToUpper(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"HELLO", "hello"},
		{"Hello World", "hello world"},
		{"", ""},
		{"already", "already"},
	}
	for _, tt := range tests {
		if got := ToLower(tt.input); got != tt.want {
			t.Errorf("ToLower(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestToTitleCase(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello world", "Hello World"},
		{"HELLO WORLD", "Hello World"},
		{"already Title", "Already Title"},
		{"", ""},
		{"single", "Single"},
	}
	for _, tt := range tests {
		if got := ToTitleCase(tt.input); got != tt.want {
			t.Errorf("ToTitleCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello world", "Hello world"},
		{"already", "Already"},
		{"", ""},
		{"a", "A"},
		{"HELLO", "HELLO"},
	}
	for _, tt := range tests {
		if got := Capitalize(tt.input); got != tt.want {
			t.Errorf("Capitalize(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRemoveSubstrings(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		removeList []string
		ignoreCase bool
		want       string
	}{
		{"basic removal", "hello world", []string{"world"}, false, "hello "},
		{"multiple removals", "foo bar baz", []string{"foo", "baz"}, false, " bar "},
		{"case sensitive miss", "Hello World", []string{"hello"}, false, "Hello World"},
		{"case insensitive hit", "Hello World", []string{"hello"}, true, " World"},
		{"empty remove list", "hello", []string{}, false, "hello"},
		{"empty string", "", []string{"a"}, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveSubstrings(tt.input, tt.removeList, tt.ignoreCase)
			if got != tt.want {
				t.Errorf("RemoveSubstrings(%q, %v, %v) = %q, want %q", tt.input, tt.removeList, tt.ignoreCase, got, tt.want)
			}
		})
	}
}

func TestReplaceSubstrings(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		replaceList []string
		ignoreCase  bool
		want        string
		wantWarns   int
	}{
		{"basic replace", "hello world", []string{"world/earth"}, false, "hello earth", 0},
		{"case insensitive", "Hello World", []string{"hello/hi"}, true, "hi World", 0},
		{"invalid entry", "hello", []string{"noslash"}, false, "hello", 1},
		{"multiple replaces", "aaa bbb", []string{"aaa/xxx", "bbb/yyy"}, false, "xxx yyy", 0},
		{"empty input", "", []string{"a/b"}, false, "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, warnings := ReplaceSubstrings(tt.input, tt.replaceList, tt.ignoreCase)
			if got != tt.want {
				t.Errorf("ReplaceSubstrings result = %q, want %q", got, tt.want)
			}
			if len(warnings) != tt.wantWarns {
				t.Errorf("ReplaceSubstrings warnings count = %d, want %d", len(warnings), tt.wantWarns)
			}
		})
	}
}

func TestSearchSubstring(t *testing.T) {
	tests := []struct {
		s, substr  string
		ignoreCase bool
		want       bool
	}{
		{"hello world", "world", false, true},
		{"hello world", "WORLD", false, false},
		{"hello world", "WORLD", true, true},
		{"hello world", "xyz", false, false},
		{"", "hello", false, false},
		{"hello", "", false, true},
	}
	for _, tt := range tests {
		got := SearchSubstring(tt.s, tt.substr, tt.ignoreCase)
		if got != tt.want {
			t.Errorf("SearchSubstring(%q, %q, %v) = %v, want %v", tt.s, tt.substr, tt.ignoreCase, got, tt.want)
		}
	}
}

func TestFindMatchingLines(t *testing.T) {
	input := "hello world\nfoo bar\nhello again\ngoodbye"

	tests := []struct {
		name       string
		substr     string
		ignoreCase bool
		wantCount  int
	}{
		{"exact match", "hello", false, 2},
		{"case insensitive", "HELLO", true, 2},
		{"case sensitive miss", "HELLO", false, 0},
		{"single match", "foo", false, 1},
		{"no match", "xyz", false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindMatchingLines(input, tt.substr, tt.ignoreCase)
			if len(got) != tt.wantCount {
				t.Errorf("FindMatchingLines(_, %q, %v) count = %d, want %d", tt.substr, tt.ignoreCase, len(got), tt.wantCount)
			}
		})
	}
}
