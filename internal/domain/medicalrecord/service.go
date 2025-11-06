//go:generate mockgen -source=service.go -destination=mocks/service.go -package=mocks

package medicalrecord

import (
	"github.com/tonitomc/healthcare-crm-api/internal/domain/medicalrecord/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Service interface {
	GetByPatientID(patientID int) (*models.MedicalRecord, error)
	Update(patientID int, dto *models.MedicalRecordUpdateDTO) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// GetByPatientID retrieves the medical record for a patient.
func (s *service) GetByPatientID(patientID int) (*models.MedicalRecord, error) {
	if patientID <= 0 {
		return nil, appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del paciente es inválido.")
	}

	record, err := s.repo.GetByPatientID(patientID)
	if err != nil {
		return nil, appErr.NewDomainError(appErr.ErrNotFound, "No se encontró el expediente médico del paciente.")
	}

	return record, nil
}

// Update merges partial updates from the DTO into the patient's medical record.
func (s *service) Update(patientID int, dto *models.MedicalRecordUpdateDTO) error {
	// 1️⃣ Validate input
	if patientID <= 0 {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del paciente es inválido.")
	}
	if dto == nil {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "Los datos de actualización son requeridos.")
	}

	// 2️⃣ Fetch existing record
	current, err := s.repo.GetByPatientID(patientID)
	if err != nil {
		return appErr.NewDomainError(appErr.ErrNotFound, "No se encontró el expediente médico para actualizar.")
	}

	// 3️⃣ Merge non-nil fields
	if dto.Medicos != nil {
		current.Medicos = dto.Medicos
	}
	if dto.Familiares != nil {
		current.Familiares = dto.Familiares
	}
	if dto.Oculares != nil {
		current.Oculares = dto.Oculares
	}
	if dto.Alergicos != nil {
		current.Alergicos = dto.Alergicos
	}
	if dto.Otros != nil {
		current.Otros = dto.Otros
	}

	// 4️⃣ Save changes
	if err := s.repo.Update(patientID, current); err != nil {
		return appErr.NewDomainError(appErr.ErrInternal, "No se pudo actualizar el expediente médico del paciente.")
	}

	return nil
}
