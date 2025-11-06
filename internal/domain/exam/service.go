package exam

import (
	"github.com/tonitomc/healthcare-crm-api/internal/domain/exam/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Service interface {
	GetByID(id int) (*models.Exam, error)
	GetByPatient(patientID int) ([]models.Exam, error)
	Create(exam *models.ExamCreateDTO) (int, error)
	Update(id int, upload *models.ExamUploadDTO) error
	Delete(id int) error
	GetPending() ([]models.Exam, error)
	GetCompleted() ([]models.Exam, error)
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

func (s *service) Create(exam *models.ExamCreateDTO) (int, error) {
	if exam.PacienteID <= 0 {
		return 0, appErr.Wrap("ExamService.Create(invalid paciente_id)", appErr.ErrInvalidInput, nil)
	}
	if exam.Tipo == "" {
		return 0, appErr.Wrap("ExamService.Create(tipo required)", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Create(exam)
}

func (s *service) Update(id int, upload *models.ExamUploadDTO) error {
	if id <= 0 {
		return appErr.Wrap("ExamService.Update(invalid id)", appErr.ErrInvalidInput, nil)
	}
	if upload.S3Key == "" {
		return appErr.Wrap("ExamService.Update(s3_key required)", appErr.ErrInvalidInput, nil)
	}
	if upload.MimeType != "application/pdf" {
		return appErr.Wrap("ExamService.Update(only PDF allowed)", appErr.ErrInvalidInput, nil)
	}
	if upload.FileSize <= 0 {
		return appErr.Wrap("ExamService.Update(invalid file_size)", appErr.ErrInvalidInput, nil)
	}
	return s.repo.Update(id, upload)
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

func (s *service) GetCompleted() ([]models.Exam, error) {
	return s.repo.GetCompleted()
}
