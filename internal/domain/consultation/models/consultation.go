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

// ConsultationWithDetails represents a consultation and its related diagnostics and treatments.
type ConsultationWithDetails struct {
	ID             int                        `json:"id"`
	PacienteID     int                        `json:"paciente_id"`
	Motivo         string                     `json:"motivo"`
	CuestionarioID int                        `json:"cuestionario_id,omitempty"`
	Fecha          string                     `json:"fecha"`
	Completada     bool                       `json:"completada"`
	Diagnostics    []DiagnosticWithTreatments `json:"diagnostics"`
}
