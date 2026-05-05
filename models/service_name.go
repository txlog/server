package models

import "time"

// ServiceName maps a raw value captured by the :svc tag in a topology
// template to a friendly display name.
// Example: MatchValue="acme-system" -> Name="ACME System"
type ServiceName struct {
	ID           int
	MatchValue   string
	Name         string
	CreatedAt    time.Time
}
