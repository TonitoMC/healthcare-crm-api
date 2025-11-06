package models

import "time"

// Patient representa un paciente en el sistema
type Patient struct {
	ID              int       `json:"id"`
	Nombre          string    `json:"nombre"`
	FechaNacimiento time.Time `json:"fecha_nacimiento"`
	Telefono        *string   `json:"telefono,omitempty"`
	Sexo            string    `json:"sexo"`
}

// PatientCreateDTO para crear un paciente
type PatientCreateDTO struct {
	Nombre          string  `json:"nombre" validate:"required"`
	FechaNacimiento string  `json:"fecha_nacimiento" validate:"required"` // format: YYYY-MM-DD
	Telefono        *string `json:"telefono,omitempty"`
	Sexo            string  `json:"sexo" validate:"required,oneof=M F"`
}

// PatientUpdateDTO para actualizar un paciente
type PatientUpdateDTO struct {
	Nombre          string  `json:"nombre" validate:"required"`
	FechaNacimiento string  `json:"fecha_nacimiento" validate:"required"`
	Telefono        *string `json:"telefono,omitempty"`
	Sexo            string  `json:"sexo" validate:"required,oneof=M F"`
}

// PatientSearchResult para resultados de búsqueda
type PatientSearchResult struct {
	ID       int     `json:"id"`
	Nombre   string  `json:"nombre"`
	Telefono *string `json:"telefono,omitempty"`
	Edad     int     `json:"edad"`
}
