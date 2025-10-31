package models

// CreateWorkDayRequest represents POST body for creating new working hours.
type CreateWorkDayRequest struct {
	DayOfWeek int           `json:"day_of_week" validate:"required,min=0,max=6"`
	Ranges    []TimeRangeIn `json:"ranges" validate:"required,dive"`
}

// TimeRangeIn uses string-based time input for JSON convenience.
type TimeRangeIn struct {
	Start string `json:"start"` // e.g. "09:00"
	End   string `json:"end"`   // e.g. "17:00"
}

// CreateSpecialDayRequest represents POST body for creating a special schedule override.
type CreateSpecialDayRequest struct {
	Date   string        `json:"date"` // YYYY-MM-DD
	Ranges []TimeRangeIn `json:"ranges"`
}
