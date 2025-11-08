package models

import "encoding/json"

type QuestionnaireCreateDTO struct {
	Nombre  string          `json:"nombre"`
	Version string          `json:"version"`
	Activo  bool            `json:"activo"`
	Schema  json.RawMessage `json:"schema"`
}

type QuestionnaireUpdateDTO struct {
	Nombre  string          `json:"nombre"`
	Version string          `json:"version"`
	Activo  bool            `json:"activo"`
	Schema  json.RawMessage `json:"schema"`
}
