package models

// DTO used for creating a reminder.
type ReminderCreateDTO struct {
	UserID      *int   `json:"usuario_id,omitempty"`
	Description string `json:"descripcion"`
	Global      bool   `json:"global"`
}
