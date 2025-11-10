package models

import "time"

type ExamCreateDTO struct {
	PacienteID int        `json:"paciente_id" validate:"required"`
	Tipo       string     `json:"tipo" validate:"required"`
	Fecha      *time.Time `jsjon:"fecha,omitempty"`
}

type ExamUploadDTO struct {
	FileSize int64 `json:"file_size" validate:"required"`
}

type ExamDTO struct {
	ID             int        `json:"id"`
	PacienteID     int        `json:"paciente_id"`
	ConsultaID     *int       `json:"consulta_id,omitempty"`
	Tipo           string     `json:"tipo"`
	Fecha          *time.Time `json:"fecha,omitempty"`
	S3Key          *string    `json:"s3_key,omitempty"`
	FileSize       *int64     `json:"file_size,omitempty"`
	MimeType       *string    `json:"mime_type,omitempty"`
	Estado         string     `json:"estado"`
	NombrePaciente string     `json:"nombre_paciente,omitempty"`
}
