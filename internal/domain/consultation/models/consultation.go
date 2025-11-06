package models

import "time"

type Consultation struct {
	ID             int       `json:"id"`
	PacienteID     int       `json:"paciente_id"`
	Motivo         string    `json:"motivo"`
	CuestionarioID int       `json:"cuestionario_id,omitempty"`
	Fecha          time.Time `json:"fecha"`
	Completada     bool      `json:"completada"`
}
