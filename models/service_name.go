package models

import "time"

// ServiceName maps a raw value captured by the :svc tag in a topology
// template to a friendly display name.
// Example: MatchValue="acme-system" -> Name="ACME System"
type ServiceName struct {
	ID         int
	MatchValue string
	Name       string
	HasPods    bool
	CreatedAt  time.Time
	// EnvironmentIDs and EnvironmentNames list the environments this service
	// belongs to. A service is only offered on /topology for these environments.
	EnvironmentIDs   []int64
	EnvironmentNames []string
}
