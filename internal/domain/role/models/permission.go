package models

type Permission struct {
	ID          int    `json:"id"`
	Name        string `json:"nombre"`
	Description string `json:"descripcion"`
}
