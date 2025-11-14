package models

import "encoding/json"

type ConsultationCreateDTO struct {
	PacienteID     int    `json:"paciente_id" validate:"required"`
	Motivo         string `json:"motivo"`
	CuestionarioID int    `json:"cuestionario_id,omitempty"`
}

type ConsultationUpdateDTO struct {
	Motivo     string `json:"motivo"`
	Completada bool   `json:"completada"`
}

type DiagnosticCreateDTO struct {
	ConsultaID    int     `json:"consulta_id"`
	Nombre        string  `json:"nombre"`
	Recomendacion *string `json:"recomendacion"`
}

type DiagnosticUpdateDTO struct {
	Nombre        string  `json:"nombre"`
	Recomendacion *string `json:"recomendacion"`
}

type TreatmentCreateDTO struct {
	Nombre           string `json:"nombre"`
	DiagnosticoID    int    `json:"diagnostico_id"`
	ComponenteActivo string `json:"componente_activo"`
	Presentacion     string `json:"presentacion"`
	Dosificacion     string `json:"dosificacion"`
	Tiempo           string `json:"tiempo"`
	Frecuencia       string `json:"frecuencia"`
}

type TreatmentUpdateDTO struct {
	Nombre           string `json:"nombre"`
	ComponenteActivo string `json:"componente_activo"`
	Presentacion     string `json:"presentacion"`
	Dosificacion     string `json:"dosificacion"`
	Tiempo           string `json:"tiempo"`
	Frecuencia       string `json:"frecuencia"`
}

type AnswersCreateDTO struct {
	CuestionarioID int             `json:"cuestionario_id"`
	Respuestas     json.RawMessage `json:"respuestas"`
}

type AnswersUpdateDTO struct {
	Respuestas json.RawMessage `json:"respuestas"`
}
