package models

type CreateReminderRequest struct {
	Description string `json:"descripcion" validate:"required,min=1"`
	Global      bool   `json:"global"`
}
