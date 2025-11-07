package adapters

import (
	"github.com/tonitomc/healthcare-crm-api/internal/domain/patient"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/patient/models"
)

type PatientAdapter struct {
	Service patient.Service
}

func NewPatientAdapter(service patient.Service) *PatientAdapter {
	return &PatientAdapter{Service: service}
}

func (p *PatientAdapter) GetNameByID(id int) (string, error) {
	patient, err := p.Service.GetByID(id)
	if err != nil {
		return "", err
	}
	return patient.Nombre, nil
}

func (p *PatientAdapter) GetByID(id int) (*models.Patient, error) {
	return p.Service.GetByID(id)
}

func (p *PatientAdapter) Exists(id int) (bool, error) {
	_, err := p.Service.GetByID(id)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (p *PatientAdapter) Create(dto *models.PatientCreateDTO) (int, error) {
	return p.Service.Create(dto)
}
