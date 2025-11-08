package models

// MedicalRecord representa los antecedentes médicos de un paciente
type MedicalRecord struct {
	ID         int     `json:"id"`
	PacienteID int     `json:"paciente_id"`
	Medicos    *string `json:"medicos,omitempty"`
	Familiares *string `json:"familiares,omitempty"`
	Oculares   *string `json:"oculares,omitempty"`
	Alergicos  *string `json:"alergicos,omitempty"`
	Otros      *string `json:"otros,omitempty"`
}
