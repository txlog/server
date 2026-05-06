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

// CompilationResult holds the result of compiling a template.
type CompilationResult struct {
	CompiledPattern string
	TagPositions    string // JSON array
	EnvGroupIndex   *int
	SvcGroupIndex   *int
	SeqGroupIndex   *int
}

// knownTags are the supported template tags and their regex capture groups.
var knownTags = map[string]string{
	":env": `(.+?)`,
	":svc": `(.+?)`,
	":seq": `(\d+)`,
	":any": `.*`,
}

// tagOrder defines the order in which tags are replaced so that longer tags
// are matched before shorter ones (avoids partial replacements).
var tagOrder = []string{":env", ":svc", ":seq", ":any"}

// CompileTemplate converts a user-friendly hostname template into a
// PostgreSQL-compatible anchored regex string. It also calculates the
// positions and capture group indices for each tag.
//
// Supported tags:
//   - :env  → (.+?)      environment identifier
//   - :svc  → (.+?)      service identifier
//   - :seq  → (\d+)      pod sequence number
//   - :any  → .*         wildcard (not captured)
//
// Literal parts of the template are preserved as exact matches (anchors).
func (tm *TopologyManager) CompileTemplate(template string) (*CompilationResult, error) {
	if template == "" {
		return nil, errors.New("template must not be empty")
	}

	// Ensure the template contains at least one known tag (excluding :any).
	hasTag := false
	for _, tag := range []string{":env", ":svc", ":seq"} {
		if strings.Contains(template, tag) {
			hasTag = true
			break
		}
	}
	if !hasTag {
		return nil, fmt.Errorf("template must contain at least one tag (:env, :svc or :seq), got: %q", template)
	}

	// Split the template into segments of literal text and tags.
	// We replace each tag with a unique placeholder.
	const placeholder = "\x00"

	// Fetch known values for sticky matching.
	var envVals, svcVals []string
	if tm.db != nil {
		rows, _ := tm.db.Query(`SELECT match_value FROM environment_names ORDER BY length(match_value) DESC`)
		if rows != nil {
			for rows.Next() {
				var v string
				if rows.Scan(&v) == nil {
					envVals = append(envVals, regexp.QuoteMeta(v))
				}
			}
			rows.Close()
		}
		rows, _ = tm.db.Query(`SELECT match_value FROM service_names ORDER BY length(match_value) DESC`)
		if rows != nil {
			for rows.Next() {
				var v string
				if rows.Scan(&v) == nil {
					svcVals = append(svcVals, regexp.QuoteMeta(v))
				}
			}
			rows.Close()
		}
	}

	envRegex := `(.+?)`
	if len(envVals) > 0 {
		envRegex = fmt.Sprintf(`((?:%s)|.+?)`, strings.Join(envVals, "|"))
	}
	svcRegex := `(.+?)`
	if len(svcVals) > 0 {
		svcRegex = fmt.Sprintf(`((?:%s)|.+?)`, strings.Join(svcVals, "|"))
	}

	type tagOccurrence struct {
		tag         string
		placeholder string
		regex       string
		isCapture   bool
	}
	var occurrences []tagOccurrence
	work := template

	// Find all tag occurrences in order of appearance in the template string.
	// We use a regex that matches any of our known tags.
	tagRegex := regexp.MustCompile(`:env|:svc|:seq|:any`)
	matches := tagRegex.FindAllStringIndex(template, -1)

	var tagPositions []string
	captureGroupCount := 0
	var envIdx, svcIdx, seqIdx *int

	tagRegexes := map[string]string{
		":env": envRegex,
		":svc": svcRegex,
		":seq": `(\d+)`,
		":any": `.*?`,
	}

	for i, loc := range matches {
		tag := template[loc[0]:loc[1]]
		tagPositions = append(tagPositions, tag)
		
		ph := fmt.Sprintf("%s%d%s", placeholder, i, placeholder)
		regex := tagRegexes[tag]
		
		// Optimization: if :any is followed by :seq, make it greedy to find the last number.
		// Otherwise, make it reluctant to not over-match the next tag.
		if tag == ":any" {
			if i+1 < len(matches) && template[matches[i+1][0]:matches[i+1][1]] == ":seq" {
				regex = `.*`
			} else {
				regex = `.*?`
			}
		}

		// Optimization: ensure :seq doesn't capture partial numbers if preceded by greedy any
		if tag == ":seq" {
			regex = `(?<!\d)(\d+)`
		}

		isCapture := tag != ":any"
		
		if isCapture {
			captureGroupCount++
			idx := captureGroupCount
			switch tag {
			case ":env":
				envIdx = &idx
			case ":svc":
				svcIdx = &idx
			case ":seq":
				seqIdx = &idx
			}
		}

		occurrences = append(occurrences, tagOccurrence{
			tag:         tag,
			placeholder: ph,
			regex:       regex,
			isCapture:   isCapture,
		})
	}

	// Replace tags with placeholders in the working string.
	// We do this backwards to not mess up indices if we were using string replacement,
	// but since we have the locations from the original template, we can just build it.
	var workBuilder strings.Builder
	last := 0
	for i, loc := range matches {
		workBuilder.WriteString(template[last:loc[0]])
		workBuilder.WriteString(occurrences[i].placeholder)
		last = loc[1]
	}
	workBuilder.WriteString(template[last:])
	work = workBuilder.String()

	// Re-build the final regex.
	phRegex := regexp.MustCompile(`\x00\d+\x00`)
	var result strings.Builder
	result.WriteString("^")
	last = 0
	phMatches := phRegex.FindAllStringIndex(work, -1)

	for i, loc := range phMatches {
		if loc[0] > last {
			literal := work[last:loc[0]]
			result.WriteString(regexp.QuoteMeta(literal))
		}

		result.WriteString(occurrences[i].regex)
		last = loc[1]
	}

	if last < len(work) {
		literal := work[last:]
		result.WriteString(regexp.QuoteMeta(literal))
	}
	result.WriteString("$")

	// Convert tagPositions to JSON
	tagPositionsJSON := "[]"
	if len(tagPositions) > 0 {
		tagPositionsJSON = `["` + strings.Join(tagPositions, `","`) + `"]`
	}

	return &CompilationResult{
		CompiledPattern: result.String(),
		TagPositions:    tagPositionsJSON,
		EnvGroupIndex:   envIdx,
		SvcGroupIndex:   svcIdx,
		SeqGroupIndex:   seqIdx,
	}, nil
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
		SELECT id, template, compiled_pattern, tag_positions,
		       env_group_index, svc_group_index, seq_group_index,
		       display_order, created_at
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
		if err := rows.Scan(&p.ID, &p.Template, &p.CompiledPattern, &p.TagPositions,
			&p.EnvGroupIndex, &p.SvcGroupIndex, &p.SeqGroupIndex,
			&p.DisplayOrder, &p.CreatedAt); err != nil {
			return nil, err
		}
		patterns = append(patterns, p)
	}
	return patterns, rows.Err()
}

