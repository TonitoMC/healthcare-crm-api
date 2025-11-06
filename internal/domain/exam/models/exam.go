package models

import "time"

type Exam struct {
	ID             int        `json:"id"`
	PacienteID     int        `json:"paciente_id"`
	ConsultaID     *int       `json:"consulta_id,omitempty"`
	Tipo           string     `json:"tipo"`
	Fecha          *time.Time `json:"fecha,omitempty"`
	S3Key          *string    `json:"s3_key,omitempty"`
	FileSize       *int64     `json:"file_size,omitempty"`
	MimeType       *string    `json:"mime_type,omitempty"`
	Estado         string     `json:"estado"`          // PENDIENTE o COMPLETADO
	NombrePaciente string     `json:"nombre_paciente"` // Nombre del paciente (JOIN)
}

type ExamCreateDTO struct {
	PacienteID int    `json:"paciente_id" validate:"required"`
	ConsultaID *int   `json:"consulta_id,omitempty"`
	Tipo       string `json:"tipo" validate:"required"`
}

type ExamUploadDTO struct {
	S3Key    string `json:"s3_key" validate:"required"`
	FileSize int64  `json:"file_size" validate:"required"`
	MimeType string `json:"mime_type" validate:"required"`
}
