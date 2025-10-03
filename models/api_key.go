package models

import (
	"time"
)

// ApiKey represents an API key for authenticating API requests
type ApiKey struct {
	ID          int        `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	KeyHash     string     `json:"-" db:"key_hash"`            // Never expose in JSON
	KeyPrefix   string     `json:"key_prefix" db:"key_prefix"` // For display purposes (e.g., "txlog_ab")
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at" db:"last_used_at"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedBy   *int       `json:"created_by" db:"created_by"`
	CreatorName string     `json:"creator_name,omitempty" db:"creator_name"` // Joined from users table
}

// ApiKeyWithSecret is used only when creating a new API key to return the actual key once
type ApiKeyWithSecret struct {
	ApiKey
	Secret string `json:"secret"` // The actual API key (only shown once)
}