// CreatePattern compiles the given template and inserts a new topology_pattern row.
func (tm *TopologyManager) CreatePattern(template string, displayOrder int) (*TopologyPattern, error) {
	res, err := tm.CompileTemplate(template)
	if err != nil {
		return nil, err
	}
	if err := tm.ValidateCompiledPattern(res.CompiledPattern); err != nil {
		return nil, err
	}

	var p TopologyPattern
	p.Template = template
	p.CompiledPattern = res.CompiledPattern
	p.TagPositions = res.TagPositions
	p.EnvGroupIndex = res.EnvGroupIndex
	p.SvcGroupIndex = res.SvcGroupIndex
	p.SeqGroupIndex = res.SeqGroupIndex
	p.DisplayOrder = displayOrder

	err = tm.db.QueryRow(`
		INSERT INTO topology_patterns (template, compiled_pattern, tag_positions, env_group_index, svc_group_index, seq_group_index, display_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`, template, res.CompiledPattern, res.TagPositions, res.EnvGroupIndex, res.SvcGroupIndex, res.SeqGroupIndex, displayOrder).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UpdatePattern recompiles and updates an existing topology_pattern row.
func (tm *TopologyManager) UpdatePattern(id int, template string, displayOrder int) error {
	res, err := tm.CompileTemplate(template)
	if err != nil {
		return err
	}
	if err := tm.ValidateCompiledPattern(res.CompiledPattern); err != nil {
		return err
	}

	_, err = tm.db.Exec(`
		UPDATE topology_patterns
		SET template = $1, compiled_pattern = $2, tag_positions = $3, 
		    env_group_index = $4, svc_group_index = $5, seq_group_index = $6,
		    display_order = $7
		WHERE id = $8
	`, template, res.CompiledPattern, res.TagPositions, res.EnvGroupIndex, res.SvcGroupIndex, res.SeqGroupIndex, displayOrder, id)
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
	_, _ = tm.SyncCompiledPatterns()
	return &e, nil
}

// UpdateEnvironmentName updates an existing environment name mapping.
func (tm *TopologyManager) UpdateEnvironmentName(id int, matchValue, name string) error {
	_, err := tm.db.Exec(`
		UPDATE environment_names
		SET match_value = $1, name = $2
		WHERE id = $3
	`, matchValue, name, id)
	if err == nil {
		_, _ = tm.SyncCompiledPatterns()
	}
	return err
}

// DeleteEnvironmentName removes an environment name mapping by ID.
func (tm *TopologyManager) DeleteEnvironmentName(id int) error {
	_, err := tm.db.Exec(`DELETE FROM environment_names WHERE id = $1`, id)
	if err == nil {
		_, _ = tm.SyncCompiledPatterns()
	}
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
	_, _ = tm.SyncCompiledPatterns()
	return &s, nil
}

// UpdateServiceName updates an existing service name mapping.
func (tm *TopologyManager) UpdateServiceName(id int, matchValue, name string, hasPods bool) error {
	_, err := tm.db.Exec(`
		UPDATE service_names
		SET match_value = $1, name = $2, has_pods = $3
		WHERE id = $4
	`, matchValue, name, hasPods, id)
	if err == nil {
		_, _ = tm.SyncCompiledPatterns()
	}
	return err
}

// DeleteServiceName removes a service name mapping by ID.
func (tm *TopologyManager) DeleteServiceName(id int) error {
	_, err := tm.db.Exec(`DELETE FROM service_names WHERE id = $1`, id)
	if err == nil {
		_, _ = tm.SyncCompiledPatterns()
	}
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

// PreviewEnvironment returns hostnames that match any topology pattern and contain the given matchValue.
func (tm *TopologyManager) PreviewEnvironment(matchValue string) ([]string, error) {
	rows, err := tm.db.Query(`
		SELECT a.hostname
		FROM assets a
		WHERE a.is_active = TRUE
		  AND a.hostname ILIKE '%' || $1 || '%'
		  AND EXISTS (
			  SELECT 1 FROM topology_patterns tp
			  WHERE a.hostname ~ tp.compiled_pattern
		  )
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

// PreviewService returns hostnames that match any topology pattern and contain the given matchValue.
func (tm *TopologyManager) PreviewService(matchValue string) ([]string, error) {
	rows, err := tm.db.Query(`
		SELECT a.hostname
		FROM assets a
		WHERE a.is_active = TRUE
		  AND a.hostname ILIKE '%' || $1 || '%'
		  AND EXISTS (
			  SELECT 1 FROM topology_patterns tp
			  WHERE a.hostname ~ tp.compiled_pattern
		  )
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

// SyncCompiledPatterns checks all stored topology patterns and recompiles their regexes.
// If the newly compiled regex differs from the one stored in the database (e.g., due to an engine update),
// it updates the database automatically. Returns the number of updated patterns.
func (tm *TopologyManager) SyncCompiledPatterns() (int, error) {
	rows, err := tm.db.Query("SELECT id, template, compiled_pattern FROM topology_patterns")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var updates []struct {
		id       int
		template string
	}

	for rows.Next() {
		var id int
		var template, storedCompiled string
		if err := rows.Scan(&id, &template, &storedCompiled); err != nil {
			return 0, err
		}

		res, err := tm.CompileTemplate(template)
		if err != nil {
			continue // Skip invalid templates
		}

		if res.CompiledPattern != storedCompiled {
			updates = append(updates, struct {
				id       int
				template string
			}{id, template})
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	for _, u := range updates {
		res, err := tm.CompileTemplate(u.template)
		if err != nil {
			continue
		}
		_, err = tm.db.Exec(`
			UPDATE topology_patterns 
			SET compiled_pattern = $1, tag_positions = $2, 
			    env_group_index = $3, svc_group_index = $4, seq_group_index = $5 
			WHERE id = $6`,
			res.CompiledPattern, res.TagPositions, res.EnvGroupIndex, res.SvcGroupIndex, res.SeqGroupIndex, u.id)
		if err != nil {
			return len(updates), err
		}
	}

	return len(updates), nil
}
