package controllers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
)

// ─────────────────────────────────────────────────────────────────────────────
// View structs
// ─────────────────────────────────────────────────────────────────────────────

// TopologyView is the data passed to topology.html.
type TopologyView struct {
	Environments      []models.EnvironmentName
	Services          []models.ServiceName
	SelectedEnv       *models.EnvironmentName
	SelectedSvc       *models.ServiceName
	Pods              []PodView
	OutOfTopology     []PodAsset
	HasPatterns       bool // true if at least one topology_pattern exists
	TotalAssets       int  // sum of all assets across pods + out-of-topology
	TotalNeedsRestart int  // sum of needs_restarting across all pods
	SelectionRequired bool // true if user needs to select env/svc to see data
}

// PodView represents one pod group within the topology view.
type PodView struct {
	PodID        string // "01", "00001", or "Default"
	Assets       []PodAsset
	TotalAssets  int
	NeedsRestart int
	HasCopyFail  bool
}

// PodAsset is a single asset row within a pod.
type PodAsset struct {
	AssetID         int
	Hostname        string
	MachineID       string
	OS              string
	AgentVersion    string
	NeedsRestarting bool
	CopyFail        bool
}

// ─────────────────────────────────────────────────────────────────────────────
// Handler
// ─────────────────────────────────────────────────────────────────────────────

