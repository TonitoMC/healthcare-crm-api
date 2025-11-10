package exam

import (
	"fmt"
	"mime/multipart"
	"time"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/exam/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type FileStorage interface {
	Upload(file multipart.File, key, contentType string) (string, error)
	Delete(key string) error
}

type Service interface {
	GetByID(id int) (*models.ExamDTO, error)
	GetByPatient(patientID int) ([]models.ExamDTO, error)
	Create(examDTO *models.ExamCreateDTO) (int, error)
	Update(id int, dto *models.ExamDTO) error
	Delete(id int) error
	GetPending() ([]models.ExamDTO, error)
	UploadExam(id int, dto *models.ExamUploadDTO, file multipart.File) (*models.ExamDTO, error)
}

type PatientProvider interface {
	GetNameByID(patientID int) (string, error)
}

type service struct {
	repo            Repository
	patientProvider PatientProvider
	storage         FileStorage
}

func NewService(repo Repository, patientProvider PatientProvider, storage FileStorage) Service {
	return &service{repo: repo, patientProvider: patientProvider, storage: storage}
}

func (s *service) GetByID(id int) (*models.ExamDTO, error) {
	if id <= 0 {
		return nil, appErr.Wrap("ExamService.GetByID", appErr.ErrInvalidInput, nil)
	}

	exam, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.enrich(*exam)
}

func (s *service) GetByPatient(patientID int) ([]models.ExamDTO, error) {
	if patientID <= 0 {
		return nil, appErr.Wrap("ExamService.GetByPatient", appErr.ErrInvalidInput, nil)
	}

	exams, err := s.repo.GetByPatient(patientID)
	if err != nil {
		return nil, err
	}

	enriched := make([]models.ExamDTO, 0, len(exams))
	for _, exam := range exams {
		dto, err := s.enrich(exam)
		if err != nil {
			return nil, err
		}
		enriched = append(enriched, *dto)
	}

	return enriched, nil
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

func (s *service) GetPending() ([]models.ExamDTO, error) {
	pendingExams, err := s.repo.GetPending()
	if err != nil {
		return nil, err
	}

	enriched := make([]models.ExamDTO, 0, len(pendingExams))
	for _, exam := range pendingExams {
		dto, err := s.enrich(exam)
		if err != nil {
			return nil, err
		}
		enriched = append(enriched, *dto)
	}

	return enriched, nil
}

func (s *service) enrich(e models.Exam) (*models.ExamDTO, error) {
	dto := &models.ExamDTO{
		ID:         e.ID,
		PacienteID: e.PacienteID,
		ConsultaID: e.ConsultaID,
		Tipo:       e.Tipo,
		Fecha:      e.Fecha,
		S3Key:      e.S3Key,
		FileSize:   e.FileSize,
		MimeType:   e.MimeType,
	}

	if s.patientProvider != nil {
		if name, err := s.patientProvider.GetNameByID(e.PacienteID); err == nil {
			dto.NombrePaciente = name
		}
	}

	return dto, nil
}

func (s *service) UploadExam(id int, dto *models.ExamUploadDTO, file multipart.File) (*models.ExamDTO, error) {
	if id <= 0 {
		return nil, appErr.Wrap("ExamService.UploadExam", appErr.ErrInvalidInput, nil)
	}
	if dto == nil {
		return nil, appErr.Wrap("ExamService.UploadExam", appErr.ErrInvalidInput, nil)
	}

	exam, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err // bubble up repo error
	}

	if s.storage == nil {
		return nil, appErr.NewDomainError(appErr.ErrInternal, "El almacenamiento no está configurado correctamente.")
	}

	// Always enforce PDF-only uploads
	mimeType := "application/pdf"

	// Generate deterministic key
	filename := fmt.Sprintf("exams/%d_%d.pdf", exam.ID, time.Now().UnixNano())

	// Upload file (PDF only)
	if _, err := s.storage.Upload(file, filename, mimeType); err != nil {
		return nil, appErr.Wrap("ExamService.UploadExam", appErr.ErrInternal, err)
	}

	// Update exam metadata
	exam.S3Key = &filename
	exam.FileSize = &dto.FileSize
	exam.MimeType = &mimeType

	if err := s.repo.Update(exam); err != nil {
		return nil, appErr.Wrap("ExamService.UploadExam", appErr.ErrInternal, err)
	}

	return s.enrich(*exam)
}
