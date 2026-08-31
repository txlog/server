package util

import "strings"

// MaskString returns a fixed-length mask string to avoid revealing the length
// of sensitive values such as passwords or secrets. Returns an empty string
// for empty input.
func MaskString(theString string) string {
	if theString == "" {
		return ""
	}
	return "********"
}

// FormatSearchTerm prepares a search string for SQL LIKE queries by:
// 1. Adding '%' wildcards at the beginning and end of the search term
// 2. Converting any '*' characters to '%' wildcards
//
// Parameters:
//   - search: The original search string to be formatted
//
// Returns:
//
//	A formatted string ready for use in SQL LIKE clauses with proper wildcards
func FormatSearchTerm(search string) string {
	search = "%" + search + "%"
	search = strings.ReplaceAll(search, "*", "%")
	return search
}

// ContainsSpecialCharacters checks if a string contains any non-alphanumeric
// characters. It iterates through each rune in the input string and returns
// true if it finds any character that is not a letter or number. Returns false
// if all characters are alphanumeric.
//
// Parameters:
//   - s: The input string to check
//
// Returns:
//   - bool: true if special characters are found, false otherwise
func ContainsSpecialCharacters(s string) bool {
	for _, r := range s {
		if !isAlphanumeric(r) {
			return true
		}
	}
	return false
}

// isAlphanumeric checks if a given rune is an alphanumeric character. It
// returns true if the rune is either a digit (0-9), an uppercase letter (A-Z),
// or a lowercase letter (a-z). Returns false for any other character including
// spaces, symbols, or Unicode.
func isAlphanumeric(r rune) bool {
	return (r >= '0' && r <= '9') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z')
}
