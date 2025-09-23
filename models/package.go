package models

import "time"

type PackageProgression struct {
	Week     time.Time `json:"week"`
	Install  int       `json:"install"`
	Upgraded int       `json:"upgraded"`
}

type PackageListing struct {
	Package       string    `json:"package"`
	Version       string    `json:"version"`
	Release       string    `json:"release"`
	Arch          string    `json:"arch"`
	Repo          string    `json:"repo"`
	TotalVersions int       `json:"total_versions"`
	MachineCount  int       `json:"machine_count"`
	LastSeen      time.Time `json:"last_seen"`
}

type Package struct {
	Name    string `json:"name" uri:"name" binding:"required"`
	Version string `json:"version" uri:"version"`
	Release string `json:"release"`
	Arch    string `json:"arch"`
	Repo    string `json:"repo"`
}
