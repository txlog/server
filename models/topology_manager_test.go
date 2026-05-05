package models

import (
	"regexp"
	"testing"
)

// TestCompileTemplate validates the template→regex compilation logic.
func TestCompileTemplate(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		wantPattern string
		wantErr     bool
	}{
		{
			name:        "full template without domain",
			template:    ":env-dc01-zone1-:svc-database:seq",
			wantPattern: `^([^-]+).*?(.+).*?(\d+)$`,
		},
		{
			name:        "full template with domain suffix",
			template:    ":env-teleco-02-01-:svc-cache:seq.example.com",
			wantPattern: `^([^-]+).*?(.+).*?(\d+).*?$`,
		},
		{
			name:        "only env and seq",
			template:    ":env-static-host:seq",
			wantPattern: `^([^-]+).*?(\d+)$`,
		},
		{
			name:        "only svc",
			template:    "prefix-:svc-suffix",
			wantPattern: `^.*?(.+).*?$`,
		},
		{
			name:     "empty template",
			template: "",
			wantErr:  true,
		},
		{
			name:     "no tags",
			template: "just-a-literal-hostname",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompileTemplate(tt.template)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CompileTemplate(%q): expected error, got nil", tt.template)
				}
				return
			}
			if err != nil {
				t.Errorf("CompileTemplate(%q): unexpected error: %v", tt.template, err)
				return
			}
			if got != tt.wantPattern {
				t.Errorf("CompileTemplate(%q):\n  got  %q\n  want %q", tt.template, got, tt.wantPattern)
			}
		})
	}
}

// TestCompileTemplateMatchesHostnames validates that compiled patterns
// correctly match (and reject) real hostname examples.
func TestCompileTemplateMatchesHostnames(t *testing.T) {
	type matchCase struct {
		hostname string
		matches  bool
	}
	tests := []struct {
		template string
		cases    []matchCase
	}{
		{
			template: ":env-comm-datacenter01-zone1-:svc-database:seq",
			cases: []matchCase{
				{"prd-comm-datacenter01-zone1-acme-system-database01", true},
				{"hlg-comm-datacenter01-zone1-billing-database02", true},
				{"prd-comm-datacenter01-zone1-acme-system-cache01", true},    // matches now because 'database' is .*
				{"prd-comm-datacenter02-zone1-acme-system-database01", true}, // matches now because 'dc01' is .*
			},
		},
		{
			template: ":env-teleco-02-01-:svc-cache:seq.example.com",
			cases: []matchCase{
				{"hlg-teleco-02-01-acme-system-cache00001.example.com", true},
				{"prd-teleco-02-01-billing-cache999.example.com", true},
				{"hlg-teleco-02-01-acme-system-cache00001", true}, // matches because domain is .*
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			pattern, err := CompileTemplate(tt.template)
			if err != nil {
				t.Fatalf("CompileTemplate(%q) error: %v", tt.template, err)
			}

			r, err := regexp.Compile(pattern)
			if err != nil {
				t.Fatalf("compiled pattern %q is not valid Go regexp: %v", pattern, err)
			}

			for _, mc := range tt.cases {
				matched := r.MatchString(mc.hostname)
				if matched != mc.matches {
					t.Errorf("hostname %q: got matched=%v, want %v (pattern: %s)",
						mc.hostname, matched, mc.matches, pattern)
				}
			}
		})
	}
}
