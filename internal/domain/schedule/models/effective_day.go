package models

import "time"

// EffectiveDay represents the resolved working hours for a given date.
// This is not stored in DB; it's computed dynamically.
type EffectiveDay struct {
	Date       time.Time   `json:"date"`
	Ranges     []TimeRange `json:"ranges"`
	IsOverride bool        `json:"is_override"` // true if came from SpecialDay
	Active     bool        `json:"active"`      // false if closed
}
