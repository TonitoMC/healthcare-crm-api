//go:generate mockgen -source=service.go -destination=mocks/service.go -package=mocks

package medicalrecord

import (
	"github.com/tonitomc/healthcare-crm-api/internal/domain/medicalrecord/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Service interface {
	GetByPatientID(patientID int) (*models.MedicalRecord, error)
	Update(patientID int, record *models.MedicalRecordUpdateDTO) error
	EnsureExists(patientID int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByPatientID(patientID int) (*models.MedicalRecord, error) {
	if patientID <= 0 {
		return nil, appErr.Wrap("MedicalRecordService.GetByPatientID", appErr.ErrInvalidInput, nil)
	}
	return s.repo.GetByPatientID(patientID)
}

func (s *service) Update(patientID int, record *models.MedicalRecordUpdateDTO) error {
	if patientID <= 0 || record == nil {
		return appErr.Wrap("MedicalRecordService.Update", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Update(patientID, record)
}

func (s *service) EnsureExists(patientID int) error {
	if patientID <= 0 {
		return appErr.Wrap("MedicalRecordService.EnsureExists", appErr.ErrInvalidInput, nil)
	}

	_, err := s.repo.GetByPatientID(patientID)
	if err != nil {
		// Si no existe, crear uno vacío
		return s.repo.Create(patientID)
	}
	return nil
}
