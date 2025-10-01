package models

import "time"

// Migration represents a database migration
type Migration struct {
	Version     int       `json:"version"`
	Filename    string    `json:"filename"`
	Applied     bool      `json:"applied"`
	AppliedAt   time.Time `json:"applied_at"`
	Description string    `json:"description"`
}

// MigrationStatus represents the overall migration status
type MigrationStatus struct {
	CurrentVersion int         `json:"current_version"`
	IsDirty        bool        `json:"is_dirty"`
	Pending        []Migration `json:"pending"`
	Applied        []Migration `json:"applied"`
	TotalCount     int         `json:"total_count"`
}