// GetTopologyIndex renders the /topology page.
// Query params: env=<environment name/value>  svc=<service name/value>
func GetTopologyIndex(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tm := models.NewTopologyManager(db)

		// Check if any patterns are configured.
		patterns, err := tm.ListPatterns()
		if err != nil {
			logger.Error("Failed to list topology patterns: " + err.Error())
		}
		hasPatterns := len(patterns) > 0

		// Load dropdowns.
		envs, _ := tm.ListEnvironmentNames()
		svcs, _ := tm.ListServiceNames()

		// Resolve selected env/svc from query params.
		envParam := c.Query("env")
		svcParam := c.Query("svc")

		var selectedEnv *models.EnvironmentName
		for i := range envs {
			if envs[i].Name == envParam || envs[i].MatchValue == envParam {
				selectedEnv = &envs[i]
				break
			}
		}

		var selectedSvc *models.ServiceName
		for i := range svcs {
			if svcs[i].Name == svcParam || svcs[i].MatchValue == svcParam {
				selectedSvc = &svcs[i]
				break
			}
		}

		view := TopologyView{
			Environments: envs,
			Services:     svcs,
			SelectedEnv:  selectedEnv,
			SelectedSvc:  selectedSvc,
			HasPatterns:  hasPatterns,
		}

		if !hasPatterns {
			c.HTML(http.StatusOK, "topology.html", gin.H{
				"Context": c,
				"title":   "Topology - Txlog Server",
				"view":    view,
			})
			return
		}

		// Check if we have both selections.
		if selectedEnv == nil || selectedSvc == nil {
			view.SelectionRequired = true
			c.HTML(http.StatusOK, "topology.html", gin.H{
				"Context": c,
				"title":   "Topology - Txlog Server",
				"view":    view,
			})
			return
		}

		// Build WHERE conditions for env/svc filters.
		envCondition := selectedEnv.MatchValue
		svcCondition := selectedSvc.MatchValue

		// Query assets and resolve topology for each.
		// This query:
		//  1. Joins assets with topology_patterns to find first match
		//  2. Extracts env, svc, seq capture groups
		//  3. Filters by env/svc if provided
		//  4. Returns all active assets with their pod IDs
		assetsQuery := buildTopologyAssetsQuery(envCondition, svcCondition)
		rows, err := db.QueryContext(c.Request.Context(), assetsQuery)
		if err != nil {
			logger.Error("Failed to query topology assets: " + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		// Group assets by pod ID.
		podMap := map[string]*PodView{}
		var podOrder []string
		var outOfTopology []PodAsset

		for rows.Next() {
			var assetID int
			var hostname, machineID string
			var envVal, svcVal, podID sql.NullString
			var agentVersion, os sql.NullString
			var needsRestarting, copyFail sql.NullBool

			if err := rows.Scan(
				&assetID, &hostname, &machineID,
				&envVal, &svcVal, &podID,
				&agentVersion, &os,
				&needsRestarting, &copyFail,
			); err != nil {
				logger.Error("Failed to scan topology row: " + err.Error())
				continue
			}

			asset := PodAsset{
				AssetID:         assetID,
				Hostname:        hostname,
				MachineID:       machineID,
				OS:              os.String,
				AgentVersion:    agentVersion.String,
				NeedsRestarting: needsRestarting.Bool,
				CopyFail:        copyFail.Bool,
			}

			if !podID.Valid || podID.String == "" {
				// No topology pattern matched — Out of Topology
				outOfTopology = append(outOfTopology, asset)
				continue
			}

			pid := podID.String
			if selectedSvc != nil && !selectedSvc.HasPods {
				pid = "Default"
			}

			if _, exists := podMap[pid]; !exists {
				podMap[pid] = &PodView{PodID: pid}
				podOrder = append(podOrder, pid)
			}
			pv := podMap[pid]
			pv.Assets = append(pv.Assets, asset)
			pv.TotalAssets++
			if asset.NeedsRestarting {
				pv.NeedsRestart++
			}
			if asset.CopyFail {
				pv.HasCopyFail = true
			}
		}

		// Build ordered pods slice.
		pods := make([]PodView, 0, len(podOrder))
		totalAssets := 0
		totalNeedsRestart := 0
		for _, pid := range podOrder {
			pv := *podMap[pid]
			totalAssets += pv.TotalAssets
			totalNeedsRestart += pv.NeedsRestart
			pods = append(pods, pv)
		}
		totalAssets += len(outOfTopology)

		view.Pods = pods
		view.OutOfTopology = outOfTopology
		view.TotalAssets = totalAssets
		view.TotalNeedsRestart = totalNeedsRestart

		c.HTML(http.StatusOK, "topology.html", gin.H{
			"Context": c,
			"title":   "Topology - Txlog Server",
			"view":    view,
		})
	}
}

// buildTopologyAssetsQuery returns the SQL to list active assets with their
// topology classification. If envFilter or svcFilter are non-empty, the query
// adds WHERE conditions to filter by the captured group values.
//
// The query tries each topology_pattern in display_order and uses the first match.
// Unmatched assets (no pattern match) are included with NULL pod_id so they
// appear in the "Out of Topology" group on the frontend.
func buildTopologyAssetsQuery(envFilter, svcFilter string) string {
	envCond := ""
	svcCond := ""
	if envFilter != "" {
		envCond = " AND best_env.match_value = '" + sanitizePatternValue(envFilter) + "'"
	}
	if svcFilter != "" {
		svcCond = " AND best_svc.match_value = '" + sanitizePatternValue(svcFilter) + "'"
	}

	return `
		SELECT
			a.asset_id,
			a.hostname,
			a.machine_id,
			resolved.env_val,
			resolved.svc_val,
			resolved.pod_id,
			a.agent_version,
			a.os,
			a.needs_restarting,
			a.copy_fail
		FROM assets a
		LEFT JOIN LATERAL (
			SELECT compiled_pattern,
				   (regexp_match(a.hostname, compiled_pattern))[env_group_index] as raw_env,
				   (regexp_match(a.hostname, compiled_pattern))[svc_group_index] as raw_svc,
				   (regexp_match(a.hostname, compiled_pattern))[seq_group_index] as raw_pod
			FROM topology_patterns
			WHERE a.hostname ~ compiled_pattern
			ORDER BY display_order, id
			LIMIT 1
		) tp ON true
		LEFT JOIN LATERAL (
			SELECT match_value
			FROM environment_names
			WHERE a.hostname ILIKE '%' || match_value || '%'
			ORDER BY length(match_value) DESC
			LIMIT 1
		) best_env ON true
		LEFT JOIN LATERAL (
			SELECT match_value
			FROM service_names
			WHERE a.hostname ILIKE '%' || match_value || '%'
			ORDER BY length(match_value) DESC
			LIMIT 1
		) best_svc ON true
		CROSS JOIN LATERAL (
			SELECT 
				COALESCE(best_env.match_value, tp.raw_env) AS env_val,
				COALESCE(best_svc.match_value, tp.raw_svc) AS svc_val,
				COALESCE(NULLIF(tp.raw_pod, ''), 'Default') AS pod_id
		) resolved
		WHERE a.is_active = TRUE` + envCond + svcCond + `
		ORDER BY resolved.pod_id NULLS LAST, a.hostname
	`
}

// sanitizePatternValue escapes single quotes to prevent SQL injection
// in the literal filter values used by buildTopologyAssetsQuery.
// Since these values come from the environment_names/service_names tables
// (not from user input directly), this is an extra safety measure.
func sanitizePatternValue(v string) string {
	result := ""
	for _, r := range v {
		if r == '\'' {
			result += "''"
		} else {
			result += string(r)
		}
	}
	return result
}
