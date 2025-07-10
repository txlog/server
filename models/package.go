package models

import "time"

type PackageProgression struct {
	Week     time.Time `json:"week"`
	Install  int       `json:"install"`
	Upgraded int       `json:"upgraded"`
}
