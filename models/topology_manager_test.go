package models

import (
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
			wantPattern: `^(.+?)-dc01-zone1-(.+?)-database(?<!\d)(\d+)$`,
		},
		{
			name:        "full template with domain suffix",
			template:    ":env-teleco-02-01-:svc-cache:seq.example.com",
			wantPattern: `^(.+?)-teleco-02-01-(.+?)-cache(?<!\d)(\d+)\.example\.com$`,
		},
		{
			name:        "only env and seq",
			template:    ":env-static-host:seq",
			wantPattern: `^(.+?)-static-host(?<!\d)(\d+)$`,
		},
		{
			name:        "only svc",
			template:    "prefix-:svc-suffix",
			wantPattern: `^prefix-(.+?)-suffix$`,
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

	tm := &TopologyManager{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tm.CompileTemplate(tt.template)
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
