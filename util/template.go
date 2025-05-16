package util

import (
	"html/template"
	"strconv"
	"strings"

	"github.com/txlog/server/version"
)

// Text2HTML converts a plain text string to HTML by replacing newlines with <br> tags
// and spaces with non-breaking space entities (&nbsp;).
// The function is useful for preserving text formatting when displaying in HTML context.
// WARNING: This function returns template.HTML which bypasses HTML escaping.
// Only use this function with trusted input to avoid XSS vulnerabilities.
//
// Parameters:
//   - s: The input string to convert
//
// Returns:
//   - template.HTML: The converted HTML-safe string
func Text2HTML(s string) template.HTML {
	s = strings.ReplaceAll(s, "\n", "<br>")
	s = strings.ReplaceAll(s, " ", "&nbsp;")
	return template.HTML(s)
}

// FormatPercentage formats a float64 percentage value into a string using Brazilian number formatting.
// It converts decimal points to commas and adds thousand separators using dots.
//
// For example:
//
//	FormatPercentage(1234.56) returns "1.234,56"
//	FormatPercentage(-1234.56) returns "-1.234,56"
//	FormatPercentage(123.45) returns "123,45"
//
// Parameters:
//   - percentage: The float64 value to be formatted
//
// Returns:
//
//	A string containing the formatted percentage value using Brazilian number format
func FormatPercentage(percentage float64) string {
	s := strconv.FormatFloat(percentage, 'f', 2, 64)
	s = strings.ReplaceAll(s, ".", ",")

	parts := strings.Split(s, ",")
	integerPart := parts[0]
	decimalPart := parts[1]

	isNegative := strings.HasPrefix(integerPart, "-")
	if isNegative {
		integerPart = integerPart[1:]
	}

	n := len(integerPart)
	if n <= 3 {
		if isNegative {
			return "-" + integerPart + "," + decimalPart
		}
		return integerPart + "," + decimalPart
	}

	var result string
	for i := 0; i < n; i++ {
		if (n-i)%3 == 0 && i != 0 {
			result += "."
		}
		result += string(integerPart[i])
	}
	if isNegative {
		return "-" + result + "," + decimalPart
	}
	return result + "," + decimalPart
}

// FormatInteger formats an integer with thousand separators using dots.
// It handles both positive and negative numbers.
//
// The function converts the integer to a string and adds a dot (.) as a
// thousand separator every three digits from right to left. If the number
// is negative, the minus sign is preserved at the beginning.
//
// Examples:
//
//	FormatInteger(1234)    returns "1.234"
//	FormatInteger(-1234)   returns "-1.234"
//	FormatInteger(1000000) returns "1.000.000"
//
// Parameters:
//   - num: The integer number to format
//
// Returns:
//
//	A string representation of the number with thousand separators
func FormatInteger(num int) string {
	s := strconv.Itoa(num)
	isNegative := strings.HasPrefix(s, "-")
	if isNegative {
		s = s[1:]
	}

	n := len(s)
	if n <= 3 {
		if isNegative {
			return "-" + s
		}
		return s
	}

	var result string
	for i := 0; i < n; i++ {
		if (n-i)%3 == 0 && i != 0 {
			result += "."
		}
		result += string(s[i])
	}
	if isNegative {
		return "-" + result
	}
	return result
}

// Iterate generates a slice of integers from start to count (inclusive).
// It returns a slice containing all integers in the range [start, count].
//
// Parameters:
//   - start: The first number in the sequence
//   - count: The last number in the sequence
//
// Returns:
//   - []int: A slice containing all integers from start to count
func Iterate(start, count int) []int {
	var items []int
	for i := start; i <= count; i++ {
		items = append(items, i)
	}
	return items
}

// Add returns the sum of two integers a and b.
func Add(a, b int) int {
	return a + b
}

// Min returns the minimum value between two integers.
// It compares two integers a and b and returns the smaller one.
func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

// Version returns the semantic version of the application.
// The version information is defined in the version package.
func Version() string {
	return version.SemVer
}

