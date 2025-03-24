package models

import "time"

type Statistic struct {
	Name       string     `json:"name"`
	Value      int        `json:"value"`
	Percentage float64    `json:"percentage"`
	UpdatedAt  *time.Time `json:"updated_at"`
}
