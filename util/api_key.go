package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// GenerateAPIKey generates a cryptographically secure API key with the format: txlog_{random_string}
// Returns the full key and its SHA-256 hash
func GenerateAPIKey() (key string, hash string, prefix string, error error) {
	// Generate 32 random bytes
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to base64url (URL-safe, no padding)
	randomString := base64.RawURLEncoding.EncodeToString(randomBytes)

	// Create the full key with prefix
	fullKey := "txlog_" + randomString

	// Generate SHA-256 hash
	hashBytes := sha256.Sum256([]byte(fullKey))
	keyHash := hex.EncodeToString(hashBytes[:])

	// Get prefix for display (first 13 characters to fit in varchar(16) with "...")
	keyPrefix := fullKey[:min(len(fullKey), 13)] + "..."

	return fullKey, keyHash, keyPrefix, nil
}

// HashAPIKey hashes an API key using SHA-256
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
