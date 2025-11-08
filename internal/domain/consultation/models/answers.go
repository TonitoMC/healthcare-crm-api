package models

import "encoding/json"

// Answers represents the stored responses for a single consultation.
type Answers struct {
	ID             int             `json:"id"`
	ConsultaID     int             `json:"consulta_id"`
	CuestionarioID int             `json:"cuestionario_id"`
	Respuestas     json.RawMessage `json:"respuestas"` // JSONB in DB
}
