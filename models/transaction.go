package models

import "time"

type Transaction struct {
	TransactionID             string            `json:"transaction_id"`
	MachineID                 string            `json:"machine_id,omitempty"`
	Hostname                  string            `json:"hostname"`
	BeginTime                 *time.Time        `json:"begin_time"`
	EndTime                   *time.Time        `json:"end_time"`
	Actions                   string            `json:"actions"`
	Altered                   string            `json:"altered"`
	User                      string            `json:"user"`
	ReturnCode                string            `json:"return_code"`
	ReleaseVersion            string            `json:"release_version"`
	CommandLine               string            `json:"command_line"`
	Comment                   string            `json:"comment"`
	ScriptletOutput           string            `json:"scriptlet_output"`
	Items                     []TransactionItem `json:"items,omitempty"`
	VulnsFixed                int               `json:"vulns_fixed,omitempty"`
	VulnsIntroduced           int               `json:"vulns_introduced,omitempty"`
	CriticalVulnsFixed        int               `json:"critical_vulns_fixed,omitempty"`
	CriticalVulnsIntroduced   int               `json:"critical_vulns_introduced,omitempty"`
	HighVulnsFixed            int               `json:"high_vulns_fixed,omitempty"`
	HighVulnsIntroduced       int               `json:"high_vulns_introduced,omitempty"`
	MediumVulnsFixed          int               `json:"medium_vulns_fixed,omitempty"`
	MediumVulnsIntroduced     int               `json:"medium_vulns_introduced,omitempty"`
	LowVulnsFixed             int               `json:"low_vulns_fixed,omitempty"`
	LowVulnsIntroduced        int               `json:"low_vulns_introduced,omitempty"`
	IsSecurityPatch           bool              `json:"is_security_patch,omitempty"`
	RiskScoreMitigated        float64           `json:"risk_score_mitigated,omitempty"`
	MaxSeverityFixed          string            `json:"max_severity_fixed,omitempty"`
	VulnerablePackagesUpdated int               `json:"vulnerable_packages_updated,omitempty"`
}
