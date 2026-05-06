package models

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// TopologyManager provides CRUD operations for topology configuration:
// patterns (hostname templates), environment names, and service names.
// It also handles compilation of user-friendly templates into PostgreSQL
// regex patterns.
type TopologyManager struct {
	db *sql.DB
}

// NewTopologyManager returns a new TopologyManager backed by the given DB.
func NewTopologyManager(db *sql.DB) *TopologyManager {
	return &TopologyManager{db: db}
}

// ─────────────────────────────────────────────────────────────────────────────
// Template compilation
// ─────────────────────────────────────────────────────────────────────────────

// knownTags are the supported template tags and their regex capture groups.
var knownTags = map[string]string{
	":env": `([^-]+)`,
	":svc": `(.+)`,
	":seq": `(?<!\d)(\d+)`,
}

// tagOrder defines the order in which tags are replaced so that longer tags
// are matched before shorter ones (avoids partial replacements).
var tagOrder = []string{":env", ":svc", ":seq"}

// CompileTemplate converts a user-friendly hostname template into a
// PostgreSQL-compatible anchored regex string.
//
// Supported tags:
//   - :env  → ([^-]+)   environment identifier
//   - :svc  → (.+)      service identifier (greedy)
//   - :seq  → (\d+)     pod sequence number
//
// Literal parts of the template are treated as non-greedy wildcards (.*?).
func CompileTemplate(template string) (string, error) {
	if template == "" {
		return "", errors.New("template must not be empty")
	}

	// Ensure the template contains at least one known tag.
	hasTag := false
	for _, tag := range tagOrder {
		if strings.Contains(template, tag) {
			hasTag = true
			break
		}
	}
	if !hasTag {
		return "", fmt.Errorf("template must contain at least one tag (:env, :svc or :seq), got: %q", template)
	}

	// Split the template into segments of literal text and tags.
	// We replace each tag with a unique placeholder, then replace all
	// text between placeholders with '.*?'.
	const placeholder = "\x00"

	// Build a replacement map: tag → placeholder+index so we can restore order.
	type tagEntry struct {
		placeholder string
		regex       string
	}
	var entries []tagEntry
	work := template

	for _, tag := range tagOrder {
		ph := fmt.Sprintf("%s%d%s", placeholder, len(entries), placeholder)
		if strings.Contains(work, tag) {
			entries = append(entries, tagEntry{placeholder: ph, regex: knownTags[tag]})
			work = strings.ReplaceAll(work, tag, ph)
		}
	}

	// Find placeholders.
	phRegex := regexp.MustCompile(`\x00\d+\x00`)
	var result strings.Builder
	result.WriteString("^")
	last := 0
	matches := phRegex.FindAllStringIndex(work, -1)

	for _, loc := range matches {
		// If there is text before the first placeholder, or between placeholders,
		// replace it with '.*?'.
		if loc[0] > last {
			result.WriteString(".*?")
		}

		// Restore the tag's regex.
		ph := work[loc[0]:loc[1]]
		for _, e := range entries {
			if e.placeholder == ph {
				result.WriteString(e.regex)
				break
			}
		}
		last = loc[1]
	}

	// If there is text after the last placeholder, replace it with '.*?'.
	if last < len(work) {
		result.WriteString(".*?")
	}
	result.WriteString("$")

	return result.String(), nil
}

