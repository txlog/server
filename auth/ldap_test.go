package auth

import (
	"testing"
)

func TestExtractUIDFromDN(t *testing.T) {
	tests := []struct {
		name     string
		dn       string
		expected string
	}{
		{
			name:     "Standard uid DN",
			dn:       "uid=john.doe,ou=users,dc=example,dc=com",
			expected: "john.doe",
		},
		{
			name:     "Simple uid",
			dn:       "uid=john,dc=example,dc=com",
			expected: "john",
		},
		{
			name:     "UID with spaces",
			dn:       "uid = john.doe , ou=users,dc=example,dc=com",
			expected: "john.doe",
		},
		{
			name:     "CN instead of uid",
			dn:       "cn=John Doe,ou=users,dc=example,dc=com",
			expected: "cn=John Doe,ou=users,dc=example,dc=com", // Returns full DN as fallback
		},
		{
			name:     "Empty DN",
			dn:       "",
			expected: "",
		},
		{
			name:     "Just uid value",
			dn:       "john.doe",
			expected: "john.doe", // No = sign, returns as-is
		},
		{
			name:     "uid with email format",
			dn:       "uid=john.doe@example.com,ou=people,dc=rda,dc=run",
			expected: "john.doe@example.com",
		},
		{
			name:     "uid with special characters",
			dn:       "uid=rodrigo.avila,ou=people,dc=rda,dc=run",
			expected: "rodrigo.avila",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractUIDFromDN(tt.dn)
			if result != tt.expected {
				t.Errorf("extractUIDFromDN(%q) = %q; want %q", tt.dn, result, tt.expected)
			}
		})
	}
}
