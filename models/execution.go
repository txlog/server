package models

import "time"

type Execution struct {
	ExecutionID           string     `json:"execution_id,omitempty"`
	MachineID             string     `json:"machine_id"`
	Hostname              string     `json:"hostname"`
	ExecutedAt            *time.Time `json:"executed_at"`
	Success               bool       `json:"success"`
	Details               string     `json:"details,omitempty"`
	TransactionsProcessed int        `json:"transactions_processed,omitempty"`
	TransactionsSent      int        `json:"transactions_sent,omitempty"`
	AgentVersion          string     `json:"agent_version,omitempty"`
	OS                    string     `json:"os,omitempty"`
}
