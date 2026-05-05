package models

import "time"

// EnvironmentName maps a raw value captured by the :env tag in a topology
// template to a friendly display name.
// Example: MatchValue="prd" -> Name="Production"
type EnvironmentName struct {
	ID         int
	MatchValue string
	Name       string
	CreatedAt  time.Time
}
