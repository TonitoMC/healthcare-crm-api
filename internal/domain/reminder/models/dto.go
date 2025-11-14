package models

type CreateReminderRequest struct {
	Description string `json:"description" validate:"required,min=1"`
	Global      bool   `json:"global"`
}
