package appointment

import (
	"time"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/appointment/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Service interface {
	GetByID(id int) (*models.Appointment, error)
	GetByDate(date time.Time) ([]models.Appointment, error)
	GetToday() ([]models.Appointment, error)
	GetBetween(start, end time.Time) ([]models.Appointment, error)
	Create(appt *models.AppointmentCreateDTO) (int, error)
	Update(id int, appt *models.AppointmentUpdateDTO) error
	Delete(id int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(id int) (*models.Appointment, error) {
	if id <= 0 {
		return nil, appErr.Wrap("AppointmentService.GetByID", appErr.ErrInvalidInput, nil)
	}
	return s.repo.GetByID(id)
}

func (s *service) GetByDate(date time.Time) ([]models.Appointment, error) {
	return s.repo.GetByDate(date)
}

func (s *service) GetToday() ([]models.Appointment, error) {
	return s.repo.GetToday()
}

func (s *service) GetBetween(start, end time.Time) ([]models.Appointment, error) {
	if start.After(end) {
		return nil, appErr.Wrap("AppointmentService.GetBetween(invalid range)", appErr.ErrInvalidInput, nil)
	}
	return s.repo.GetBetween(start, end)
}

func (s *service) Create(appt *models.AppointmentCreateDTO) (int, error) {
	if appt.PacienteID == nil && appt.Nombre == nil {
		return 0, appErr.Wrap("AppointmentService.Create(must provide paciente_id or nombre)", appErr.ErrInvalidInput, nil)
	}
	if appt.Duracion <= 0 {
		return 0, appErr.Wrap("AppointmentService.Create(duracion must be > 0)", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Create(appt)
}

func (s *service) Update(id int, appt *models.AppointmentUpdateDTO) error {
	if id <= 0 {
		return appErr.Wrap("AppointmentService.Update(invalid id)", appErr.ErrInvalidInput, nil)
	}
	if appt.Duracion != nil && *appt.Duracion <= 0 {
		return appErr.Wrap("AppointmentService.Update(duracion must be > 0)", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Update(id, appt)
}

func (s *service) Delete(id int) error {
	if id <= 0 {
		return appErr.Wrap("AppointmentService.Delete", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Delete(id)
}
