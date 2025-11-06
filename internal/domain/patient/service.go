//go:generate mockgen -source=service.go -destination=mocks/service.go -package=mocks

package patient

import (
	"github.com/tonitomc/healthcare-crm-api/internal/domain/patient/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Service interface {
	GetByID(id int) (*models.Patient, error)
	GetAll() ([]models.Patient, error)
	Create(patient *models.PatientCreateDTO) (int, error)
	Update(id int, patient *models.PatientUpdateDTO) error
	Delete(id int) error
	SearchByName(name string) ([]models.PatientSearchResult, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(id int) (*models.Patient, error) {
	if id <= 0 {
		return nil, appErr.Wrap("PatientService.GetByID", appErr.ErrInvalidInput, nil)
	}
	return s.repo.GetByID(id)
}

func (s *service) GetAll() ([]models.Patient, error) {
	return s.repo.GetAll()
}

func (s *service) Create(patient *models.PatientCreateDTO) (int, error) {
	if patient == nil || patient.Nombre == "" || patient.Sexo == "" {
		return 0, appErr.Wrap("PatientService.Create", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Create(patient)
}

func (s *service) Update(id int, patient *models.PatientUpdateDTO) error {
	if id <= 0 || patient == nil {
		return appErr.Wrap("PatientService.Update", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Update(id, patient)
}

func (s *service) Delete(id int) error {
	if id <= 0 {
		return appErr.Wrap("PatientService.Delete", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Delete(id)
}

func (s *service) SearchByName(name string) ([]models.PatientSearchResult, error) {
	if name == "" {
		return nil, appErr.Wrap("PatientService.SearchByName", appErr.ErrInvalidInput, nil)
	}
	return s.repo.SearchByName(name)
}

func (s *service) GetNameByID(patientID int) (string, error) {
	if patientID <= 0 {
		return "", appErr.Wrap("PatientService.GetNameByID", appErr.ErrInvalidInput, nil)
	}

	patient, err := s.repo.GetByID(patientID)
	if err != nil {
		return "", err
	}

	return patient.Nombre, nil
}
