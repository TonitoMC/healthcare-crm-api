package models

import (
	"time"

	patientModels "github.com/tonitomc/healthcare-crm-api/internal/domain/patient/models"
)

type Appointment struct {
	ID         int       `json:"id"`
	PacienteID *int      `json:"paciente_id,omitempty"`
	Nombre     *string   `json:"nombre,omitempty"` // Para citas sin paciente
	Fecha      time.Time `json:"fecha"`
	Duracion   int64     `json:"duracion"` // segundos
	// Datos enriquecidos del join con paciente
	NombrePaciente   *string    `json:"nombre_paciente,omitempty"`
	TelefonoPaciente *string    `json:"telefono_paciente,omitempty"`
	FechaNacimiento  *time.Time `json:"fecha_nacimiento,omitempty"`
}

type AppointmentCreateDTO struct {
	PacienteID *int      `json:"paciente_id,omitempty"`
	Nombre     *string   `json:"nombre,omitempty"`
	Fecha      time.Time `json:"fecha" validate:"required"`
	Duracion   int64     `json:"duracion" validate:"required"`
}

type AppointmentUpdateDTO struct {
	Fecha    *time.Time `json:"fecha,omitempty"`
	Duracion *int64     `json:"duracion,omitempty"`
}

// AppointmentWithNewPatientDTO - Para crear paciente y cita en una transacción
type AppointmentWithNewPatientDTO struct {
	PatientData     patientModels.PatientCreateDTO `json:"patient_data" validate:"required"`
	AppointmentData struct {
		Fecha    time.Time `json:"fecha" validate:"required"`
		Duracion int64     `json:"duracion" validate:"required"`
	} `json:"appointment_data" validate:"required"`
}

type AvailabilitySlot struct {
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Available bool      `json:"available"`
}
