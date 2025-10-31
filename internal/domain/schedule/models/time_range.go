package models

import "time"

// TimeRange represents a start–end pair within the same day.
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// IsValid checks basic invariant: Start < End
func (tr TimeRange) IsValid() bool {
	return tr.End.After(tr.Start)
}
