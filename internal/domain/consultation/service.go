//go:generate mockgen -source=service.go -destination=mocks/service.go -package=mocks

package consultation

import (
	"github.com/tonitomc/healthcare-crm-api/internal/domain/consultation/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Service interface {
	GetByID(id int) (*models.Consultation, error)
	GetByPatient(patientID int) ([]models.Consultation, error)
	Create(consultation *models.ConsultationCreateDTO) (int, error)
	Update(id int, consultation *models.ConsultationUpdateDTO) error
	Delete(id int) error
	GetIncomplete(patientID int) ([]models.Consultation, error)
	MarkComplete(id int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(id int) (*models.Consultation, error) {
	if id <= 0 {
		return nil, appErr.Wrap("ConsultationService.GetByID", appErr.ErrInvalidInput, nil)
	}
	return s.repo.GetByID(id)
}

func (s *service) GetByPatient(patientID int) ([]models.Consultation, error) {
	if patientID <= 0 {
		return nil, appErr.Wrap("ConsultationService.GetByPatient", appErr.ErrInvalidInput, nil)
	}
	return s.repo.GetByPatient(patientID)
}

func (s *service) Create(consultation *models.ConsultationCreateDTO) (int, error) {
	if consultation == nil || consultation.PacienteID <= 0 {
		return 0, appErr.Wrap("ConsultationService.Create", appErr.ErrInvalidInput, nil)
	}

	if consultation.Motivo == "" {
		consultation.Motivo = "Consulta General"
	}

	return s.repo.Create(consultation)
}

func (s *service) Update(id int, consultation *models.ConsultationUpdateDTO) error {
	if id <= 0 || consultation == nil {
		return appErr.Wrap("ConsultationService.Update", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Update(id, consultation)
}

func (s *service) Delete(id int) error {
	if id <= 0 {
		return appErr.Wrap("ConsultationService.Delete", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Delete(id)
}

func (s *service) GetIncomplete(patientID int) ([]models.Consultation, error) {
	if patientID <= 0 {
		return nil, appErr.Wrap("ConsultationService.GetIncomplete", appErr.ErrInvalidInput, nil)
	}
	return s.repo.GetIncomplete(patientID)
}

func (s *service) MarkComplete(id int) error {
	if id <= 0 {
		return appErr.Wrap("ConsultationService.MarkComplete", appErr.ErrInvalidInput, nil)
	}
	return s.repo.MarkComplete(id)
}
