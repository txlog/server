package models

import "time"

type Transaction struct {
	TransactionID   string            `json:"transaction_id"`
	MachineID       string            `json:"machine_id,omitempty"`
	Hostname        string            `json:"hostname"`
	BeginTime       *time.Time        `json:"begin_time"`
	EndTime         *time.Time        `json:"end_time"`
	Actions         string            `json:"actions"`
	Altered         string            `json:"altered"`
	User            string            `json:"user"`
	ReturnCode      string            `json:"return_code"`
	ReleaseVersion  string            `json:"release_version"`
	CommandLine     string            `json:"command_line"`
	Comment         string            `json:"comment"`
	ScriptletOutput string            `json:"scriptlet_output"`
	Items           []TransactionItem `json:"items,omitempty"`
}
