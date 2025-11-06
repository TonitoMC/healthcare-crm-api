//go:generate mockgen -source=service.go -destination=mocks/service.go -package=mocks

package consultation

import (
	"time"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/consultation/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Service interface {
	GetAll() ([]models.Consultation, error)
	GetByID(id int) (*models.Consultation, error)
	GetByPatient(patientID int) ([]models.Consultation, error)
	Create(dto *models.ConsultationCreateDTO) (int, error)
	Update(id int, dto *models.ConsultationUpdateDTO) error
	Delete(id int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAll() ([]models.Consultation, error) {
	return s.repo.GetAll()
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

func (s *service) Create(dto *models.ConsultationCreateDTO) (int, error) {
	if dto == nil {
		return 0, appErr.Wrap("ConsultationService.Create", appErr.ErrInvalidInput, nil)
	}
	if dto.PacienteID <= 0 {
		return 0, appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del paciente es inválido.")
	}
	if dto.Motivo == "" {
		return 0, appErr.NewDomainError(appErr.ErrInvalidInput, "El motivo de la consulta es requerido.")
	}
	if dto.CuestionarioID <= 0 {
		return 0, appErr.NewDomainError(appErr.ErrInvalidInput, "El cuestionario asociado es inválido.")
	}

	now := time.Now().Truncate(24 * time.Hour)

	consultation := &models.Consultation{
		PacienteID:     dto.PacienteID,
		Motivo:         dto.Motivo,
		CuestionarioID: dto.CuestionarioID,
		Fecha:          now,
		Completada:     false,
	}

	id, err := s.repo.Create(consultation)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *service) Update(id int, dto *models.ConsultationUpdateDTO) error {
	if id <= 0 || dto == nil {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "Datos inválidos para actualización.")
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if dto.Motivo == "" {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El motivo de la consulta es requerido.")
	}

	existing.Motivo = dto.Motivo
	existing.Completada = dto.Completada

	if err := s.repo.Update(existing); err != nil {
		return err
	}

	return nil
}

func (s *service) Delete(id int) error {
	if id <= 0 {
		return appErr.Wrap("ConsultationService.Delete", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Delete(id)
}
