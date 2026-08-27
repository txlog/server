package tests

import (
	"testing"

	"github.com/lib/pq"
)

// TestVersionSortKeyOrdering checks that version_sort_key orders RPM version
// strings numerically instead of lexicographically (1.18.0 above 1.9.4).
func TestVersionSortKeyOrdering(t *testing.T) {
	db := setupIntegrationTestDB(t)
	defer db.Close()

	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "numeric segments",
			input: []string{"1.9.4", "1.18.0", "1.10.2", "1.2.3"},
			want:  []string{"1.18.0", "1.10.2", "1.9.4", "1.2.3"},
		},
		{
			name:  "release strings",
			input: []string{"9.el9", "10.el9", "1.el9_4"},
			want:  []string{"10.el9", "9.el9", "1.el9_4"},
		},
		{
			name:  "alphanumeric suffix",
			input: []string{"1.2.3rc1", "1.2.10", "1.2.3"},
			want:  []string{"1.2.10", "1.2.3rc1", "1.2.3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(
				`SELECT v FROM unnest($1::text[]) v ORDER BY version_sort_key(v) DESC`,
				pq.Array(tt.input),
			)
			if err != nil {
				t.Fatalf("query failed: %v", err)
			}
			defer rows.Close()

			var got []string
			for rows.Next() {
				var v string
				if err := rows.Scan(&v); err != nil {
					t.Fatalf("scan failed: %v", err)
				}
				got = append(got, v)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}
