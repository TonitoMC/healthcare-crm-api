package models

import "time"

// SpecialDay defines working hours for a specific calendar date.
// Overrides the regular weekly schedule.
type SpecialDay struct {
	ID     int         `json:"id"`
	Date   time.Time   `json:"date"` // YYYY-MM-DD
	Ranges []TimeRange `json:"ranges"`
	Active bool        `json:"active"`
}
