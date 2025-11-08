package models

import "encoding/json"

type Questionnaire struct {
	ID      int             `json:"id"`
	Nombre  string          `json:"nombre"`
	Version string          `json:"version"`
	Activo  bool            `json:"activo"`
	Schema  json.RawMessage `json:"schema"`
}
