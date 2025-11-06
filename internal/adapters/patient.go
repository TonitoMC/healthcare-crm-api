package adapters

import "github.com/tonitomc/healthcare-crm-api/internal/domain/patient"

type PatientAdapter struct {
	Service patient.Service
}

func (p *PatientAdapter) GetNameByID(id int) (string, error) {
	patient, err := p.Service.GetByID(id)
	if err != nil {
		return "", err
	}
	return patient.Nombre, nil
}

func (p *PatientAdapter) Exists(id int) (bool, error) {
	_, err := p.Service.GetByID(id)
	if err != nil {
		return false, err
	}
	return true, nil
}
