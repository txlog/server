package models

// ResolvedTopology holds the result of resolving a hostname against the
// configured topology patterns.
type ResolvedTopology struct {
	EnvironmentValue string // raw value captured by :env, e.g. "prd"
	EnvironmentName  string // friendly name, e.g. "Production"
	ServiceValue     string // raw value captured by :svc, e.g. "acme-system"
	ServiceName      string // friendly name, e.g. "ACME System"
	PodID            string // value captured by :seq, e.g. "01"; "Default" when no match
}

// ResolveHostname resolves a hostname against all configured topology_patterns
// in display_order. Returns nil if no patterns are configured. If a pattern
// matches, it extracts :env, :svc, :seq capture groups and looks up friendly
// names in environment_names and service_names. Assets without a :seq match
// are assigned to pod "Default".
func (tm *TopologyManager) ResolveHostname(hostname string) (*ResolvedTopology, error) {
	// Try each pattern in order until one matches.
	// The query uses regexp_matches to extract the three capture groups.
	// Groups: 1=:env, 2=:svc, 3=:seq (order determined by tag replacement)
	//
	// NOTE: The compiled_pattern always has exactly the groups in the order
	// they appear in the template. Since CompileTemplate replaces tags in the
	// fixed order [:env, :svc, :seq], group 1=env, 2=svc, 3=seq.
	const query = `
		SELECT
			COALESCE((regexp_match($1, tp.compiled_pattern))[1], '') AS env_val,
			COALESCE((regexp_match($1, tp.compiled_pattern))[2], '') AS svc_val,
			COALESCE((regexp_match($1, tp.compiled_pattern))[3], '') AS seq_val
		FROM topology_patterns tp
		WHERE $1 ~ tp.compiled_pattern
		ORDER BY tp.display_order, tp.id
		LIMIT 1
	`

	var envVal, svcVal, seqVal string
	err := tm.db.QueryRow(query, hostname).Scan(&envVal, &svcVal, &seqVal)
	if err != nil {
		// No match — hostname is out of topology.
		return nil, nil //nolint:nilerr
	}

	rt := &ResolvedTopology{
		EnvironmentValue: envVal,
		ServiceValue:     svcVal,
		PodID:            seqVal,
	}
	if rt.PodID == "" {
		rt.PodID = "Default"
	}

	// Look up friendly environment name.
	if envVal != "" {
		var name string
		err := tm.db.QueryRow(
			`SELECT name FROM environment_names WHERE match_value = $1 LIMIT 1`,
			envVal,
		).Scan(&name)
		if err == nil {
			rt.EnvironmentName = name
		} else {
			rt.EnvironmentName = envVal // fallback to raw value
		}
	}

	// Look up friendly service name.
	if svcVal != "" {
		var name string
		err := tm.db.QueryRow(
			`SELECT name FROM service_names WHERE match_value = $1 LIMIT 1`,
			svcVal,
		).Scan(&name)
		if err == nil {
			rt.ServiceName = name
		} else {
			rt.ServiceName = svcVal // fallback to raw value
		}
	}

	return rt, nil
}
