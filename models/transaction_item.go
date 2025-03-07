package models

type TransactionItem struct {
	Action   string `json:"action"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Release  string `json:"release"`
	Epoch    string `json:"epoch"`
	Arch     string `json:"arch"`
	Repo     string `json:"repo"`
	FromRepo string `json:"from_repo,omitempty"`
}
