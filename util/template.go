package util

import (
	"html/template"
	"strconv"
	"strings"

	"github.com/txlog/server/version"
)

func Text2HTML(s string) template.HTML {
	s = strings.ReplaceAll(s, "\n", "<br>")
	s = strings.ReplaceAll(s, " ", "&nbsp;")
	return template.HTML(s)
}

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

func Iterate(start, count int) []int {
	var items []int
	for i := start; i <= count; i++ {
		items = append(items, i)
	}
	return items
}

func Add(a, b int) int {
	return a + b
}
func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func Version() string {
	return version.SemVer
}

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
