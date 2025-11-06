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
