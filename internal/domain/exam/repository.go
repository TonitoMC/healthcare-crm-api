//go:generate mockgen -source=repository.go -destination=mocks/repository.go -package=mockspackage exam

package exam

import (
	"database/sql"
	"time"

	"github.com/tonitomc/healthcare-crm-api/internal/database"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/exam/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Repository interface {
	GetByID(id int) (*models.Exam, error)
	GetByPatient(patientID int) ([]models.Exam, error)
	Create(exam *models.ExamCreateDTO) (int, error)
	Update(id int, upload *models.ExamUploadDTO) error
	Delete(id int) error
	GetPending() ([]models.Exam, error)
	GetCompleted() ([]models.Exam, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(id int) (*models.Exam, error) {
	var e models.Exam
	err := r.db.QueryRow(`
		SELECT id, paciente_id, consulta_id, tipo, fecha, s3_key, file_size, mime_type
		FROM examenes
		WHERE id = $1
	`, id).Scan(&e.ID, &e.PacienteID, &e.ConsultaID, &e.Tipo, &e.Fecha, &e.S3Key, &e.FileSize, &e.MimeType)

	if err != nil {
		return nil, database.MapSQLError(err, "ExamRepository.GetByID")
	}

	e.Estado = "PENDIENTE"
	if e.S3Key != nil && *e.S3Key != "" {
		e.Estado = "COMPLETADO"
	}

	return &e, nil
}

func (r *repository) GetByPatient(patientID int) ([]models.Exam, error) {
	rows, err := r.db.Query(`
		SELECT id, paciente_id, consulta_id, tipo, fecha, s3_key, file_size, mime_type
		FROM examenes
		WHERE paciente_id = $1
		ORDER BY fecha DESC NULLS LAST
	`, patientID)
	if err != nil {
		return nil, database.MapSQLError(err, "ExamRepository.GetByPatient")
	}
	defer rows.Close()

	var exams []models.Exam
	for rows.Next() {
		var e models.Exam
		if err := rows.Scan(&e.ID, &e.PacienteID, &e.ConsultaID, &e.Tipo, &e.Fecha, &e.S3Key, &e.FileSize, &e.MimeType); err != nil {
			return nil, appErr.Wrap("ExamRepository.GetByPatient(scan)", appErr.ErrInternal, err)
		}
		e.Estado = "PENDIENTE"
		if e.S3Key != nil && *e.S3Key != "" {
			e.Estado = "COMPLETADO"
		}
		exams = append(exams, e)
	}

	return exams, nil
}

func (r *repository) Create(exam *models.ExamCreateDTO) (int, error) {
	var id int
	err := r.db.QueryRow(`
		INSERT INTO examenes (paciente_id, consulta_id, tipo)
		VALUES ($1, $2, $3)
		RETURNING id
	`, exam.PacienteID, exam.ConsultaID, exam.Tipo).Scan(&id)

	if err != nil {
		return 0, database.MapSQLError(err, "ExamRepository.Create")
	}
	return id, nil
}

func (r *repository) Update(id int, upload *models.ExamUploadDTO) error {
	now := time.Now()
	res, err := r.db.Exec(`
		UPDATE examenes
		SET s3_key = $1, file_size = $2, mime_type = $3, fecha = $4
		WHERE id = $5
	`, upload.S3Key, upload.FileSize, upload.MimeType, now, id)

	if err != nil {
		return database.MapSQLError(err, "ExamRepository.Update")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("ExamRepository.Update", appErr.ErrNotFound, nil)
	}

	return nil
}

func (r *repository) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM examenes WHERE id = $1`, id)
	if err != nil {
		return database.MapSQLError(err, "ExamRepository.Delete")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("ExamRepository.Delete", appErr.ErrNotFound, nil)
	}

	return nil
}

func (r *repository) GetPending() ([]models.Exam, error) {
	rows, err := r.db.Query(`
		SELECT e.id, e.paciente_id, e.consulta_id, e.tipo, e.fecha, e.s3_key, e.file_size, e.mime_type,
		       p.nombre as nombre_paciente
		FROM examenes e
		LEFT JOIN pacientes p ON e.paciente_id = p.id
		WHERE e.s3_key IS NULL OR e.s3_key = ''
		ORDER BY e.fecha DESC NULLS LAST
	`)
	if err != nil {
		return nil, database.MapSQLError(err, "ExamRepository.GetPending")
	}
	defer rows.Close()

	var exams []models.Exam
	for rows.Next() {
		var e models.Exam
		var nombrePaciente *string
		if err := rows.Scan(&e.ID, &e.PacienteID, &e.ConsultaID, &e.Tipo, &e.Fecha, &e.S3Key, &e.FileSize, &e.MimeType, &nombrePaciente); err != nil {
			return nil, appErr.Wrap("ExamRepository.GetPending(scan)", appErr.ErrInternal, err)
		}
		e.Estado = "PENDIENTE"
		if nombrePaciente != nil {
			e.NombrePaciente = *nombrePaciente
		}
		exams = append(exams, e)
	}

	return exams, nil
}

func (r *repository) GetCompleted() ([]models.Exam, error) {
	rows, err := r.db.Query(`
		SELECT e.id, e.paciente_id, e.consulta_id, e.tipo, e.fecha, e.s3_key, e.file_size, e.mime_type,
		       p.nombre as nombre_paciente
		FROM examenes e
		LEFT JOIN pacientes p ON e.paciente_id = p.id
		WHERE e.s3_key IS NOT NULL AND e.s3_key != ''
		ORDER BY e.fecha DESC
	`)
	if err != nil {
		return nil, database.MapSQLError(err, "ExamRepository.GetCompleted")
	}
	defer rows.Close()

	var exams []models.Exam
	for rows.Next() {
		var e models.Exam
		var nombrePaciente *string
		if err := rows.Scan(&e.ID, &e.PacienteID, &e.ConsultaID, &e.Tipo, &e.Fecha, &e.S3Key, &e.FileSize, &e.MimeType, &nombrePaciente); err != nil {
			return nil, appErr.Wrap("ExamRepository.GetCompleted(scan)", appErr.ErrInternal, err)
		}
		e.Estado = "COMPLETADO"
		if nombrePaciente != nil {
			e.NombrePaciente = *nombrePaciente
		}
		exams = append(exams, e)
	}

	return exams, nil
}
