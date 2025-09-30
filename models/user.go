package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID          int       `json:"id" db:"id"`
	Sub         string    `json:"sub" db:"sub"` // OIDC Subject identifier
	Email       string    `json:"email" db:"email"`
	Name        string    `json:"name" db:"name"`
	Picture     string    `json:"picture" db:"picture"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	IsAdmin     bool      `json:"is_admin" db:"is_admin"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	LastLoginAt time.Time `json:"last_login_at" db:"last_login_at"`
}

// UserSession represents a user session
type UserSession struct {
	ID        string    `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}
