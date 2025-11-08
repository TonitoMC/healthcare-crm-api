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
	GetByPatientWithDetails(patientID int) ([]models.ConsultationWithDetails, error)
	Create(dto *models.ConsultationCreateDTO) (int, error)
	Update(id int, dto *models.ConsultationUpdateDTO) error
	Delete(id int) error
	MarkComplete(id int) error
	MarkPending(id int) error

	// --- Diagnostics ---
	GetDiagnosticsByConsultation(consultationID int) ([]models.Diagnostic, error)
	GetDiagnosticByID(id int) (*models.Diagnostic, error)
	CreateDiagnostic(dto *models.DiagnosticCreateDTO) (int, error)
	UpdateDiagnostic(id int, dto *models.DiagnosticUpdateDTO) error
	DeleteDiagnostic(id int) error

	// --- Treatments ---
	GetTreatmentsByDiagnostic(diagnosticID int) ([]models.Treatment, error)
	GetTreatmentByID(id int) (*models.Treatment, error)
	CreateTreatment(dto *models.TreatmentCreateDTO) (int, error)
	UpdateTreatment(id int, dto *models.TreatmentUpdateDTO) error
	DeleteTreatment(id int) error
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

func (s *service) GetByPatientWithDetails(patientID int) ([]models.ConsultationWithDetails, error) {
	if patientID <= 0 {
		return nil, appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del paciente es inválido.")
	}

	consultations, err := s.repo.GetByPatient(patientID)
	if err != nil {
		return nil, err
	}

	var result []models.ConsultationWithDetails

	for _, c := range consultations {
		diagnostics, err := s.repo.GetDiagnosticsByConsultation(c.ID)
		if err != nil {
			return nil, err
		}

		var diagDetails []models.DiagnosticWithTreatments
		for _, d := range diagnostics {
			treatments, err := s.repo.GetTreatmentsByDiagnostic(d.ID)
			if err != nil {
				return nil, err
			}

			diagDetails = append(diagDetails, models.DiagnosticWithTreatments{
				ID:            d.ID,
				ConsultaID:    d.ConsultaID,
				Nombre:        d.Nombre,
				Recomendacion: d.Recomendacion,
				Treatments:    treatments,
			})
		}

		result = append(result, models.ConsultationWithDetails{
			ID:             c.ID,
			PacienteID:     c.PacienteID,
			Motivo:         c.Motivo,
			CuestionarioID: c.CuestionarioID,
			Fecha:          c.Fecha.Format("2006-01-02"),
			Completada:     c.Completada,
			Diagnostics:    diagDetails,
		})
	}

	return result, nil
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

func (s *service) MarkComplete(id int) error {
	if id <= 0 {
		return appErr.Wrap("ConsultationService.MarkComplete", appErr.ErrInvalidInput, nil)
	}

	consultation, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	consultation.Completada = true

	return s.repo.Update(consultation)
}

func (s *service) MarkPending(id int) error {
	if id <= 0 {
		return appErr.Wrap("ConsultationService.MarkComplete", appErr.ErrInvalidInput, nil)
	}

	consultation, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	consultation.Completada = false

	return s.repo.Update(consultation)
}

// --- DIAGNOSTICS ---

func (s *service) GetDiagnosticsByConsultation(consultationID int) ([]models.Diagnostic, error) {
	if consultationID <= 0 {
		return nil, appErr.NewDomainError(appErr.ErrInvalidInput, "El ID de la consulta es inválido.")
	}
	return s.repo.GetDiagnosticsByConsultation(consultationID)
}

func (s *service) GetDiagnosticByID(id int) (*models.Diagnostic, error) {
	if id <= 0 {
		return nil, appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del diagnóstico es inválido.")
	}
	return s.repo.GetDiagnosticByID(id)
}

func (s *service) CreateDiagnostic(dto *models.DiagnosticCreateDTO) (int, error) {
	if dto == nil {
		return 0, appErr.NewDomainError(appErr.ErrInvalidInput, "Datos de diagnóstico inválidos.")
	}
	if dto.ConsultaID <= 0 {
		return 0, appErr.NewDomainError(appErr.ErrInvalidInput, "El ID de la consulta asociada es inválido.")
	}
	if dto.Nombre == "" {
		return 0, appErr.NewDomainError(appErr.ErrInvalidInput, "El nombre del diagnóstico es requerido.")
	}

	diagnostic := &models.Diagnostic{
		ConsultaID:    dto.ConsultaID,
		Nombre:        dto.Nombre,
		Recomendacion: dto.Recomendacion,
	}

	id, err := s.repo.CreateDiagnostic(diagnostic)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *service) UpdateDiagnostic(id int, dto *models.DiagnosticUpdateDTO) error {
	if id <= 0 || dto == nil {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "Datos inválidos para la actualización del diagnóstico.")
	}
	if dto.Nombre == "" {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El nombre del diagnóstico es requerido.")
	}

	existing, err := s.repo.GetDiagnosticByID(id)
	if err != nil {
		return err
	}

	existing.Nombre = dto.Nombre
	existing.Recomendacion = dto.Recomendacion

	if err := s.repo.UpdateDiagnostic(existing); err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteDiagnostic(id int) error {
	if id <= 0 {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del diagnóstico es inválido.")
	}
	return s.repo.DeleteDiagnostic(id)
}

// --- TREATMENTS ---

func (s *service) GetTreatmentsByDiagnostic(diagnosticID int) ([]models.Treatment, error) {
	if diagnosticID <= 0 {
		return nil, appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del diagnóstico es inválido.")
	}
	return s.repo.GetTreatmentsByDiagnostic(diagnosticID)
}

func (s *service) GetTreatmentByID(id int) (*models.Treatment, error) {
	if id <= 0 {
		return nil, appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del tratamiento es inválido.")
	}
	return s.repo.GetTreatmentByID(id)
}

func (s *service) CreateTreatment(dto *models.TreatmentCreateDTO) (int, error) {
	if dto == nil {
		return 0, appErr.NewDomainError(appErr.ErrInvalidInput, "Datos de tratamiento inválidos.")
	}
	if dto.DiagnosticoID <= 0 {
		return 0, appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del diagnóstico asociado es inválido.")
	}
	if dto.Nombre == "" {
		return 0, appErr.NewDomainError(appErr.ErrInvalidInput, "El nombre del tratamiento es requerido.")
	}
	if dto.ComponenteActivo == "" {
		return 0, appErr.NewDomainError(appErr.ErrInvalidInput, "El componente activo es requerido.")
	}

	treatment := &models.Treatment{
		Nombre:           dto.Nombre,
		DiagnosticoID:    dto.DiagnosticoID,
		ComponenteActivo: dto.ComponenteActivo,
		Presentacion:     dto.Presentacion,
		Dosificacion:     dto.Dosificacion,
		Tiempo:           dto.Tiempo,
		Frecuencia:       dto.Frecuencia,
	}

	id, err := s.repo.CreateTreatment(treatment)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *service) UpdateTreatment(id int, dto *models.TreatmentUpdateDTO) error {
	if id <= 0 || dto == nil {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "Datos inválidos para la actualización del tratamiento.")
	}
	if dto.Nombre == "" {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El nombre del tratamiento es requerido.")
	}
	if dto.ComponenteActivo == "" {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El componente activo es requerido.")
	}

	existing, err := s.repo.GetTreatmentByID(id)
	if err != nil {
		return err
	}

	existing.Nombre = dto.Nombre
	existing.ComponenteActivo = dto.ComponenteActivo
	existing.Presentacion = dto.Presentacion
	existing.Dosificacion = dto.Dosificacion
	existing.Tiempo = dto.Tiempo
	existing.Frecuencia = dto.Frecuencia

	if err := s.repo.UpdateTreatment(existing); err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteTreatment(id int) error {
	if id <= 0 {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del tratamiento es inválido.")
	}
	return s.repo.DeleteTreatment(id)
}
