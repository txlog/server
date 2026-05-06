package main
import (
	"fmt"
	"regexp"
)
func main() {
	pattern := `^(?<!\d)(\d+)\-((?:prd)|.+?)\-((?:api)|.+?)$`
	// In Go, (?<!\d) isn't supported, so let's simulate the capture groups Postgres sees
	// Postgres captures all unescaped ( and ).
	// Let's print out what Postgres does.
}
