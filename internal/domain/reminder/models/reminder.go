package models

import "time"

// Reminder represents a full record from the DB.
type Reminder struct {
	ID          int        `json:"id"`
	UserID      *int       `json:"usuario_id,omitempty"`
	Description string     `json:"descripcion"`
	Global      bool       `json:"global"`
	CreatedAt   time.Time  `json:"fecha_creacion"`
	CompletedAt *time.Time `json:"fecha_completado,omitempty"`
}
