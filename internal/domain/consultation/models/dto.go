package models

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
	ConsultaID    int
	Nombre        string
	Recomendacion *string
}

type DiagnosticUpdateDTO struct {
	Nombre        string
	Recomendacion *string
}

type TreatmentCreateDTO struct {
	Nombre           string
	DiagnosticoID    int
	ComponenteActivo string
	Presentacion     string
	Dosificacion     string
	Tiempo           string
	Frecuencia       string
}

type TreatmentUpdateDTO struct {
	Nombre           string
	ComponenteActivo string
	Presentacion     string
	Dosificacion     string
	Tiempo           string
	Frecuencia       string
}