// DnfUser extracts the username from a DNF-style user string format.
// It processes strings in the format "display name <username>" and returns the username portion.
//
// The function handles three cases:
// 1. Empty string: returns "Unknown"
// 2. String with angle brackets (<>): extracts and returns content between brackets
// 3. String without angle brackets: returns the original string unchanged
//
// Parameters:
//   - user: A string that may contain a DNF-style user format
//
// Returns:
//   - A string containing either the extracted username, "Unknown", or the unchanged input
//
// Example:
//
//	DnfUser("rodrigo avila <rodrigo.avila>") // returns "rodrigo.avila"
//	DnfUser("") // returns "Unknown"
//	DnfUser("simple.user") // returns "simple.user"
func DnfUser(user string) string {
	// user can be a string like "rodrigo avila <rodrigo.avila>". But we need to return only what's between < and >.
	// If user is empty, return "Unknown"
	if user == "" {
		return "Unknown"
	}
	if strings.Contains(user, "<") && strings.Contains(user, ">") {
		start := strings.Index(user, "<")
		end := strings.Index(user, ">")
		if start != -1 && end != -1 {
			return user[start+1 : end]
		}
	}
	// If user is not in the format "rodrigo avila <rodrigo.avila>", return the user
	// as is.
	return user
}

// Brand returns the appropriate SVG brand logo filename based on the Linux distribution name.
// It performs a case-insensitive search for known distribution names in the input string
// and returns the corresponding SVG filename.
//
// Known distributions:
//   - AlmaLinux -> brand-almalinux.svg
//   - CentOS -> brand-centos.svg
//   - Fedora -> brand-fedora.svg
//   - Oracle -> brand-oracle.svg
//   - Red Hat -> brand-redhat.svg
//   - Rocky -> brand-rocky.svg
//
// If no known distribution is found in the input string, it returns "brand-linux.svg"
//
// Parameters:
//   - brand: a string containing the Linux distribution name
//
// Returns:
//   - string: the filename of the corresponding brand SVG logo
func Brand(brand string) string {
	if strings.Contains(strings.ToLower(brand), "almalinux") {
		return "brand-almalinux.svg"
	}

	if strings.Contains(strings.ToLower(brand), "centos") {
		return "brand-centos.svg"
	}

	if strings.Contains(strings.ToLower(brand), "fedora") {
		return "brand-fedora.svg"
	}

	if strings.Contains(strings.ToLower(brand), "oracle") {
		return "brand-oracle.svg"
	}

	if strings.Contains(strings.ToLower(brand), "red hat") {
		return "brand-redhat.svg"
	}

	if strings.Contains(strings.ToLower(brand), "rocky") {
		return "brand-rocky.svg"
	}

	return "brand-linux.svg"
}

// HasAction checks if a given action is present in a list of actions or matches a specific action.
// It supports both single word actions and comma-separated lists of action codes.
//
// Parameters:
//   - actions: A string containing either a comma-separated list of action codes (I,D,O,U,E,R,C)
//     or a single action word. Action codes are:
//     I = Install
//     D = Downgrade
//     O = Obsolete
//     U = Upgrade
//     E = Removed
//     R = Reinstall
//     C = Reason Change
//   - action: The specific action to check for, either as a full word (e.g., "Install")
//     or as a single action that should match exactly
//
// Returns:
//   - bool: true if the action is found in the actions list or matches the single action,
//     false otherwise
func HasAction(actions, action string) bool {
	// actions can be a comma-separated list of characters, e.g.
	// "I,D,O,U,E,R,C"; or a word like "Install", "Upgrade", etc. if actions
	// is a word, we need to compare it with the action. if actions is a list,
	// we need to check if the action is in the list.
	// From https://dnf.readthedocs.io/en/latest/command_ref.html#history-command
	actionsList := strings.Split(actions, ",")
	for _, a := range actionsList {
		a = strings.TrimSpace(a)
		switch a {
		case "I":
			if action == "Install" {
				return true
			}
		case "D":
			if action == "Downgrade" {
				return true
			}
		case "O":
			if action == "Obsolete" {
				return true
			}
		case "U":
			if action == "Upgrade" {
				return true
			}
		case "E":
			if action == "Removed" {
				return true
			}
		case "R":
			if action == "Reinstall" {
				return true
			}
		case "C":
			if action == "Reason Change" {
				return true
			}
		default:
			if a == action {
				return true
			}
		}
	}

	return false
}