// ValidateCompiledPattern checks whether a compiled regex is valid in
// PostgreSQL by executing: SELECT ” ~ $1
func (tm *TopologyManager) ValidateCompiledPattern(pattern string) error {
	var ok bool
	err := tm.db.QueryRow(`SELECT '' ~ $1`, pattern).Scan(&ok)
	if err != nil {
		return fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// topology_patterns CRUD
// ─────────────────────────────────────────────────────────────────────────────

// ListPatterns returns all topology patterns ordered by display_order.
func (tm *TopologyManager) ListPatterns() ([]TopologyPattern, error) {
	rows, err := tm.db.Query(`
		SELECT id, template, compiled_pattern, display_order, created_at
		FROM topology_patterns
		ORDER BY display_order, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patterns []TopologyPattern
	for rows.Next() {
		var p TopologyPattern
		if err := rows.Scan(&p.ID, &p.Template, &p.CompiledPattern, &p.DisplayOrder, &p.CreatedAt); err != nil {
			return nil, err
		}
		patterns = append(patterns, p)
	}
	return patterns, rows.Err()
}

// CreatePattern compiles the given template and inserts a new topology_pattern row.
func (tm *TopologyManager) CreatePattern(template string, displayOrder int) (*TopologyPattern, error) {
	compiled, err := CompileTemplate(template)
	if err != nil {
		return nil, err
	}
	if err := tm.ValidateCompiledPattern(compiled); err != nil {
		return nil, err
	}

	var p TopologyPattern
	p.Template = template
	p.CompiledPattern = compiled
	p.DisplayOrder = displayOrder

	err = tm.db.QueryRow(`
		INSERT INTO topology_patterns (template, compiled_pattern, display_order)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, template, compiled, displayOrder).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UpdatePattern recompiles and updates an existing topology_pattern row.
func (tm *TopologyManager) UpdatePattern(id int, template string, displayOrder int) error {
	compiled, err := CompileTemplate(template)
	if err != nil {
		return err
	}
	if err := tm.ValidateCompiledPattern(compiled); err != nil {
		return err
	}

	_, err = tm.db.Exec(`
		UPDATE topology_patterns
		SET template = $1, compiled_pattern = $2, display_order = $3
		WHERE id = $4
	`, template, compiled, displayOrder, id)
	return err
}

// DeletePattern removes a topology_pattern row by ID.
func (tm *TopologyManager) DeletePattern(id int) error {
	_, err := tm.db.Exec(`DELETE FROM topology_patterns WHERE id = $1`, id)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// environment_names CRUD
// ─────────────────────────────────────────────────────────────────────────────

// ListEnvironmentNames returns all environment name mappings ordered by display_order.
func (tm *TopologyManager) ListEnvironmentNames() ([]EnvironmentName, error) {
	rows, err := tm.db.Query(`
		SELECT id, match_value, name, created_at
		FROM environment_names
		ORDER BY name, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envs []EnvironmentName
	for rows.Next() {
		var e EnvironmentName
		if err := rows.Scan(&e.ID, &e.MatchValue, &e.Name, &e.CreatedAt); err != nil {
			return nil, err
		}
		envs = append(envs, e)
	}
	return envs, rows.Err()
}

// CreateEnvironmentName inserts a new environment name mapping.
func (tm *TopologyManager) CreateEnvironmentName(matchValue, name string) (*EnvironmentName, error) {
	var e EnvironmentName
	err := tm.db.QueryRow(`
		INSERT INTO environment_names (match_value, name)
		VALUES ($1, $2)
		RETURNING id, match_value, name, created_at
	`, matchValue, name).Scan(&e.ID, &e.MatchValue, &e.Name, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// UpdateEnvironmentName updates an existing environment name mapping.
func (tm *TopologyManager) UpdateEnvironmentName(id int, matchValue, name string) error {
	_, err := tm.db.Exec(`
		UPDATE environment_names
		SET match_value = $1, name = $2
		WHERE id = $3
	`, matchValue, name, id)
	return err
}

// DeleteEnvironmentName removes an environment name mapping by ID.
func (tm *TopologyManager) DeleteEnvironmentName(id int) error {
	_, err := tm.db.Exec(`DELETE FROM environment_names WHERE id = $1`, id)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// service_names CRUD
// ─────────────────────────────────────────────────────────────────────────────

// ListServiceNames returns all service name mappings ordered by display_order.
func (tm *TopologyManager) ListServiceNames() ([]ServiceName, error) {
	rows, err := tm.db.Query(`
		SELECT id, match_value, name, has_pods, created_at
		FROM service_names
		ORDER BY name, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var svcs []ServiceName
	for rows.Next() {
		var s ServiceName
		if err := rows.Scan(&s.ID, &s.MatchValue, &s.Name, &s.HasPods, &s.CreatedAt); err != nil {
			return nil, err
		}
		svcs = append(svcs, s)
	}
	return svcs, rows.Err()
}

// CreateServiceName inserts a new service name mapping.
func (tm *TopologyManager) CreateServiceName(matchValue, name string, hasPods bool) (*ServiceName, error) {
	var s ServiceName
	err := tm.db.QueryRow(`
		INSERT INTO service_names (match_value, name, has_pods)
		VALUES ($1, $2, $3)
		RETURNING id, match_value, name, has_pods, created_at
	`, matchValue, name, hasPods).Scan(&s.ID, &s.MatchValue, &s.Name, &s.HasPods, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateServiceName updates an existing service name mapping.
func (tm *TopologyManager) UpdateServiceName(id int, matchValue, name string, hasPods bool) error {
	_, err := tm.db.Exec(`
		UPDATE service_names
		SET match_value = $1, name = $2, has_pods = $3
		WHERE id = $4
	`, matchValue, name, hasPods, id)
	return err
}

// DeleteServiceName removes a service name mapping by ID.
func (tm *TopologyManager) DeleteServiceName(id int) error {
	_, err := tm.db.Exec(`DELETE FROM service_names WHERE id = $1`, id)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// Preview helper
// ─────────────────────────────────────────────────────────────────────────────

// PreviewPattern returns hostnames from the assets table that match the
// given compiled regex pattern. Limited to 20 results for the admin preview.
func (tm *TopologyManager) PreviewPattern(compiledPattern string) ([]string, error) {
	rows, err := tm.db.Query(`
		SELECT hostname
		FROM assets
		WHERE is_active = TRUE
		  AND hostname ~ $1
		ORDER BY hostname
		LIMIT 20
	`, compiledPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hostnames []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, err
		}
		hostnames = append(hostnames, h)
	}
	return hostnames, rows.Err()
}

// PreviewEnvironment returns hostnames whose captured env group matches the given matchValue.
func (tm *TopologyManager) PreviewEnvironment(matchValue string) ([]string, error) {
	rows, err := tm.db.Query(`
		SELECT a.hostname
		FROM assets a
		INNER JOIN LATERAL (
			SELECT (regexp_match(a.hostname, compiled_pattern))[1] as raw_env
			FROM topology_patterns
			WHERE a.hostname ~ compiled_pattern
			ORDER BY display_order, id
			LIMIT 1
		) tp ON tp.raw_env ILIKE '%' || $1 || '%'
		WHERE a.is_active = TRUE
		ORDER BY a.hostname
		LIMIT 20
	`, matchValue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hostnames []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, err
		}
		hostnames = append(hostnames, h)
	}
	return hostnames, rows.Err()
}

// PreviewService returns hostnames whose captured svc group matches the given matchValue.
func (tm *TopologyManager) PreviewService(matchValue string) ([]string, error) {
	rows, err := tm.db.Query(`
		SELECT a.hostname
		FROM assets a
		INNER JOIN LATERAL (
			SELECT (regexp_match(a.hostname, compiled_pattern))[2] as raw_svc
			FROM topology_patterns
			WHERE a.hostname ~ compiled_pattern
			ORDER BY display_order, id
			LIMIT 1
		) tp ON tp.raw_svc ILIKE '%' || $1 || '%'
		WHERE a.is_active = TRUE
		ORDER BY a.hostname
		LIMIT 20
	`, matchValue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hostnames []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, err
		}
		hostnames = append(hostnames, h)
	}
	return hostnames, rows.Err()
}

