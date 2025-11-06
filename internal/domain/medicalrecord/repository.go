package medicalrecord
//go:generate mockgen -source=repository.go -destination=mocks/repository.go -package=mocks


import (
	"database/sql"

	"github.com/tonitomc/healthcare-crm-api/internal/database"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/medicalrecord/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Repository interface {
	GetByPatientID(patientID int) (*models.MedicalRecord, error)
	Create(patientID int) error
	Update(patientID int, record *models.MedicalRecordUpdateDTO) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByPatientID(patientID int) (*models.MedicalRecord, error) {
	var rec models.MedicalRecord
	err := r.db.QueryRow(`
		SELECT id, paciente_id, medicos, familiares, oculares, alergicos, otros
		FROM antecedentes
		WHERE paciente_id = $1
	`, patientID).Scan(&rec.ID, &rec.PacienteID, &rec.Medicos, &rec.Familiares, &rec.Oculares, &rec.Alergicos, &rec.Otros)

	if err != nil {
		return nil, database.MapSQLError(err, "MedicalRecordRepository.GetByPatientID")
	}
	return &rec, nil
}

func (r *repository) Create(patientID int) error {
	_, err := r.db.Exec(`
		INSERT INTO antecedentes (paciente_id) VALUES ($1)
	`, patientID)
	if err != nil {
		return database.MapSQLError(err, "MedicalRecordRepository.Create")
	}
	return nil
}

func (r *repository) Update(patientID int, record *models.MedicalRecordUpdateDTO) error {
	res, err := r.db.Exec(`
		UPDATE antecedentes
		SET medicos = $1, familiares = $2, oculares = $3, alergicos = $4, otros = $5
		WHERE paciente_id = $6
	`, record.Medicos, record.Familiares, record.Oculares, record.Alergicos, record.Otros, patientID)

	if err != nil {
		return database.MapSQLError(err, "MedicalRecordRepository.Update")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("MedicalRecordRepository.Update", appErr.ErrNotFound, nil)
	}

	return nil
}
