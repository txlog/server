package models

import (
	"slices"
	"testing"
)

// TestTopologyServiceEnvironments verifies that a service's environment
// associations round-trip through Create/List and that Update replaces them
// instead of appending.
func TestTopologyServiceEnvironments(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tm := NewTopologyManager(db)

	envA, err := tm.CreateEnvironmentName("tstenva", "Test Env A")
	if err != nil {
		t.Fatalf("CreateEnvironmentName(A): %v", err)
	}
	defer func() { _ = tm.DeleteEnvironmentName(envA.ID) }()

	envB, err := tm.CreateEnvironmentName("tstenvb", "Test Env B")
	if err != nil {
		t.Fatalf("CreateEnvironmentName(B): %v", err)
	}
	defer func() { _ = tm.DeleteEnvironmentName(envB.ID) }()

	svc, err := tm.CreateServiceName("tstsvc", "Test Service", false, []int{envA.ID})
	if err != nil {
		t.Fatalf("CreateServiceName: %v", err)
	}
	defer func() { _ = tm.DeleteServiceName(svc.ID) }()

	assertEnvIDs := func(step string, want ...int) {
		t.Helper()
		svcs, err := tm.ListServiceNames()
		if err != nil {
			t.Fatalf("%s: ListServiceNames: %v", step, err)
		}
		idx := slices.IndexFunc(svcs, func(s ServiceName) bool { return s.ID == svc.ID })
		if idx < 0 {
			t.Fatalf("%s: service %d not returned by ListServiceNames", step, svc.ID)
		}
		got := svcs[idx].EnvironmentIDs
		if len(got) != len(want) {
			t.Fatalf("%s: got environment ids %v, want %v", step, got, want)
		}
		for _, id := range want {
			if !slices.Contains(got, int64(id)) {
				t.Errorf("%s: environment id %d missing from %v", step, id, got)
			}
		}
	}

	assertEnvIDs("after create", envA.ID)

	if err := tm.UpdateServiceName(svc.ID, "tstsvc", "Test Service", false, []int{envB.ID}); err != nil {
		t.Fatalf("UpdateServiceName: %v", err)
	}
	assertEnvIDs("after update", envB.ID)
}

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
			name:        "with any placeholder",
			template:    ":env:any:svc:any:seq",
			wantPattern: `^(.+?).*?(.+?).*(?<!\d)(\d+)$`,
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
			if got.CompiledPattern != tt.wantPattern {
				t.Errorf("CompileTemplate(%q):\n  got  %q\n  want %q", tt.template, got.CompiledPattern, tt.wantPattern)
			}
		})
	}
}
