package models

// MedicalRecordUpdateDTO para actualizar antecedentes
type MedicalRecordUpdateDTO struct {
	Medicos    *string `json:"medicos,omitempty"`
	Familiares *string `json:"familiares,omitempty"`
	Oculares   *string `json:"oculares,omitempty"`
	Alergicos  *string `json:"alergicos,omitempty"`
	Otros      *string `json:"otros,omitempty"`
}
