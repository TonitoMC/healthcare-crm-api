package models

type Treatment struct {
	ID               int
	Nombre           string
	DiagnosticoID    int
	ComponenteActivo string
	Presentacion     string
	Dosificacion     string
	Tiempo           string
	Frecuencia       string
}
