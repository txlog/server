package models

import "time"

// TopologyPattern represents a hostname template used to classify assets
// into environments, services, and pods.
// Users define templates using :env, :svc, :seq tags (e.g. ":env-dc01-:svc-db:seq").
// The server compiles these to PostgreSQL-compatible regex patterns.
type TopologyPattern struct {
	ID              int
	Template        string
	CompiledPattern string
	TagPositions    string // JSON array: [":env", ":any", ":svc", ":any", ":seq"]
	EnvGroupIndex   *int   // nullable: capture group index for :env
	SvcGroupIndex   *int   // nullable: capture group index for :svc
	SeqGroupIndex   *int   // nullable: capture group index for :seq
	DisplayOrder    int
	CreatedAt       time.Time
}
