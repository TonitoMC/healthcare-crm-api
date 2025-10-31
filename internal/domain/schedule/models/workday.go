package models

// WorkDay defines the recurring schedule for a specific weekday.
// day_of_week: 1–7 (Mon–Sun)
type WorkDay struct {
	ID        int         `json:"id"`
	DayOfWeek int         `json:"day_of_week"`
	Ranges    []TimeRange `json:"ranges"`
	Active    bool        `json:"active"`
}
