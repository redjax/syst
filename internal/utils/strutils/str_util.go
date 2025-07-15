package strutils

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ToUpper returns s converted to uppercase.
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLower returns s converted to lowercase.
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToTitleCase returns the string with the first letter of each word capitalized.
// e.g. "hello world" → "Hello World"
func ToTitleCase(s string) string {
	// Create a Unicode-aware title caser
	caser := cases.Title(language.English)

	// Apply title casing to lowercase string
	return caser.String(strings.ToLower(s))
}

// Capitalize returns the string with only the first character uppercased.
// e.g. "hello world" → "Hello world"
func Capitalize(s string) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}

// RemoveSubstrings removes all instances of the substrings in the list from the input string.
func RemoveSubstrings(s string, removeList []string) string {
	for _, r := range removeList {
		s = strings.ReplaceAll(s, r, "")
	}
	return s
}

// ReplaceSubstrings applies search/replace substitutions in the form "search/replace"
func ReplaceSubstrings(s string, replaceList []string) (string, []string) {
	var warnings []string

	for _, r := range replaceList {
		parts := strings.SplitN(r, "/", 2)
		if len(parts) != 2 {
			warnings = append(warnings, r)
			continue
		}
		search := parts[0]
		replace := parts[1]
		s = strings.ReplaceAll(s, search, replace)
	}
	return s, warnings
}
