package patient

//go:generate mockgen -source=repository.go -destination=mocks/repository.go -package=mocks

import (
	"database/sql"
	"time"

	"github.com/tonitomc/healthcare-crm-api/internal/database"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/patient/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

type Repository interface {
	GetByID(id int) (*models.Patient, error)
	GetAll() ([]models.Patient, error)
	Create(patient *models.PatientCreateDTO) (int, error)
	Update(id int, patient *models.PatientUpdateDTO) error
	Delete(id int) error
	SearchByName(name string) ([]models.PatientSearchResult, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(id int) (*models.Patient, error) {
	var p models.Patient
	err := r.db.QueryRow(`
		SELECT id, nombre, fecha_nacimiento, telefono, sexo
		FROM pacientes
		WHERE id = $1
	`, id).Scan(&p.ID, &p.Nombre, &p.FechaNacimiento, &p.Telefono, &p.Sexo)
	if err != nil {
		return nil, database.MapSQLError(err, "PatientRepository.GetByID")
	}
	return &p, nil
}

func (r *repository) GetAll() ([]models.Patient, error) {
	rows, err := r.db.Query(`
		SELECT id, nombre, fecha_nacimiento, telefono, sexo
		FROM pacientes
		ORDER BY nombre
	`)
	if err != nil {
		return nil, database.MapSQLError(err, "PatientRepository.GetAll")
	}
	defer rows.Close()

	var patients []models.Patient
	for rows.Next() {
		var p models.Patient
		if err := rows.Scan(&p.ID, &p.Nombre, &p.FechaNacimiento, &p.Telefono, &p.Sexo); err != nil {
			return nil, appErr.Wrap("PatientRepository.GetAll(scan)", appErr.ErrInternal, err)
		}
		patients = append(patients, p)
	}

	if len(patients) == 0 {
		return nil, appErr.Wrap("PatientRepository.GetAll", appErr.ErrNotFound, nil)
	}

	return patients, nil
}

func (r *repository) Create(patient *models.PatientCreateDTO) (int, error) {
	fecha, err := time.Parse("2006-01-02", patient.FechaNacimiento)
	if err != nil {
		return 0, appErr.Wrap("PatientRepository.Create(parse_date)", appErr.ErrInvalidInput, err)
	}

	var id int
	err = r.db.QueryRow(`
		INSERT INTO pacientes (nombre, fecha_nacimiento, telefono, sexo)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, patient.Nombre, fecha, patient.Telefono, patient.Sexo).Scan(&id)
	if err != nil {
		return 0, database.MapSQLError(err, "PatientRepository.Create")
	}
	return id, nil
}

func (r *repository) Update(id int, patient *models.PatientUpdateDTO) error {
	fecha, err := time.Parse("2006-01-02", patient.FechaNacimiento)
	if err != nil {
		return appErr.Wrap("PatientRepository.Update(parse_date)", appErr.ErrInvalidInput, err)
	}

	res, err := r.db.Exec(`
		UPDATE pacientes
		SET nombre = $1, fecha_nacimiento = $2, telefono = $3, sexo = $4
		WHERE id = $5
	`, patient.Nombre, fecha, patient.Telefono, patient.Sexo, id)
	if err != nil {
		return database.MapSQLError(err, "PatientRepository.Update")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("PatientRepository.Update", appErr.ErrNotFound, nil)
	}

	return nil
}

func (r *repository) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM pacientes WHERE id = $1`, id)
	if err != nil {
		return database.MapSQLError(err, "PatientRepository.Delete")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return appErr.Wrap("PatientRepository.Delete", appErr.ErrNotFound, nil)
	}

	return nil
}

func (r *repository) SearchByName(name string) ([]models.PatientSearchResult, error) {
	rows, err := r.db.Query(`
		SELECT id, nombre, telefono, 
		       EXTRACT(YEAR FROM AGE(fecha_nacimiento))::int as edad
		FROM pacientes
		WHERE unaccent(nombre) ILIKE '%' || unaccent($1) || '%'
		ORDER BY nombre
		LIMIT 20
	`, name)
	if err != nil {
		return nil, database.MapSQLError(err, "PatientRepository.SearchByName")
	}
	defer rows.Close()

	var results []models.PatientSearchResult
	for rows.Next() {
		var r models.PatientSearchResult
		if err := rows.Scan(&r.ID, &r.Nombre, &r.Telefono, &r.Edad); err != nil {
			return nil, appErr.Wrap("PatientRepository.SearchByName(scan)", appErr.ErrInternal, err)
		}
		results = append(results, r)
	}

	return results, nil
}
