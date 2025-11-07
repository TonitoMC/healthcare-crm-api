package models

import "time"

// WorkingHours representa horarios laborales regulares
type WorkingHours struct {
	ID           int        `json:"id"`
	DiaSemana    int        `json:"dia_semana"` // 1=Lunes, 7=Domingo
	HoraApertura *time.Time `json:"hora_apertura,omitempty"`
	HoraCierre   *time.Time `json:"hora_cierre,omitempty"`
	Abierto      bool       `json:"abierto"`
}

// SpecialHours representa horarios especiales para fechas específicas
type SpecialHours struct {
	ID           int        `json:"id"`
	Fecha        time.Time  `json:"fecha"`
	HoraApertura *time.Time `json:"hora_apertura,omitempty"`
	HoraCierre   *time.Time `json:"hora_cierre,omitempty"`
	Abierto      bool       `json:"abierto"`
}

// WorkingHoursUpdateDTO para actualizar horarios laborales
type WorkingHoursUpdateDTO struct {
	HoraApertura *time.Time `json:"hora_apertura,omitempty"`
	HoraCierre   *time.Time `json:"hora_cierre,omitempty"`
	Abierto      *bool      `json:"abierto,omitempty"`
}

// SpecialHoursCreateDTO para crear horario especial
type SpecialHoursCreateDTO struct {
	Fecha        time.Time  `json:"fecha" validate:"required"`
	HoraApertura *time.Time `json:"hora_apertura,omitempty"`
	HoraCierre   *time.Time `json:"hora_cierre,omitempty"`
	Abierto      bool       `json:"abierto"`
}

// AvailabilitySlot representa un slot de tiempo disponible
type AvailabilitySlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
