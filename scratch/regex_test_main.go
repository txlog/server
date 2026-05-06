package main

import (
	"fmt"
	"regexp"
)

func main() {
	pattern := `^((?:prd-cpaas|prd)|.+?)-((?:api-server|api)|.+?)-(\d+)$`
	re := regexp.MustCompile(pattern)

	tests := []string{
		"prd-cpaas-api-server-01",
		"prd-api-02",
		"unknown-cpaas-api-03", // unknown env
		"prd-cpaas-unknown-server-04", // unknown svc
	}

	for _, t := range tests {
		matches := re.FindStringSubmatch(t)
		if len(matches) > 0 {
			fmt.Printf("String: %s\n", t)
			fmt.Printf("  Env: %s\n", matches[1])
			fmt.Printf("  Svc: %s\n", matches[2])
			fmt.Printf("  Seq: %s\n", matches[3])
		} else {
			fmt.Printf("String: %s NO MATCH\n", t)
		}
	}
}
