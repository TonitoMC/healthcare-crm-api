package exam

import (
	"time"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/exam/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Service interface {
	GetByID(id int) (*models.Exam, error)
	GetByPatient(patientID int) ([]models.Exam, error)
	Create(examDTO *models.ExamCreateDTO) (int, error)
	Update(id int, dto *models.ExamDTO) error
	Delete(id int) error
	GetPending() ([]models.Exam, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(id int) (*models.Exam, error) {
	if id <= 0 {
		return nil, appErr.Wrap("ExamService.GetByID", appErr.ErrInvalidInput, nil)
	}
	return s.repo.GetByID(id)
}

func (s *service) GetByPatient(patientID int) ([]models.Exam, error) {
	if patientID <= 0 {
		return nil, appErr.Wrap("ExamService.GetByPatient", appErr.ErrInvalidInput, nil)
	}

	return s.repo.GetByPatient(patientID)
}

func (s *service) Create(examDTO *models.ExamCreateDTO) (int, error) {
	if examDTO.PacienteID <= 0 {
		return 0, appErr.Wrap("ExamService.Create(invalid paciente_id)", appErr.ErrInvalidInput, nil)
	}
	if examDTO.Tipo == "" {
		return 0, appErr.Wrap("ExamService.Create(tipo required)", appErr.ErrInvalidInput, nil)
	}

	now := time.Now()

	if examDTO.Fecha == nil {
		examDTO.Fecha = &now
	}

	exam := &models.Exam{
		PacienteID: examDTO.PacienteID,
		Tipo:       examDTO.Tipo,
		Fecha:      examDTO.Fecha,
	}

	return s.repo.Create(exam)
}

func (s *service) Update(id int, dto *models.ExamDTO) error {
	if id <= 0 {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "ID inválido para examen.")
	}

	// Fetch existing exam
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// PacienteID (must be positive if provided)
	if dto.PacienteID > 0 {
		existing.PacienteID = dto.PacienteID
	} else if dto.PacienteID < 0 {
		return appErr.NewDomainError(appErr.ErrInvalidInput, "El ID del paciente es inválido.")
	}

	// ConsultaID (optional)
	if dto.ConsultaID != nil {
		existing.ConsultaID = dto.ConsultaID
	}

	// Tipo (if provided)
	if dto.Tipo != "" {
		existing.Tipo = dto.Tipo
	}

	// Fecha (optional, defaults to existing)
	if dto.Fecha != nil {
		existing.Fecha = dto.Fecha
	}

	// Only allow updates if *all three* are present (S3Key, FileSize, MimeType)
	if dto.S3Key != nil || dto.FileSize != nil || dto.MimeType != nil {
		if dto.S3Key == nil || dto.FileSize == nil || dto.MimeType == nil {
			return appErr.NewDomainError(
				appErr.ErrInvalidInput,
				"Campos de carga incompletos: debe incluir s3_key, file_size y mime_type.",
			)
		}
		existing.S3Key = dto.S3Key
		existing.FileSize = dto.FileSize
		existing.MimeType = dto.MimeType
	}

	if err := s.repo.Update(existing); err != nil {
		return err
	}

	return nil
}

func (s *service) Delete(id int) error {
	if id <= 0 {
		return appErr.Wrap("ExamService.Delete", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Delete(id)
}

func (s *service) GetPending() ([]models.Exam, error) {
	return s.repo.GetPending()
}
