//go:generate mockgen -source=repository.go -destination=mocks/repository.go -package=mocks

package consultation

import (
	"database/sql"
	"time"

	"github.com/tonitomc/healthcare-crm-api/internal/database"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/consultation/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Repository interface {
	GetByID(id int) (*models.Consultation, error)
	GetByPatient(patientID int) ([]models.Consultation, error)
	Create(consultation *models.ConsultationCreateDTO) (int, error)
	Update(id int, consultation *models.ConsultationUpdateDTO) error
	Delete(id int) error
	GetIncomplete(patientID int) ([]models.Consultation, error)
	MarkComplete(id int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(id int) (*models.Consultation, error) {
	var c models.Consultation
	err := r.db.QueryRow(`
		SELECT id, paciente_id, motivo, cuestionario_id, fecha, completada
		FROM consultas
		WHERE id = $1
	`, id).Scan(&c.ID, &c.PacienteID, &c.Motivo, &c.CuestionarioID, &c.Fecha, &c.Completada)

	if err != nil {
		return nil, database.MapSQLError(err, "ConsultationRepository.GetByID")
	}

	return &c, nil
}

func (r *repository) GetByPatient(patientID int) ([]models.Consultation, error) {
	rows, err := r.db.Query(`
		SELECT id, paciente_id, motivo, cuestionario_id, fecha, completada
		FROM consultas
		WHERE paciente_id = $1
		ORDER BY fecha DESC
	`, patientID)
	if err != nil {
		return nil, database.MapSQLError(err, "ConsultationRepository.GetByPatient")
	}
	defer rows.Close()

	var consultations []models.Consultation
	for rows.Next() {
		var c models.Consultation
		if err := rows.Scan(&c.ID, &c.PacienteID, &c.Motivo, &c.CuestionarioID, &c.Fecha, &c.Completada); err != nil {
			return nil, appErr.Wrap("ConsultationRepository.GetByPatient(scan)", appErr.ErrInternal, err)
		}
		consultations = append(consultations, c)
	}

	return consultations, nil
}

func (r *repository) Create(consultation *models.ConsultationCreateDTO) (int, error) {
	var id int
	err := r.db.QueryRow(`
		INSERT INTO consultas (paciente_id, motivo, cuestionario_id, fecha)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, consultation.PacienteID, consultation.Motivo, consultation.CuestionarioID, time.Now()).Scan(&id)

	if err != nil {
		return 0, database.MapSQLError(err, "ConsultationRepository.Create")
	}
	return id, nil
}

func (r *repository) Update(id int, consultation *models.ConsultationUpdateDTO) error {
	res, err := r.db.Exec(`
		UPDATE consultas
		SET motivo = $1
		WHERE id = $2
	`, consultation.Motivo, id)

	if err != nil {
		return database.MapSQLError(err, "ConsultationRepository.Update")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("ConsultationRepository.Update", appErr.ErrNotFound, nil)
	}

	return nil
}

func (r *repository) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM consultas WHERE id = $1`, id)
	if err != nil {
		return database.MapSQLError(err, "ConsultationRepository.Delete")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("ConsultationRepository.Delete", appErr.ErrNotFound, nil)
	}

	return nil
}

func (r *repository) GetIncomplete(patientID int) ([]models.Consultation, error) {
	rows, err := r.db.Query(`
		SELECT id, paciente_id, motivo, cuestionario_id, fecha, completada
		FROM consultas
		WHERE paciente_id = $1 AND completada = FALSE
		ORDER BY fecha DESC
	`, patientID)
	if err != nil {
		return nil, database.MapSQLError(err, "ConsultationRepository.GetIncomplete")
	}
	defer rows.Close()

	var consultations []models.Consultation
	for rows.Next() {
		var c models.Consultation
		if err := rows.Scan(&c.ID, &c.PacienteID, &c.Motivo, &c.CuestionarioID, &c.Fecha, &c.Completada); err != nil {
			return nil, appErr.Wrap("ConsultationRepository.GetIncomplete(scan)", appErr.ErrInternal, err)
		}
		consultations = append(consultations, c)
	}

	return consultations, nil
}

func (r *repository) MarkComplete(id int) error {
	res, err := r.db.Exec(`
		UPDATE consultas
		SET completada = TRUE
		WHERE id = $1
	`, id)

	if err != nil {
		return database.MapSQLError(err, "ConsultationRepository.MarkComplete")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("ConsultationRepository.MarkComplete", appErr.ErrNotFound, nil)
	}

	return nil
}
